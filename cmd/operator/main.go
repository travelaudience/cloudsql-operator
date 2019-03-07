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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

	"github.com/travelaudience/cloudsql-operator/pkg/configuration"
	"github.com/travelaudience/cloudsql-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-operator/pkg/crds"
	"github.com/travelaudience/cloudsql-operator/pkg/signals"
	"github.com/travelaudience/cloudsql-operator/pkg/version"
)

var (
	// configFile holds the path to the configuration file.
	configFile string
)

var (
	// config holds the configuration object.
	config configuration.Configuration
)

func init() {
	flag.StringVar(&configFile, "config-file", "", "the path to the configuration file")
}

func main() {
	// Parse the provided command-line flags.
	flag.Parse()

	// Parse the provided configuration file or use a default configuration.
	if configFile != "" {
		config = configuration.MustNewConfigurationFromFile(configFile)
	} else {
		config = configuration.NewDefaultConfiguration()
	}

	// Enable logging at the requested level.
	if v, err := log.ParseLevel(config.Logging.Level); err != nil {
		log.Fatalf("%q is not a valid log level", config.Logging.Level)
	} else {
		log.SetLevel(v)
	}

	// Setup a signal handler so we can gracefully shutdown when requested to.
	stopCh := signals.SetupSignalHandler()
	// Birth cry.
	log.WithField(constants.Version, version.Version).Infof("%s is starting", constants.ApplicationName)

	// Create a Kubernetes configuration object.
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Cluster.Kubeconfig)
	if err != nil {
		log.Fatalf("failed to build kubeconfig: %v", err)
	}
	// Create a Kubernetes client.
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed to build kubernetes client: %v", err)
	}

	// Create an event recorder so we can emit events during leader election and afterwards.
	eb := record.NewBroadcaster()
	eb.StartLogging(log.Debugf)
	eb.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	er := eb.NewRecorder(scheme.Scheme, corev1.EventSource{Component: constants.ApplicationName})

	// Compute the identity of the current instance of the application based on the current hostname.
	// The hostname is appended with a unique identifier in order to prevent two instances running on the same host from becoming active.
	id, err := os.Hostname()
	if err != nil {
		log.Fatalf("failed to compute identity: %v", err)
	}
	id = fmt.Sprintf("%s-%s", id, string(uuid.NewUUID()))

	// Setup a resource lock so we can perform leader election.
	rl, _ := resourcelock.New(
		resourcelock.EndpointsResourceLock,
		config.Cluster.Namespace,
		constants.ApplicationName,
		kubeClient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: er,
		},
	)

	// Perform leader election so that at most a single instance of the application is active at any given moment.
	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// We've started leading, so we can start our controllers.
				// The controllers will run under the specified context, and will stop whenever said context is canceled.
				// However, we must also make sure that they stop whenever we receive a shutdown signal.
				// Hence, we must create a new child context and wait in a separate goroutine for "stopCh" to be notified of said shutdown signal.
				ctx, fn := context.WithCancel(ctx)
				defer fn()
				go func() {
					<-stopCh
					fn()
				}()
				run(ctx, kubeConfig)
			},
			OnStoppedLeading: func() {
				// We've stopped leading, so we must exit immediately.
				log.Fatalf("leader election lost")
			},
			OnNewLeader: func(identity string) {
				// Report who the current leader is for debugging purposes.
				log.Debugf("current leader: %s", identity)
			},
		},
	})
}

func run(ctx context.Context, kubeConfig *rest.Config) {
	// Create a client for the apiextensions.k8s.io/v1beta1 so that we can create or update our CRDs.
	extsClient, err := extsclientset.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed to build kubernetes client: %v", err)
	}
	// Create or update our CRDs.
	if err := crds.CreateOrUpdateCRDs(extsClient); err != nil {
		log.Fatalf("failed to create or update crds: %v", err)
	}
	// Wait for the context to be canceled.
	<-ctx.Done()
	// Confirm successful shutdown.
	log.WithField(constants.Version, version.Version).Infof("%s is shutting down", constants.ApplicationName)
	// There is a goroutine in the background trying to renew the leader election lock.
	// Hence, we must manually exit now that we know controllers have been properly shutdown.
	os.Exit(0)
}
