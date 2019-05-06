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
	externalip "github.com/glendc/go-external-ip"
)

// determineExternalIP attempts to determine the external IP of the host where the end-to-end test suite is running on.
func determineExternalIP() (string, error) {
	// Create a new consensus with the default configuration.
	c := externalip.NewConsensus(externalip.DefaultConsensusConfig(), nil)
	// Add sources known to correctly return an IPV4
	if err := c.AddVoter(externalip.NewHTTPSource("https://ipinfo.io/ip"), 1); err != nil {
		return "", err
	}
	if err := c.AddVoter(externalip.NewHTTPSource("https://myexternalip.com/raw"), 1); err != nil {
		return "", err
	}
	if err := c.AddVoter(externalip.NewHTTPSource("https://checkip.amazonaws.com"), 1); err != nil {
		return "", err
	}
	// Try to determine the external IP.
	i, err := c.ExternalIP()
	if err != nil {
		return "", err
	}
	return i.String(), nil
}
