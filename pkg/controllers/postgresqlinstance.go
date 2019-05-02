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

package controllers

import (
	"fmt"

	cloudsqladmin "google.golang.org/api/sqladmin/v1beta4"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/util/slice"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	v1alpha1client "github.com/travelaudience/cloudsql-postgres-operator/pkg/client/clientset/versioned"
	v1alpha1informers "github.com/travelaudience/cloudsql-postgres-operator/pkg/client/informers/externalversions/cloudsql/v1alpha1"
	v1alpha1listers "github.com/travelaudience/cloudsql-postgres-operator/pkg/client/listers/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/configuration"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/crds"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/google"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/pointers"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/strings"
)

const (
	// postgresqlInstanceControllerName is the name of the controller for PostgresqlInstance resources.
	postgresqlInstanceControllerName = "postgresqlinstance-controller"
	// postgresqlInstanceControllerThreadiness is the number of workers controller for PostgresqlInstance resource will use to process items from its work queue.
	postgresqlInstanceControllerThreadiness = 1
	// databaseInstancePasswordLength is the length of the random password generated for every CSQLP instance.
	passwordLength = 36
	// passwordAlphabet is the alphabet used to generate the random password for CSQLP instances.
	passwordAlphabet = `abcdefghijklmnopqrstuvwxyz0123456789~!@#$%^&*()_-+={[}]|\:;"'<,>.?/`
)

// PostgresqlInstanceController is the controller for PostgresqlInstance resources.
type PostgresqlInstanceController struct {
	// PostgresqlInstanceController is based-off of a generic controller.
	*genericController
	// cloudsqlClient is a client for the Cloud SQL Admin API.
	cloudsqlClient *cloudsqladmin.Service
	// er is an EventRecorder through which we can emit events associated with PostgresqlInstance resources.
	er record.EventRecorder
	// kubeClient is a client to the Kubernetes API.
	kubeClient kubernetes.Interface
	// namespace is the namespace where cloudsql-postgres-operator is deployed.
	namespace string
	// postgresqlInstanceLister is a lister for PostgresqlInstance resources.
	postgresqlInstanceLister v1alpha1listers.PostgresqlInstanceLister
	// projectID is the ID of the GCP project where cloudsql-postgres-operator is managing CSQLP instances.
	projectID string
	// selfClient is a client to the "cloudsql.travelaudience.com" API.
	selfClient v1alpha1client.Interface
}

// NewPostgresqlInstance Controller creates a new instance of the controller for PostgresqlInstance resources.
func NewPostgresqlInstanceController(config configuration.Configuration, kubeClient kubernetes.Interface, selfClient v1alpha1client.Interface, er record.EventRecorder, postgresqlInstanceInformer v1alpha1informers.PostgresqlInstanceInformer, cloudsqlClient *cloudsqladmin.Service) *PostgresqlInstanceController {
	// Create a new instance of the controller for PostgresqlInstance resources using the specified name and threadiness.
	c := &PostgresqlInstanceController{
		cloudsqlClient:           cloudsqlClient,
		genericController:        newGenericController(postgresqlInstanceControllerName, postgresqlInstanceControllerThreadiness),
		er:                       er,
		kubeClient:               kubeClient,
		namespace:                config.Cluster.Namespace,
		postgresqlInstanceLister: postgresqlInstanceInformer.Lister(),
		projectID:                config.GCP.ProjectID,
		selfClient:               selfClient,
	}
	// Make the controller wait for the caches to sync.
	c.hasSyncedFuncs = []cache.InformerSynced{
		postgresqlInstanceInformer.Informer().HasSynced,
	}
	// Make "processQueueItem" the handler for items popped out of the work queue.
	c.syncHandler = c.processQueueItem

	// Setup an event handler to inform us when PostgresqlInstance resources change.
	postgresqlInstanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueue(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			c.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueue(obj)
		},
	})

	// Return the instance of the controller for PostgresqlInstance resources created above.
	return c
}

// processQueueItem attempts to reconcile the state of the PostgresqlInstance resource pointed at by the specified key.
func (c *PostgresqlInstanceController) processQueueItem(key string) error {
	// Grab the name of the PostgresqlInstance resource from the specified key.
	// NOTE: PostgresqlInstance is cluster-scoped, and hence there is no associated namespace.
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key %q", key))
		return nil
	}

	// Get the PostgresqlInstance resource with the specified name.
	p, err := c.postgresqlInstanceLister.Get(name)
	if err != nil {
		// The PostgresqlInstance may no longer exist, in which case we stop processing.
		if kubeerrors.IsNotFound(err) {
			c.logger.Debugf("postgresqlinstance %q in work queue no longer exists", key)
			return nil
		}
		return err
	}
	// Create a deep copy of the PostgresqlInstance resource so we don't possibly mutate the cache.
	p = p.DeepCopy()

	// Check whether the PostgresqlInstance resource is being deleted (indicated by a non-zero deletion timestamp).
	if p.DeletionTimestamp.IsZero() {
		// The PostgresqlInstance resource is not being deleted, so we must add the finalizer in case it is not already present.
		if !slice.ContainsString(p.Finalizers, constants.CleanupFinalizer, nil) {
			p.Finalizers = append(p.Finalizers, constants.CleanupFinalizer)
			if p, err = c.selfClient.CloudsqlV1alpha1().PostgresqlInstances().Update(p); err != nil {
				return err
			}
		}
	} else {
		// The PostgresqlInstance resource is being deleted, so we must delete the CSQLP instance and remove the finalizer.
		if slice.ContainsString(p.Finalizers, constants.CleanupFinalizer, nil) {
			if err := c.deleteInstance(p); err != nil {
				return err
			}
			p.Finalizers = slice.RemoveString(p.Finalizers, constants.CleanupFinalizer, nil)
			if _, err = c.selfClient.CloudsqlV1alpha1().PostgresqlInstances().Update(p); err != nil {
				return err
			}
		}
		// The finalizer has finished, so there is nothing else to do.
		return nil
	}

	// If the PostgresqlInstance resource is marked as being paused, stop processing immediately.
	if p.Spec.Paused {
		c.logger.Warnf("skipping paused postgresqlinstance %q", name)
		return nil
	}

	// Check whether a CSQLP instance with the specified ".spec.name" already exists, and create it if necessary.
	c.logger.Debugf("checking whether an instance with name %q already exists", p.Spec.Name)
	instance, err := c.cloudsqlClient.Instances.Get(c.projectID, p.Spec.Name).Do()
	if err != nil {
		// If we've got an error other than "404 NOT FOUND", we stop processing and propagate it.
		if !google.IsNotFound(err) {
			c.logger.Debugf("failed to check if an instance with name %q exists: %v", p.Spec.Name, err)
			return fmt.Errorf("failed to check if an instance with name %q exists: %v", p.Spec.Name, err)
		}
		// At this point we know that no instance having ".spec.name" as its name exists, so we proceed to creating it.
		if instance, err = c.createInstance(p); err != nil {
			// Creation of the CSQLP instance failed with a transient error.
			return err
		} else if instance == nil {
			// Creation of the CSQLP instance failed with a permanent error.
			return nil
		}
	}

	// Check whether the CSQLP instance is in a state other than "RUNNABLE", in which case we skip further processing (but don't error).
	// This may happen, for instance, if the CSQLP instance is still being created, or if it is down for maintenance.
	if instance.State != constants.DatabaseInstanceStateRunnable {
		c.logger.Infof("skipping sync of instance %q in state %q", p.Spec.Name, instance.State)
		return nil
	}
	// Check whether the CSQLP instance is currently running.
	// If it is not running, we also skip further processing (but don't error).
	if instance.Settings.ActivationPolicy != constants.DatabaseInstanceActivationPolicyAlways {
		c.logger.Infof("skipping sync of instance %q that is currently shut down", p.Spec.Name)
		return nil
	}

	// Create the secret associated with the current PostgresqlInstance resource already exists, if necessary.
	s, err := c.kubeClient.CoreV1().Secrets(c.namespace).Get(p.Name, metav1.GetOptions{})
	if err != nil {
		// If we've got an error otherr than "404 NOT FOUND", we stop processing and propagate it.
		if !kubeerrors.IsNotFound(err) {
			c.logger.Debugf("failed to check if the secret associated with %q already exists: %v", p.Name, err)
			return err
		}
		// At this point we know that the secret associated with the current PostgresqlInstance resource must be created.
		if s, err = c.createInstanceSecret(p); err != nil {
			c.logger.Debugf("failed to create the secret associated with %q: %v", p.Name, err)
			return err
		}
	}
	// Check whether we need to generate and set a password for the CSQLP instance.
	if password, exists := s.Data[constants.PostgresqlInstancePasswordKey]; !exists || len(password) == 0 {
		if err := c.setInstancePassword(p, s); err != nil {
			c.logger.Debugf("failed to set instance password for %q", p.Name)
			return err
		}
	}

	// Update the CSQLP instance's settings if necessary.
	instance, err = c.maybeUpdateInstance(p, instance)
	if err != nil {
		return err
	}
	return nil
}

// createInstance attempts to create a CSQLP instance based on the specified PostgresqlInstance resource.
func (c *PostgresqlInstanceController) createInstance(postgresqlInstance *v1alpha1api.PostgresqlInstance) (*cloudsqladmin.DatabaseInstance, error) {
	c.logger.Infof("creating instance %q", postgresqlInstance.Spec.Name)
	// Build the DatabaseInstance object based on the specified PostgresqlInstance resource.
	instance := buildDatabaseInstance(postgresqlInstance)
	// Attempt to create the DatabaseInstance object.
	_, err := c.cloudsqlClient.Instances.Insert(c.projectID, instance).Do()
	if err != nil {
		if google.IsConflict(err) {
			// We've been told that the instance needs to be created, but the Cloud SQL Admin API is reporting a conflict
			// This most probably means that a CSQLP instance with ".spec.name" as its name had previously existed but has been recently deleted.
			// Hence, we log but do not propagate the error, since subsequent attempts to create the instance are likely to fail as well until ".spec.name" becomes available again.
			c.logger.Errorf("the name %q seems to be unavailable - has an instance with such a name been deleted recently?", instance.Name)
			return nil, nil
		}
		if google.IsBadRequest(err) {
			// We've been told that the instance's specification is invalid.
			// This most probably means that the user has specified an invalid value for some field under ".spec".
			// Hence, we log but do not propagate the error, since subsequent attempts to create the instance are likely to fail as well until ".spec" is fixed.
			c.logger.Errorf("the instance's specification is invalid: %v", err)
			return nil, nil
		}
		// The Cloud SQL Admin API returned a different error, which we propagate so that creation may be retried.
		return nil, err
	}
	// Grab and return the most up-to-date representation of the CSQLP instance.
	return c.cloudsqlClient.Instances.Get(c.projectID, postgresqlInstance.Spec.Name).Do()
}

// createInstanceSecret creates the (initially empty) secret associated with the specified PostgresqlInstance resource.
func (c *PostgresqlInstanceController) createInstanceSecret(postgresqlInstance *v1alpha1api.PostgresqlInstance) (*corev1.Secret, error) {
	return c.kubeClient.CoreV1().Secrets(c.namespace).Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.LabelAppKey: constants.ApplicationName,
			},
			Name:      postgresqlInstance.Name,
			Namespace: c.namespace,
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
	})
}

// deleteInstance attempts to delete the CSQLP instance associated with the specified PostgresqlInstance resource.
func (c *PostgresqlInstanceController) deleteInstance(postgresqlInstance *v1alpha1api.PostgresqlInstance) error {
	c.logger.Debugf("checking whether instance %q needs to be deleted", postgresqlInstance.Spec.Name)
	// Before issuing a delete request (which can result in a "409 CONFLICT" response in case the CSQLP instance has already and recently been deleted), make sure the CSQLP instance is still listed.
	if _, err := c.cloudsqlClient.Instances.Get(c.projectID, postgresqlInstance.Spec.Name).Do(); err != nil {
		if google.IsNotFound(err) {
			c.logger.Debugf("instance %q has already been deleted", postgresqlInstance.Spec.Name)
			return nil
		}
		return err
	}
	c.logger.Infof("deleting instance %q", postgresqlInstance.Spec.Name)
	// At this point we know the CSQLP instance already exists, so we issue the delete request.
	if _, err := c.cloudsqlClient.Instances.Delete(c.projectID, postgresqlInstance.Spec.Name).Do(); err != nil {
		return err
	}
	c.logger.Debugf("instance %q has been deleted", postgresqlInstance.Spec.Name)
	return nil
}

// maybeUpdateInstance checks whether the settings for the CSQLP instance must be updated, and updates it if necessary.
func (c *PostgresqlInstanceController) maybeUpdateInstance(postgresqlInstance *v1alpha1api.PostgresqlInstance, databaseInstance *cloudsqladmin.DatabaseInstance) (*cloudsqladmin.DatabaseInstance, error) {
	c.logger.Debugf("checking whether the settings for instance %q must be updated", postgresqlInstance.Spec.Name)
	// Update the CSQLP instance's settings according to the PostgresqlInstance resource.
	mustUpdate := updateDatabaseInstanceSettings(postgresqlInstance, databaseInstance)
	if !mustUpdate {
		// No differences have been detected, so there is nothing to do.
		c.logger.Debugf("the settings for instance %q are up-to-date", postgresqlInstance.Spec.Name)
		return databaseInstance, nil
	}
	// At this point we know we have to update the CSQLP instance's settings.
	c.logger.Debugf("the settings for instance %q must be updated", postgresqlInstance.Spec.Name)
	_, err := c.cloudsqlClient.Instances.Update(c.projectID, databaseInstance.Name, databaseInstance).Do()
	if err != nil {
		if google.IsConflict(err) {
			// The Cloud SQL Admin API is reporting a conflict.
			// This most probably means that an update is already in progress, in which case we must wait.
			// Hence, we log but do not propagate the error, waiting until the next iteration of the controller to actually perform the update.
			c.logger.Errorf("conflict reported while trying to update the settings for instance %q - maybe another update is currently in progress?", postgresqlInstance.Spec.Name)
			return nil, nil
		}
		if google.IsBadRequest(err) {
			// We've been told that the instance's specification is invalid.
			// This most probably means that the user has specified an invalid value for some field under ".spec".
			// Hence, we log but do not propagate the error, since subsequent attempts to create the instance are likely to fail as well until ".spec" is fixed.
			c.logger.Errorf("the settings for instance %q are invalid: %v", postgresqlInstance.Spec.Name, err)
			return nil, nil
		}
		// The Cloud SQL Admin API returned a different error, which we propagate so that creation may be retried.
		return nil, err
	}
	return c.cloudsqlClient.Instances.Get(c.projectID, postgresqlInstance.Spec.Name).Do()
}

// setInstancePassword generates a random password for the CSQLP instance's "postgres" user, sets it on the CSQLP instance and writes it to the specified secret.
func (c *PostgresqlInstanceController) setInstancePassword(postgresqlInstance *v1alpha1api.PostgresqlInstance, secret *corev1.Secret) error {
	c.logger.Debugf("setting the %q user's password for instance %q", constants.PostgresqlInstanceUsernameValue, postgresqlInstance.Spec.Name)
	// Create a User object representing the "postgres" user and having a randomly-generated password.
	u := &cloudsqladmin.User{
		Name:     constants.PostgresqlInstanceUsernameValue,
		Password: strings.RandomStringWithLength(passwordLength, passwordAlphabet),
	}
	// Update the "postgres" user with the generated password.
	_, err := c.cloudsqlClient.Users.Update(c.projectID, postgresqlInstance.Spec.Name, u.Name, u).Do()
	if err != nil {
		return err
	}
	// Update the secret with the "postgres" user's username and password.
	if secret.StringData == nil {
		secret.StringData = make(map[string]string, 2)
	}
	secret.StringData[constants.PostgresqlInstanceUsernameKey] = u.Name
	secret.StringData[constants.PostgresqlInstancePasswordKey] = u.Password
	_, err = c.kubeClient.CoreV1().Secrets(secret.Namespace).Update(secret)
	return err
}
