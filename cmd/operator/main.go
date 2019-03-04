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
	log "github.com/sirupsen/logrus"

	"github.com/travelaudience/cloudsql-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-operator/pkg/signals"
	"github.com/travelaudience/cloudsql-operator/pkg/version"
)

func main() {
	// Setup a signal handler so we can gracefully shutdown when requested to.
	stopCh := signals.SetupSignalHandler()
	// Birth cry.
	log.WithField(constants.Version, version.Version).Infof("%s is starting", constants.ApplicationName)
	// Wait for the shutdown signal to be received.
	<-stopCh
	// Confirm successful shutdown.
	log.WithField(constants.Version, version.Version).Infof("%s is shutting down", constants.ApplicationName)
}
