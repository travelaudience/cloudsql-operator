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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	v1alpha1informers "github.com/travelaudience/cloudsql-postgres-operator/pkg/client/informers/externalversions/cloudsql/v1alpha1"
	v1alpha1listers "github.com/travelaudience/cloudsql-postgres-operator/pkg/client/listers/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/configuration"
)

const (
	// postgresqlInstanceControllerName is the name of the controller for PostgresqlInstance resources.
	postgresqlInstanceControllerName = "postgresqlinstance-controller"
	// postgresqlInstanceControllerThreadiness is the number of workers controller for PostgresqlInstance resource will use to process items from its work queue.
	postgresqlInstanceControllerThreadiness = 1
)

// PostgresqlInstanceController is the controller for PostgresqlInstance resources.
type PostgresqlInstanceController struct {
	// PostgresqlInstanceController is based-off of a generic controller.
	*genericController
	// cloudsqlClient is a client for the Cloud SQL Admin API.
	cloudsqlClient *cloudsqladmin.Service
	// kubeClient is a client to the Kubernetes API.
	kubeClient kubernetes.Interface
	// er is an EventRecorder through which we can emit events associated with PostgresqlInstance resources.
	er record.EventRecorder
	// postgresqlInstanceLister is a lister for PostgresqlInstance resources.
	postgresqlInstanceLister v1alpha1listers.PostgresqlInstanceLister
	// projectID is the ID of the GCP project where cloudsql-postgres-operator is managing CSQLP instances.
	projectID string
}

// NewPostgresqlInstance Controller creates a new instance of the controller for PostgresqlInstance resources.
func NewPostgresqlInstanceController(config configuration.Configuration, kubeClient kubernetes.Interface, er record.EventRecorder, postgresqlInstanceInformer v1alpha1informers.PostgresqlInstanceInformer, cloudsqlClient *cloudsqladmin.Service) *PostgresqlInstanceController {
	// Create a new instance of the controller for PostgresqlInstance resources using the specified name and threadiness.
	c := &PostgresqlInstanceController{
		cloudsqlClient:           cloudsqlClient,
		genericController:        newGenericController(postgresqlInstanceControllerName, postgresqlInstanceControllerThreadiness),
		kubeClient:               kubeClient,
		er:                       er,
		postgresqlInstanceLister: postgresqlInstanceInformer.Lister(),
		projectID:                config.GCP.ProjectID,
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
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("postgresqlinstance %q in work queue no longer exists", key))
			return nil
		}
		return err
	}

	// If the PostgresqlInstance resource is marked as being paused, stop processing immediately.
	if p.Spec.Paused {
		c.logger.Warnf("skipping paused postgresqlinstance %q", name)
		return nil
	}

	// TODO: Implement.

	return nil
}
