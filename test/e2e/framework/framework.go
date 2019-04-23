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
	"fmt"

	"github.com/onsi/ginkgo"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	cloudsqlpostgresoperator "github.com/travelaudience/cloudsql-postgres-operator/pkg/client/clientset/versioned"
)

// Framework groups together utility methods and clients used by the end-to-end test suite.
type Framework struct {
	// KubeClient is a client to the Kubernetes API.
	KubeClient kubernetes.Interface
	// ProjectId is the ID of the GCP project where cloudsql-postgres-operator is managing instances.
	ProjectId string
	// SelfClient is a client to the "cloudsql.travelaudience.com" API.
	SelfClient cloudsqlpostgresoperator.Interface
}

// New returns a new instance of the framework.
func New(kubeconfig, projectId string) *Framework {
	// Create a Kubernetes configuration object.
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("failed to build kubeconfig: %v", err)
	}
	// Create a client for the Kubernetes API.
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed to build kubernetes client: %v", err)
	}
	// Create a client for the "cloudsql.travelaudience.com" API.
	selfClient, err := cloudsqlpostgresoperator.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed to build \"cloudsql.travelaudience.com\" client: %v", err)
	}
	// Return a new instance of the test framework.
	return &Framework{
		KubeClient: kubeClient,
		SelfClient: selfClient,
	}
}

// AdmissionIt is a wrapper for "ginkgo.It" that adds the "[Admission]" tag automatically.
func AdmissionIt(text string, body interface{}, timeout ...float64) bool {
	return ginkgo.It(fmt.Sprintf("%s [Admission]", text), body, timeout...)
}
