/*
Copyright 2019 The cloudsql-postgres-operator Authors.

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
	"crypto/tls"
	"encoding/pem"
	"reflect"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/cert"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/configuration"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/crds"
)

const (
	// cloudsqlPostgresOperatorServiceName is the name of the service used to back the admission webhook.
	cloudsqlPostgresOperatorServiceName = "cloudsql-postgres-operator"
	// mutatingWebhookConfigurationResourceName is the name to use when creating the MutatingWebhookConfiguration resource.
	mutatingWebhookConfigurationResourceName = "cloudsql-postgres-operator"
	// podKind is the kind that corresponds to Pod resources.
	podKind = "Pod"
	// podPlural is the plural name that corresponds to Pod resources.
	podPlural = "pods"
	// podWebhookName is the name of the admission webhook that deals with Pod resources.
	podWebhookName = "pod.cloudsql.travelaudience.com"
	// postgresqlInstanceWebhookName is the name of the admission webhook that deals with PostgresqlInstance resources.
	postgresqlInstanceWebhookName = "postgresqlinstance.cloudsql.travelaudience.com"
)

var (
	// podFailurePolicy is the failure policy to use for the admission webhook that deals with Pod resources.
	podFailurePolicy = admissionregistrationv1beta1.Ignore
	// postgresInstanceFailurePolicy is the failure policy to use for the admission webhook that deals with PostgresqlInstance resources.
	postgresInstanceFailurePolicy = admissionregistrationv1beta1.Fail
)

// Register registers the admission webhook by making sure a MutatingWebhookConfiguration resource with the desired configuration exists.
func (w *Webhook) Register(kubeClient kubernetes.Interface, cfg configuration.Configuration) error {
	// Make sure the secret containing the required TLS material exists, creating it if necessary.
	sec, err := w.ensureTLSSecret()
	if err != nil {
		return err
	}
	// Parse the PEM-encoded TLS material contained in the secret.
	crt, err := tls.X509KeyPair(sec.Data[v1.TLSCertKey], sec.Data[v1.TLSPrivateKeyKey])
	if err != nil {
		return err
	}
	// Store the TLS material so it can be used for serving the admission webhook later on.
	w.tlsCertificate = crt

	// Create the admission webhook configuration object containing the desired configuration.
	desiredCfg := w.buildMutatingWebhookConfigurationObject()

	// Attempt to register the admission webhook.
	_, err = kubeClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(desiredCfg)
	if err == nil {
		// Registration was successful.
		return nil
	}
	if !errors.IsAlreadyExists(err) {
		// The admission webhook is not registered yet but we've got an unexpected error while registering it.
		return err
	}

	// At this point the admission webhook is already registered but the spec of the corresponding MutatingWebhookConfiguration resource may differ.

	// Read the latest version of the MutatingWebhookConfiguration resource and check whether the specs match.
	currentCfg, err := kubeClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(mutatingWebhookConfigurationResourceName, metav1.GetOptions{})
	if err != nil {
		// We've failed to fetch the latest version of the MutatingWebhookConfiguration resource.
		return err
	}
	if reflect.DeepEqual(currentCfg.Webhooks, desiredCfg.Webhooks) {
		// The specs match, so there's nothing left to do.
		return nil
	}

	// Attempt to update the resource by setting the resulting resource's ".spec" field according to the desired value.
	currentCfg.Webhooks = desiredCfg.Webhooks
	if _, err := kubeClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(currentCfg); err != nil {
		return err
	}
	return nil
}

// buildMutatingWebhookConfigurationObject builds the MutatingWebhookConfiguration object used to register the admission webhook.
func (w *Webhook) buildMutatingWebhookConfigurationObject() *admissionregistrationv1beta1.MutatingWebhookConfiguration {
	// PEM-encode the TLS certificate so we can use it as the value of ".webhooks[*].clientConfig.caBundle".
	caBundle := pem.EncodeToMemory(&pem.Block{
		Type:  cert.CertificateBlockType,
		Bytes: w.tlsCertificate.Certificate[0],
	})
	// Return the MutatingWebhookConfiguration object.
	return &admissionregistrationv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.LabelAppKey: constants.ApplicationName,
			},
			Name: mutatingWebhookConfigurationResourceName,
		},
		Webhooks: []admissionregistrationv1beta1.Webhook{
			{
				Name: podWebhookName,
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
						},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups: []string{
								v1.SchemeGroupVersion.Group,
							},
							APIVersions: []string{
								v1.SchemeGroupVersion.Version,
							},
							Resources: []string{
								podPlural,
							},
						},
					},
				},
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					Service: &admissionregistrationv1beta1.ServiceReference{
						Name:      cloudsqlPostgresOperatorServiceName,
						Namespace: w.namespace,
						Path:      &admissionPath,
					},
					CABundle: caBundle,
				},
				FailurePolicy: &podFailurePolicy,
			},
			{
				Name: postgresqlInstanceWebhookName,
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
							admissionregistrationv1beta1.Update,
							admissionregistrationv1beta1.Delete,
						},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups: []string{
								v1alpha1.SchemeGroupVersion.Group,
							},
							APIVersions: []string{
								v1alpha1.SchemeGroupVersion.Version,
							},
							Resources: []string{
								crds.PostgresqlInstancePlural,
							},
						},
					},
				},
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					Service: &admissionregistrationv1beta1.ServiceReference{
						Name:      cloudsqlPostgresOperatorServiceName,
						Namespace: w.namespace,
						Path:      &admissionPath,
					},
					CABundle: caBundle,
				},
				FailurePolicy: &postgresInstanceFailurePolicy,
			},
		},
	}
}
