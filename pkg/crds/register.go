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

package crds

import (
	"context"
	"fmt"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	extsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	watchapi "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/watch"

	"github.com/travelaudience/cloudsql-operator/pkg/constants"
	kubernetesutil "github.com/travelaudience/cloudsql-operator/pkg/util/kubernetes"
)

const (
	// waitCRDEstablishedTimeout specifies how long we wait for a given CustomResourceDefinition resource to become "Established" before timing out.
	waitCRDEstablishedTimeout = 15 * time.Second
)

// RegisterCRDs creates or updates our CustomResourceDefinition resources, waiting for them to reach the "Established" condition.
func CreateOrUpdateCRDs(extsClient extsclientset.Interface) error {
	for _, crd := range crds {
		// Create the CustomResourceDefinition resource.
		if err := createCRD(extsClient, crd); err != nil {
			return err
		}
		// Wait for the CustomResourceDefinition resource to reach the "Established" condition.
		if err := waitCRDEstablished(extsClient, crd); err != nil {
			return err
		}
	}
	return nil
}

// createCRD creates the specified CustomResourceDefinition resource.
func createCRD(extsClient extsclientset.Interface, crd *extsv1beta1.CustomResourceDefinition) error {
	// Attempt to create the CustomResourceDefinition resource.
	log.WithField(constants.Kind, crd.Spec.Names.Kind).Debug("creating crd")
	_, err := extsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err == nil {
		// Creation was successful, so there's nothing else to do.
		return nil
	}
	if !errors.IsAlreadyExists(err) {
		// The CustomResourceDefinition resource doesn't exist yet, but we've got an unexpected error while creating it.
		return err
	}

	// At this point the CustomResourceDefinition resource already exists but its spec may differ (since our API is not yet stable).
	log.WithField(constants.Kind, crd.Spec.Names.Kind).Debug("crd already exists")

	// Read the latest version of the CustomResourceDefinition resource.
	d, err := extsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// If the specs match, there is nothing to do.
	if reflect.DeepEqual(d.Spec, crd.Spec) {
		return nil
	}

	log.WithField(constants.Kind, crd.Spec.Names.Kind).Debug("updating crd")

	// Overwrite the CustomResourceDefinition resource's current ".spec" field with the desired one.
	d.Spec = crd.Spec
	if _, err := extsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(d); err != nil {
		return err
	}

	log.WithField(constants.Kind, crd.Spec.Names.Kind).Debug("crd has been updated")
	return nil
}

// isCRDEstablished returns whether the specified CustomResourceDefinition resource has reached the "Established" condition.
func isCRDEstablished(crd *extsv1beta1.CustomResourceDefinition) bool {
	for _, cond := range crd.Status.Conditions {
		if cond.Type == extsv1beta1.Established {
			if cond.Status == extsv1beta1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

// waitCRDEstablished blocks until the specified CustomResourceDefinition resource has reached the "Established" condition.
func waitCRDEstablished(extsClient extsclientset.Interface, crd *extsv1beta1.CustomResourceDefinition) error {
	// Grab a ListerWatcher with which we can watch the CustomResourceDefinition resource.
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = kubernetesutil.ByCoordinates(crd.Namespace, crd.Name).String()
			return extsClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watchapi.Interface, error) {
			options.FieldSelector = kubernetesutil.ByCoordinates(crd.Namespace, crd.Name).String()
			return extsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Watch(options)
		},
	}

	log.WithField(constants.Kind, crd.Spec.Names.Kind).Debug("waiting for crd to become established")

	// Watch for updates to the specified CustomResourceDefinition resource until it reaches the "Established" condition or until "waitCRDReadyTimeout" elapses.
	ctx, fn := context.WithTimeout(context.Background(), waitCRDEstablishedTimeout)
	defer fn()
	last, err := watch.UntilWithSync(ctx, lw, &extsv1beta1.CustomResourceDefinition{}, nil, func(event watchapi.Event) (bool, error) {
		// Grab the current resource from the event.
		obj := event.Object.(*extsv1beta1.CustomResourceDefinition)
		// Return true if and only if the CustomResourceDefinition resource has reached the "Established" condition.
		return isCRDEstablished(obj), nil
	})
	if err != nil {
		// We've got an error while watching the specified CustomResourceDefinition resource.
		return err
	}
	if last == nil {
		// We've got no events for the CustomResourceDefinition resource, which represents an error.
		return fmt.Errorf("no events received for crd %q", crd.Name)
	}

	// At this point we are sure the CustomResourceDefinition resource has reached the "Established" condition, so we return.
	log.WithField(constants.Kind, crd.Spec.Names.Kind).Debug("crd established")
	return nil
}
