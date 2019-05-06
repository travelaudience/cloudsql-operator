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
	cloudsqladmin "google.golang.org/api/sqladmin/v1beta4"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	cloudsqlpostgresoperator "github.com/travelaudience/cloudsql-postgres-operator/pkg/client/clientset/versioned"
	googleutil "github.com/travelaudience/cloudsql-postgres-operator/pkg/util/google"
)

// Framework groups together utility methods and clients used by the end-to-end test suite.
type Framework struct {
	// CloudSQLClient is a client for the Cloud SQL Admin API.
	CloudSQLClient *cloudsqladmin.Service
	// ExternalIP is the external IP of the host where the end-to-end test suite is running on.
	ExternalIP string
	// KubeClient is a client to the Kubernetes API.
	KubeClient kubernetes.Interface
	// Namespace is the name of the namespace where cloudsql-postgres-operator is running.
	Namespace string
	// ProjectId is the ID of the GCP project where cloudsql-postgres-operator is managing instances.
	ProjectId string
	// SelfClient is a client to the "cloudsql.travelaudience.com" API.
	SelfClient cloudsqlpostgresoperator.Interface
}

// New returns a new instance of the framework.
func New(kubeconfig, namespace, pathToAdminKey, projectId string) *Framework {
	// Determine our external IP.
	ip, err := determineExternalIP()
	if err != nil {
		log.Fatalf("failed to determine our external ip: %v", err)
	}
	if ip == "" {
		log.Fatal("failed to determine our external ip")
	}
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
	// Create a client for the Cloud SQL Admin API.
	cloudsqlClient, err := googleutil.NewCloudSQLAdminClient(pathToAdminKey)
	if err != nil {
		log.Fatalf("failed to build cloud sql admin api client: %v", err)
	}
	// Return a new instance of the test framework.
	return &Framework{
		CloudSQLClient: cloudsqlClient,
		ExternalIP:     ip,
		KubeClient:     kubeClient,
		Namespace:      namespace,
		ProjectId:      projectId,
		SelfClient:     selfClient,
	}
}

// AdmissionIt is a wrapper for "ginkgo.It" that adds the "[Admission]" tag automatically.
func AdmissionIt(text string, body interface{}, timeout ...float64) bool {
	return ginkgo.It(fmt.Sprintf("%s [Admission]", text), body, timeout...)
}

// LifecycleIt is a wrapper for "ginkgo.It" that adds the "[Lifecycle]" tag automatically.
func LifecycleIt(text string, body interface{}, timeout ...float64) bool {
	return ginkgo.It(fmt.Sprintf("%s [Lifecycle]", text), body, timeout...)
}
