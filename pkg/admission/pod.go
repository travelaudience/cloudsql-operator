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
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/crds"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/pointers"
)

const (
	// clientServiceAccountKeyKey is the name of the key containing the JSON credentials for the IAM service account with the "roles/cloudsql.client" role.
	clientServiceAccountKeyKey = "credentials.json"
	// cloudSQLProxyContainerName is the name of the Cloud SQL proxy container injected in each pod.
	cloudSQLProxyContainerName = "cloud-sql-proxy"
	// cloudSQLProxyContainerPortMinValue is the maximum value to use when drawing a random port number for the Cloud SQL proxy container.
	cloudSQLProxyContainerPortMaxValue = 65535
	// cloudSQLProxyContainerPortMinValue is the minimum value to use when drawing a random port number for the Cloud SQL proxy container.
	cloudSQLProxyContainerPortMinValue = 49152
	// credentialsSecretVolumeName is the name of the volume containing the credentials for connecting to the CSQLP instance.
	credentialsSecretVolumeName = "credentials"
	// credentialsSecretVolumeMountPath is the path where the secret containing the credentials for connecting to the CSQLP instance will be mounted.
	credentialsSecretVolumeMountPath = "/secret"
	// ipAddressTypePublic is the value used to indicate to Cloud SQL proxy that it should connect to a CSQLP instance via its public IP.
	ipAddressTypePublic = "PUBLIC"
	// ipAddressTypePrivate  is the value used to indicate to Cloud SQL proxy that it should connect to a CSQLP instance via its private IP.
	ipAddressTypePrivate = "PRIVATE"
	// pghostEnvVarName is the name of the "PGHOST" environment variable injected in each container.
	pghostEnvVarName = "PGHOST"
	// pghostEnvVarValue is the value of the "PGHOST" environment variable injected in each container.
	pghostEnvVarValue = "localhost"
	// pgportEnvVarName is the name of the "PGPORT" environment variable injected in each container.
	pgportEnvVarName = "PGPORT"
	// pguserEnvVarName is the name of the "PGUSER" environment variable injected in each container.
	pguserEnvVarName = "PGUSER"
	// pgpassfileEnvVarName is the name of the "PGPASSFILE" environment variable injected in each container.
	pgpassfileEnvVarName = "PGPASSFILE"
	// pgpassConfKey is the name of the key containing the username and password combination for the CSQLP instance.
	pgpassConfKey = "pgpass.conf"
	// pgpassConfValueFormatString is the format string used when creating the file containing the username and password combination for the CSQLP instance.
	pgpassConfValueFormatString = "*:*:*:%s:%s"
	// secretNameFormatString is the name of the secret used to store the credentials for connecting to the CSQLP instance.
	secretNameFormatString = "%s-cloud-sql-proxy"
)

// mutatePodInternal checks whether the provided Pod resource is requesting access to a CSQLP instance, and performs injection of the Cloud SQL proxy sidecar.
func (w *Webhook) mutatePod(namespace string, currentObj *corev1.Pod) (*corev1.Pod, error) {
	pod, err := func() (*corev1.Pod, error) {
		// Check whether we have been asked to connect to a CSQLP instance.
		v, exists := currentObj.Annotations[constants.PostgresqlInstanceNameAnnotationKey]
		if !exists || v == "" {
			return currentObj, nil
		}

		// Clone the current object so that we can safely mutate it.
		mutatedObj := currentObj.DeepCopy()

		// Check whether the referenced PostgresqlInstance resource exists or not.
		postgresqlInstance, err := w.selfClient.CloudsqlV1alpha1().PostgresqlInstances().Get(v, metav1.GetOptions{})
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				return nil, fmt.Errorf("postgresqlinstance %q does not exist: %v", v, err)
			}
			return nil, fmt.Errorf("failed to get postgresql instance %q: %v", v, err)
		}

		// Grab the secret associated with the PostgresqlInstance resource.
		postgresqlInstanceSecret, err := w.kubeClient.CoreV1().Secrets(w.namespace).Get(postgresqlInstance.Name, metav1.GetOptions{})
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				return nil, fmt.Errorf("the secret associated with postgresqlinstance %q does not exist: %v", postgresqlInstance.Name, err)
			}
			return nil, fmt.Errorf("failed to get the secret associated with postgresqlinstance %q: %v", postgresqlInstance.Name, err)
		}

		// Make sure that the connection name for the PostgresqlInstance has already been reported.
		if postgresqlInstance.Status.ConnectionName == "" {
			return nil, fmt.Errorf("failed to get the connection name associated with postgresqlinstance %q: %v", postgresqlInstance.Name, err)
		}

		// Build the Secret object that represents the desired state of the namespace-local secret containing "pgpass.conf" for the PostgresqlInstance resource.
		localPostgresqlInstanceSecretName := fmt.Sprintf(secretNameFormatString, postgresqlInstance.Name)
		localPostgresqlInstanceSecret := w.buildLocalPostgresqlInstanceSecret(namespace, localPostgresqlInstanceSecretName, postgresqlInstance, postgresqlInstanceSecret)

		// Make sure the namespace-local secret containing "pgpass.conf" for the PostgresqlInstance resource exists and is up-to-date.
		_, err = w.kubeClient.CoreV1().Secrets(localPostgresqlInstanceSecret.Namespace).Create(localPostgresqlInstanceSecret)
		if err != nil {
			if !kubeerrors.IsAlreadyExists(err) {
				return nil, fmt.Errorf("failed to create the local secret associated with postgresqlinstance %q: %v", postgresqlInstance.Name, err)
			}
			// At this point we know that the namespace-local secret containing "pgpass.conf" for the PostgresqlInstance resource already exists.
			// This most probably means that a Pod requesting access to the current CSQLP instance has already been created at some point.
			// However, the secret may not be up-to-date, so we patch it as necessary in order to make sure its contents are valid.
			currentLocalPostgresqlInstanceSecret, err := w.kubeClient.CoreV1().Secrets(localPostgresqlInstanceSecret.Namespace).Get(localPostgresqlInstanceSecret.Name, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get the local secret associated with postgresqlinstance %q: %v", postgresqlInstance.Name, err)
			}
			updatedLocalPostgresqlInstanceSecret := currentLocalPostgresqlInstanceSecret.DeepCopy()
			updatedLocalPostgresqlInstanceSecret.StringData = localPostgresqlInstanceSecret.StringData
			_, err = w.patchSecret(currentLocalPostgresqlInstanceSecret, updatedLocalPostgresqlInstanceSecret)
			if err != nil {
				return nil, fmt.Errorf("failed to patch the local secret associated with postgresqlinstance %q: %v", postgresqlInstance.Name, err)
			}
		}

		// Add the namespace-local secret as a volume.
		mutatedObj.Spec.Volumes = append(mutatedObj.Spec.Volumes, corev1.Volume{
			Name: credentialsSecretVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					// Use 0400 as the default mode for files created as a result of mounting the secret.
					DefaultMode: pointers.NewInt32(256),
					Optional:    pointers.NewBool(false),
					SecretName:  localPostgresqlInstanceSecretName,
				},
			},
		})

		// Draw a random port to use for the Cloud SQL proxy.
		port := getFreeRandomPort(mutatedObj)

		// Modify existing containers in order to mount the namespace-local secret as a volume and to inject the required "PG*" variables.
		for idx := range mutatedObj.Spec.Containers {
			c := &mutatedObj.Spec.Containers[idx]
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				MountPath: credentialsSecretVolumeMountPath,
				Name:      credentialsSecretVolumeName,
				ReadOnly:  true,
			})
			c.Env = append(c.Env, []corev1.EnvVar{
				{
					Name:  pghostEnvVarName,
					Value: pghostEnvVarValue,
				},
				{
					Name:  pgportEnvVarName,
					Value: strconv.Itoa(int(port)),
				},
				{
					Name:  pguserEnvVarName,
					Value: string(postgresqlInstanceSecret.Data[constants.PostgresqlInstanceUsernameKey]),
				},
				{
					Name:  pgpassfileEnvVarName,
					Value: path.Join(credentialsSecretVolumeMountPath, pgpassConfKey),
				},
			}...)
		}

		// Inject the Cloud SQL proxy container.
		mutatedObj.Spec.Containers = append(mutatedObj.Spec.Containers, w.buildCloudSQLProxyContainer(postgresqlInstance, port))

		// Signal that the Cloud SQL proxy sidecar has been injected and return.
		mutatedObj.Annotations[constants.ProxyInjectedAnnotationKey] = "true"
		return mutatedObj, nil
	}()
	if err != nil {
		// Log the error, associating it with the namespace and name of the pod being processed.
		log.WithFields(log.Fields{
			"namespace": namespace,
			"pod": currentObj.Name,
		}).Error(err.Error())
	}
	return pod, err
}

// buildCloudSQLProxyContainer builds the Cloud SQL proxy container to inject.
func (w *Webhook) buildCloudSQLProxyContainer(postgresqlInstance *v1alpha1api.PostgresqlInstance, port int32) corev1.Container {
	ipAddressTypes := make([]string, 0)
	if *postgresqlInstance.Spec.Networking.PublicIP.Enabled {
		ipAddressTypes = append(ipAddressTypes, ipAddressTypePublic)
	}
	if *postgresqlInstance.Spec.Networking.PrivateIP.Enabled {
		ipAddressTypes = append(ipAddressTypes, ipAddressTypePrivate)
	}
	return corev1.Container{
		Name:  cloudSQLProxyContainerName,
		Image: w.cloudsqlProxyImage,
		Command: []string{
			"/cloud_sql_proxy",
			fmt.Sprintf("-credential_file=%s", path.Join(credentialsSecretVolumeMountPath, clientServiceAccountKeyKey)),
			fmt.Sprintf("-instances=%s=tcp:%d", postgresqlInstance.Status.ConnectionName, port),
			fmt.Sprintf("-ip_address_types=%s", strings.Join(ipAddressTypes, ",")),
		},
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: port,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: credentialsSecretVolumeMountPath,
				Name:      credentialsSecretVolumeName,
				ReadOnly:  true,
			},
		},
	}
}

// buildLocalPostgresqlInstanceSecret builds the namespace-local secret containing the "pgpass.conf" file used to connect to the CSQLP instance represented by the provided PostgresqlInstance resource.
func (w *Webhook) buildLocalPostgresqlInstanceSecret(namespace, name string, postgresqlInstance *v1alpha1api.PostgresqlInstance, postgresqlInstanceSecret *corev1.Secret) *corev1.Secret {
	c := w.clientServiceAccountKey
	u := string(postgresqlInstanceSecret.Data[constants.PostgresqlInstanceUsernameKey])
	p := strings.Replace(strings.Replace(string(postgresqlInstanceSecret.Data[constants.PostgresqlInstancePasswordKey]), `\`, `\\`, -1), `:`, `\:`, -1)
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.LabelAppKey: constants.ApplicationName,
			},
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         v1alpha1api.SchemeGroupVersion.String(),
					Kind:               crds.PostgresqlInstanceKind,
					Name:               postgresqlInstance.Name,
					UID:                postgresqlInstance.UID,
					Controller:         pointers.NewBool(true),
					BlockOwnerDeletion: pointers.NewBool(true),
				},
			},
		},
		StringData: map[string]string{
			clientServiceAccountKeyKey: c,
			pgpassConfKey:              fmt.Sprintf(pgpassConfValueFormatString, u, p),
		},
	}
	return s
}

// getFreeRandomPort returns a random port drawn from the random port range (49152-65535) that is not already in use in the provided pod.
func getFreeRandomPort(pod *corev1.Pod) int32 {
	// Build the map of used ports by iterating over every container.
	usedPorts := make(map[int32]bool, 0)
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			usedPorts[port.ContainerPort] = true
		}
	}
	// Draw a random port that is not already in use and return it once found.
	for {
		port := int32(rand.IntnRange(cloudSQLProxyContainerPortMinValue, cloudSQLProxyContainerPortMaxValue))
		if _, exists := usedPorts[port]; !exists {
			return port
		}
	}
}

// patchPostgresqlInstance updates the provided PostgresqlInstance using patch semantics.
// If there are no changes to be made, no patch is performed.
func (w *Webhook) patchSecret(oldObj, newObj *corev1.Secret) (*corev1.Secret, error) {
	// Return if there are no changes to be made.
	if reflect.DeepEqual(oldObj, newObj) {
		return newObj, nil
	}
	// Prepare the patch to apply based on the provided objects.
	oldBytes, err := json.Marshal(oldObj)
	if err != nil {
		return nil, err
	}
	newBytes, err := json.Marshal(newObj)
	if err != nil {
		return nil, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldBytes, newBytes, &v1alpha1api.PostgresqlInstance{})
	if err != nil {
		return nil, err
	}
	// Apply the patch.
	return w.kubeClient.CoreV1().Secrets(oldObj.Namespace).Patch(oldObj.Name, types.MergePatchType, patchBytes)
}
