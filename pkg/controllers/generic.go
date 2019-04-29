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
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller represents a controller that handles Kubernetes resources.
type Controller interface {
	// Run instructs the workers to start processing items from the work queue.
	Run(ctx context.Context) error
}

// genericController contains basic functionality shared by all controllers.
type genericController struct {
	// hasSyncedFuncs are the functions used to determine if caches are synced.
	hasSyncedFuncs []cache.InformerSynced
	// logger is the logger that the controller will use.
	logger log.FieldLogger
	// name is the name of the controller.
	name string
	// syncHandler is a function that takes a work item and processes it.
	syncHandler func(key string) error
	// threadiness is the number of workers to use for processing items from the work queue.
	threadiness int
	// workqueue is a rate limited work queue.
	// It is used to queue work to be processed instead of performing it as soon as a change happens.
	// This means we can ensure we only process a fixed amount of resources at a time, and makes it easy to ensure we are never processing the same resource simultaneously in two different worker goroutines.
	workqueue workqueue.RateLimitingInterface
}

// Run starts the controller, blocking until the specified context is canceled.
func (c *genericController) Run(ctx context.Context) error {
	// Handle any possible crashes and shutdown the work queue when we're done.
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.logger.Debugf("starting %q", c.name)

	// Wait for the caches to be synced before starting workers.
	c.logger.Debug("waiting for informer caches to be synced")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.hasSyncedFuncs...); !ok {
		return fmt.Errorf("failed to wait for informer caches to be synced")
	}

	c.logger.Debug("starting workers")

	// Launch "threadiness" workers to process items from the work queue.
	for i := 0; i < c.threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, ctx.Done())
	}

	c.logger.Debug("started workers")

	// Block until the context is canceled.
	<-ctx.Done()

	c.logger.Debug("stopped workers")

	return nil
}

// newGenericController returns a new generic controller.
func newGenericController(name string, threadiness int) *genericController {
	return &genericController{
		logger:      log.WithField("controller", name),
		workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
		threadiness: threadiness,
		name:        name,
	}
}

// runWorker is a long-running function that will continually call the processNextWorkItem function in order to read and process an item from the work queue.
func (c *genericController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the work queue and attempt to process it by calling syncHandler.
func (c *genericController) processNextWorkItem() bool {
	// Read an item from the work queue.
	obj, shutdown := c.workqueue.Get()
	// Return immediately if we've been told to shut down.
	if shutdown {
		return false
	}

	// Wrap this block in a func so we can defer the call to c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the work queue knows we have finished processing this item.
		// We also must remember to call Forget if we do not want this work item to be re-queued.
		// For example, we do not call Forget if a transient error occurs.
		// Instead the item is put back on the work queue and attempted again after a back-off period.
		defer c.workqueue.Done(obj)

		var (
			key string
			ok  bool
		)

		// We expect strings to come off the work queue.
		// These are of the form "namespace/name".
		// We do this as the delayed nature of the work queue means the items in the informer cache may actually be more up-to-date than when the item was initially put onto the work queue.
		if key, ok = obj.(string); !ok {
			// As the item in the work queue is actually invalid, we call "Forget" here else we'd go into a loop of attempting to process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected a string to come off of the work queue but got %#v", obj))
			return nil
		}
		// Call "syncHandler", passing it the "namespace/name" string that corresponds to the resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the work queue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing %q: %s", key, err.Error())
		}
		// Finally, and if no error occurs, we call "Forget" on this item so it does not get queued again until another change happens.
		c.workqueue.Forget(obj)
		c.logger.Debugf("successfully synced %q", key)
		return nil
	}(obj)

	// If we've got an error, pass it to "HandleError" so a back-off behavior can be applied.
	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

// enqueue takes a resource, computes its resource key and puts it as a work item onto the work queue.
func (c *genericController) enqueue(obj interface{}) {
	var (
		err error
		key string
	)
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
