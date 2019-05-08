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
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/crds"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/pointers"
)

const (
	// PostgresqlTestPodInitialDelay represents how long after the test pod is launched will the test "psql" command be executed.
	PostgresqlTestPodInitialDelay = 5 * time.Second
)

// CreatePostgresqlTestPod creates a pod requesting access to the CSQLP instance associated with the provided PostgresqlInstance resource.
// A few seconds after the pod starts running, the specified SQL command will be executed using "psql".
// This initial delay is meant to give time to the Cloud SQL proxy sidecar to be ready to accept connections.
// The test pod is garbage-collected once the provided PostgresqlInstance resource is deleted.
func (f *Framework) CreatePostgresqlTestPod(postgresqlInstance *v1alpha1api.PostgresqlInstance, sqlCommand string) (*corev1.Pod, error) {
	return f.KubeClient.CoreV1().Pods(corev1.NamespaceDefault).Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				constants.PostgresqlInstanceNameAnnotationKey: postgresqlInstance.Name,
			},
			GenerateName: fmt.Sprintf("%s-e2e-", constants.ApplicationName),
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
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "postgres",
					Image: "postgres:9.6",
					Command: []string{
						"/bin/bash",
						"-c",
						fmt.Sprintf("sleep %d && psql -c '%s'", int(PostgresqlTestPodInitialDelay.Seconds()), sqlCommand),
					},
				},
			},
		},
	})
}

// WaitUntilPodLogLineMatches waits until a line in the logs for the first container of the provided pod matches the provided regular expression.
func (f *Framework) WaitUntilPodLogLineMatches(ctx context.Context, pod *corev1.Pod, regex string) error {
	req := f.KubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: pod.Spec.Containers[0].Name,
		Follow:    true,
	})
	r, err := req.Stream()
	if err != nil {
		return err
	}
	defer r.Close()
	rd := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to find a matching log line within the specified timeout")
		default:
			// Read a single line from the logs and check whether it matches the specified regular expression.
			str, err := rd.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF {
				return fmt.Errorf("failed to find a matching log line before EOF")
			}
			m, err := regexp.MatchString(regex, str)
			if err != nil {
				return err
			}
			// If the current log line matches the regular expression, return.
			if m {
				return nil
			}
		}
	}
}
