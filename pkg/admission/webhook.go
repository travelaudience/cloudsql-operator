/*
Copyright 2019 The cloudsql-operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package admission

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"

	"github.com/travelaudience/cloudsql-operator/pkg/apis/cloudsql/v1alpha1"
	cloudsqlclient "github.com/travelaudience/cloudsql-operator/pkg/client/clientset/versioned"
	"github.com/travelaudience/cloudsql-operator/pkg/configuration"
	"github.com/travelaudience/cloudsql-operator/pkg/crds"
)

const (
	// healthzPath is the path where the "/healthz" endpoint is served.
	healthzPath = "/healthz"
)

var (
	// admissionPath is the path where the handler for admission requests is served.
	admissionPath = "/admissionrequests"
	// postgresqlInstanceGvk is the GroupVersionKind that corresponds to PostgresqlInstance resources.
	postgresqlInstanceGvk = &schema.GroupVersionKind{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    crds.PostgresqlInstanceKind,
	}
	// postgresqlInstanceGvr is the GroupVersionResource that corresponds to PostgresqlInstance resources.
	postgresqlInstanceGvr = metav1.GroupVersionResource{
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Resource: crds.PostgresqlInstancePlural,
	}
	// patchType is the type of patch sent in admission responses.
	patchType = admissionv1beta1.PatchTypeJSONPatch
)

// Webhook represents an instance of the admission webhook.
type Webhook struct {
	// bindAddress is the bind address to use for the server.
	bindAddress string
	// cloudsqlClient is a client to the cloudsql-operator API.
	cloudsqlClient cloudsqlclient.Interface
	// codecs is the codec factory to use to serialize/deserialize resources.
	codecs serializer.CodecFactory
	// kubeClient is the Kubernetes client to use.
	kubeClient kubernetes.Interface
	// namespace is the namespace where cloudsql-operator is deployed.
	namespace string
	// tlsCertificate is the TLS certificate (and private key) used to register and serve the webhook.
	tlsCertificate tls.Certificate
}

// NewWebhook creates a new instance of the admission webhook.
func NewWebhook(kubeClient kubernetes.Interface, cloudsqlClient cloudsqlclient.Interface, config configuration.Configuration) (*Webhook, error) {
	// Create a new scheme and register the PostgresqlInstance type so we can serialize/deserialize it.
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.PostgresqlInstance{})
	return &Webhook{
		bindAddress:    config.Admission.BindAddress,
		cloudsqlClient: cloudsqlClient,
		codecs:         serializer.NewCodecFactory(scheme),
		kubeClient:     kubeClient,
		namespace:      config.Cluster.Namespace,
	}, nil
}

// Run starts the HTTP server that backs the admission webhook.
func (w *Webhook) Run(stopCh chan struct{}) error {
	// Create an HTTP server and register handler functions to back the admission webhook.
	mux := http.NewServeMux()
	mux.HandleFunc(admissionPath, w.handleAdmission)
	mux.HandleFunc(healthzPath, handleHealthz)
	srv := http.Server{
		Addr:    w.bindAddress,
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{w.tlsCertificate},
		},
	}

	// Shutdown the server when stopCh is closed.
	go func() {
		<-stopCh
		ctx, fn := context.WithTimeout(context.Background(), 5*time.Second)
		defer fn()
		if err := srv.Shutdown(ctx); err != nil {
			log.Errorf("failed to shutdown the admission webhook: %v", err)
		} else {
			log.Debug("the admission webhook has been shutdown")
		}
	}()

	// Start listening and serving requests.
	log.Debug("starting the admission webhook")
	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// handleAdmission handles the HTTP portion of admission.
func (w *Webhook) handleAdmission(res http.ResponseWriter, req *http.Request) {
	// Read the request's body.
	var body []byte
	if req.Body != nil {
		if data, err := ioutil.ReadAll(req.Body); err == nil {
			body = data
		}
	}

	// Fail if the request's method is not "POST".
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Fail if the request's content type is not "application/json".
	if req.Header.Get("Content-Type") != "application/json" {
		res.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	// aReq is the AdmissionReview that was sent to the admission webhook.
	aReq := admissionv1beta1.AdmissionReview{}
	// rRes is the AdmissionReview that will be returned.
	aRes := admissionv1beta1.AdmissionReview{}

	// Deserialize the requested AdmissionReview and, if successful, pass it to the provided admission function.
	deserializer := w.codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &aReq); err != nil {
		aRes.Response = admissionResponseFromError(err)
	} else {
		aRes.Response = w.validateAndMutateResource(aReq)
	}
	// Set the request's UID in the response object.
	aRes.Response.UID = aReq.Request.UID

	// Serialize the response AdmissionReview.
	resp, err := json.Marshal(aRes)
	if err != nil {
		log.Errorf("failed to write admissionreview: %v", err)
		return
	}
	if _, err := res.Write(resp); err != nil {
		log.Errorf("failed to write admissionreview: %v", err)
		return
	}
}

// readObject reads the object pointed by the provided coordinates from the Kubernetes API.
func (w *Webhook) readObject(kind *schema.GroupVersionKind, namespace, name string) (runtime.Object, error) {
	switch kind {
	case postgresqlInstanceGvk:
		return w.cloudsqlClient.CloudsqlV1alpha1().PostgresqlInstances().Get(name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("unsupported gvk: %s", kind.String())
	}
}

// validateAndMutateResource takes an AdmissionReview object and performs validation and mutation on the associated resource.
func (w *Webhook) validateAndMutateResource(rev admissionv1beta1.AdmissionReview) *admissionv1beta1.AdmissionResponse {
	var (
		// currentObj will contain the resource in its current form.
		// It MUST NOT be modified, as it is used as the basis for the patch to apply as a result of the current request.
		currentObj runtime.Object
		// currentGVK will contain the GVK (Group/Version/Kind) of the current resource.
		// It is used to identify the kind of resource (PostgresqlInstance/...) we are dealing with in the current request.
		currentGVK *schema.GroupVersionKind
		// mutatedObj will contain a clone of currentObj.
		// It will be modified as required in order to explicitly set the values of all annotations.
		mutatedObj runtime.Object
		// previousObj will contain the previous version of the current resource.
		// It will only be set and considered in case the current operation's type is "UPDATE".
		previousObj runtime.Object
		// err will contain any error we may encounter during the validation/mutation process.
		err error
	)

	// Set currentGVK based on the provided GVR (group/version/resource).
	switch rev.Request.Resource {
	case postgresqlInstanceGvr:
		// We're dealing with a PostgresqlInstance resource.
		currentGVK = postgresqlInstanceGvk
	default:
		// We're dealing with an unsupported resource, so we must fail.
		return admissionResponseFromError(fmt.Errorf("failed to validate resource with unsupported gvr %s", rev.Request.Resource.String()))
	}

	// Populate currentObj and previousObj based on the type of the current operation.
	switch rev.Request.Operation {
	case admissionv1beta1.Create:
		// Deserialize the current object.
		currentObj, _, err = w.codecs.UniversalDeserializer().Decode(rev.Request.Object.Raw, currentGVK, nil)
		if err != nil {
			return admissionResponseFromError(fmt.Errorf("failed to deserialize the current object: %v", err))
		}
	case admissionv1beta1.Update:
		// Deserialize the current object.
		currentObj, _, err = w.codecs.UniversalDeserializer().Decode(rev.Request.Object.Raw, currentGVK, nil)
		if err != nil {
			return admissionResponseFromError(fmt.Errorf("failed to deserialize the current object: %v", err))
		}
		// Deserialize the previous object.
		previousObj, _, err = w.codecs.UniversalDeserializer().Decode(rev.Request.OldObject.Raw, currentGVK, nil)
		if err != nil {
			return admissionResponseFromError(fmt.Errorf("failed to deserialize the previous object: %v", err))
		}
	case admissionv1beta1.Delete:
		// DELETE requests do not populate "rev.Request.Object" or "rev.Request.OldObject".
		// Hence we try to read it directly from the Kubernetes API using the data at hand.
		previousObj, err = w.readObject(currentGVK, rev.Request.Namespace, rev.Request.Name)
		if err != nil {
			return admissionResponseFromError(fmt.Errorf("failed to read deleted object: %v", err))
		}
	default:
		// We don't support acting upon the current operation, so we should reject it.
		// This should never happen in practice, as it probably means that the MutatingWebhookConfiguration resource contains incorrect data.
		return admissionResponseFromError(fmt.Errorf("unsupported operation %q", rev.Request.Operation))
	}

	// Perform validation on the current resource according to its type.
	switch currentGVK {
	case postgresqlInstanceGvk:
		var (
			currentPostgresqlInstance, previousPostgresqlInstance *v1alpha1.PostgresqlInstance
		)
		// If currentObj is not nil, cast it to PostgresqlInstance.
		if currentObj != nil {
			currentPostgresqlInstance = currentObj.(*v1alpha1.PostgresqlInstance)
		}
		// If previousObj is not nil, cast it to PostgresqlInstance.
		if previousObj != nil {
			previousPostgresqlInstance = previousObj.(*v1alpha1.PostgresqlInstance)
		}
		mutatedObj, err = w.validateAndMutatePostgresqlInstance(currentPostgresqlInstance, previousPostgresqlInstance)
	default:
		return admissionResponseFromError(fmt.Errorf("failed to validate resource of unsupported type %v", reflect.TypeOf(currentObj)))
	}

	// If an error was returned as a result of validation, we must fail.
	if err != nil {
		return admissionResponseFromError(err)
	}
	// If currentObj is nil (i.e. the current request is a DELETE request), there's nothing to mutate.
	if currentObj == nil {
		return admissionResponseOK()
	}
	// In all other cases, we admit the request and provide a (possibly empty) patch to be applied to the resource.
	return admissionResponseWithPatch(currentObj, mutatedObj)
}

// admissionResponseFromError creates an admission response based on the specified error.
func admissionResponseFromError(err error) *admissionv1beta1.AdmissionResponse {
	return &admissionv1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// admissionResponseOK creates an admission response that allows the current operation.
func admissionResponseOK() *admissionv1beta1.AdmissionResponse {
	return &admissionv1beta1.AdmissionResponse{
		Allowed: true,
	}
}

// admissionResponseWithPatch created an admission response that allows the current operation and specifies a patch to be applied to the resource.
func admissionResponseWithPatch(currentObj, mutatedObj runtime.Object) *admissionv1beta1.AdmissionResponse {
	// Create a patch containing the changes to apply to the resource.
	patch, err := CreateRFC6902Patch(currentObj, mutatedObj)
	if err != nil {
		return admissionResponseFromError(fmt.Errorf("failed to create patch: %v", err))
	}
	// Return an admission response that admits the resource and contains the patch to be applied.
	return &admissionv1beta1.AdmissionResponse{
		Allowed:   true,
		Patch:     patch,
		PatchType: &patchType,
	}
}
