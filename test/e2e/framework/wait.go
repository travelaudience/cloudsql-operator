// +build e2e

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

package framework

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	watchapi "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/watch"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
)

// WaitUntilPodCondition blocks until the specified condition function is verified, or until the provided context times out.
func (f *Framework) WaitUntilPodCondition(ctx context.Context, pod *corev1.Pod, fn watch.ConditionFunc) error {
	// Create a selector that targets the specified PostgresqlInstance resource.
	fs := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.namespace==%s,metadata.name==%s", pod.Namespace, pod.Name))
	// Grab a ListerWatcher with which we can watch the Ingress resource.
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fs.String()
			return f.KubeClient.CoreV1().Pods(pod.Namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watchapi.Interface, error) {
			options.FieldSelector = fs.String()
			return f.KubeClient.CoreV1().Pods(pod.Namespace).Watch(options)
		},
	}
	// Watch for updates to the specified Ingress resource until fn is satisfied.
	last, err := watch.UntilWithSync(ctx, lw, &corev1.Pod{}, nil, fn)
	if err != nil {
		return err
	}
	if last == nil {
		return fmt.Errorf("no events received for pod \"%s/%s\"", pod.Namespace, pod.Name)
	}
	return nil
}

// WaitUntilPodRunning blocks until the specified pod is running and ready.
func (f *Framework) WaitUntilPodRunning(ctx context.Context, pod *corev1.Pod) error {
	return f.WaitUntilPodCondition(ctx, pod, func(event watchapi.Event) (bool, error) {
		switch event.Type {
		case watchapi.Added, watchapi.Modified:
			return event.Object.(*corev1.Pod).Status.Phase == corev1.PodRunning, nil
		default:
			return false, fmt.Errorf("got event of unexpected type %q", event.Type)
		}
	})
}

// WaitUntilPostgresqlInstanceCondition blocks until the specified condition function is verified, or until the provided context times out.
func (f *Framework) WaitUntilPostgresqlInstanceCondition(ctx context.Context, postgresqlInstance *v1alpha1api.PostgresqlInstance, fn watch.ConditionFunc) error {
	// Create a selector that targets the specified PostgresqlInstance resource.
	fs := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name==%s", postgresqlInstance.Name))
	// Grab a ListerWatcher with which we can watch the Ingress resource.
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fs.String()
			return f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watchapi.Interface, error) {
			options.FieldSelector = fs.String()
			return f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Watch(options)
		},
	}
	// Watch for updates to the specified Ingress resource until fn is satisfied.
	last, err := watch.UntilWithSync(ctx, lw, &v1alpha1api.PostgresqlInstance{}, nil, fn)
	if err != nil {
		return err
	}
	if last == nil {
		return fmt.Errorf("no events received for postgresqlinstance %q", postgresqlInstance.Name)
	}
	return nil
}

// WaitUntilPostgresqlInstanceStatusCondition blocks until the specified condition type is found on the PostgresqlInstance resource's ".status.conditions" field with the provided status.
func (f *Framework) WaitUntilPostgresqlInstanceStatusCondition(ctx context.Context, postgresqlInstance *v1alpha1api.PostgresqlInstance, conditionType v1alpha1api.PostgresqlInstanceStatusConditionType, conditionStatus corev1.ConditionStatus) error {
	return f.WaitUntilPostgresqlInstanceCondition(ctx, postgresqlInstance, func(event watchapi.Event) (bool, error) {
		switch event.Type {
		case watchapi.Added, watchapi.Modified:
			cnd, err := GetPostgresqlInstanceCondition(event.Object.(*v1alpha1api.PostgresqlInstance), v1alpha1api.PostgresqlInstanceStatusConditionTypeReady)
			if err != nil {
				return false, nil
			}
			return cnd.Status == conditionStatus, nil
		default:
			return false, fmt.Errorf("got event of unexpected type %q", event.Type)
		}
	})
}
