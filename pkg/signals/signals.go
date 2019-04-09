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

package signals

import (
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler registers a listener for the SIGTERM and SIGINT signals.
// A channel is returned, which is closed on one of these signals.
// If a second signal is caught, the application is terminated immediately with exit code 1.
func SetupSignalHandler() chan struct{} {
	stopCh := make(chan struct{})
	termCh := make(chan os.Signal, 2)
	// Notify termCh of SIGINT and SIGTERM.
	signal.Notify(termCh, syscall.SIGINT, syscall.SIGTERM)
	// Wait for a signal to be received.
	go func() {
		<-termCh
		// The first signal was received, so we close the channel.
		close(stopCh)
		<-termCh
		// The second signal was received, so we exit immediately.
		os.Exit(1)
	}()
	return stopCh
}
