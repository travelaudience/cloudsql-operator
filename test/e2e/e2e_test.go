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

package e2e_test

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	e2eframework "github.com/travelaudience/cloudsql-postgres-operator/test/e2e/framework"
)

var (
	// kubeconfig is the path to the kubeconfig file to use when running outside a Kubernetes cluster.
	kubeconfig string
	// logLevel is the log level to use while running the end-to-end test suite.
	logLevel string
	// namespace is the name of the namespace where cloudsql-postgres-operator is running.
	namespace string
	// pathToAdminKey is the path to a file containing credentials for an Iam service account with the "roles/cloudsql.admin" role.
	pathToAdminKey string
	// projectId is the ID of the GCP project where cloudsql-postgres-operator is managing instances.
	projectId string
)

var (
	// framework is the instance of the test framework to be used.
	f *e2eframework.Framework
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "the path to the kubeconfig file to use")
	flag.StringVar(&logLevel, "log-level", log.InfoLevel.String(), "the log level to use while running the end-to-end test suite")
	flag.StringVar(&namespace, "namespace", constants.ApplicationName, "the name of the namespace where cloudsql-postgres-operator is running")
	flag.StringVar(&pathToAdminKey, "path-to-admin-key", "", `the path to a file containing credentials for an iam service account with the "roles/cloudsql.admin" role`)
	flag.StringVar(&projectId, "project-id", "", "the id of the gcp project where cloudsql-postgres-operator is managing instances")
	flag.Parse()
}

var _ = BeforeSuite(func() {
	// Create a new instance of the test framework.
	f = e2eframework.New(kubeconfig, namespace, pathToAdminKey, projectId)
})

// TestEndToEnd runs the end-to-end test suite.
func TestEndToEnd(t *testing.T) {
	// Initialize our source of randomness.
	rand.Seed(time.Now().UnixNano())

	// Set the log level to the requested value.
	if l, err := log.ParseLevel(logLevel); err != nil {
		log.Fatal(err)
	} else {
		log.SetLevel(l)
	}

	// Register a failure handler and run the end-to-end test suite.
	RegisterFailHandler(Fail)
	RunSpecs(t, fmt.Sprintf("%s end-to-end test suite", constants.ApplicationName))
}
