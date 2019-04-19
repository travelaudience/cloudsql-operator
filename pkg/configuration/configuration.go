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

package configuration

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
)

// Admission holds admission-related configuration options.
type Admission struct {
	// BindAddress is the "host:port" combination used to serve the admission webhook.
	BindAddress string `toml:"bind_address"`
}

// setDefaults sets default values where necessary.
func (a *Admission) setDefaults() {
	if a.BindAddress == "" {
		a.BindAddress = constants.DefaultWebhookBindAddress
	}
}

// Cluster holds cluster-related configuration options.
type Cluster struct {
	// Kubeconfig holds the path to the kubeconfig file to use.
	Kubeconfig string `toml:"kubeconfig"`
	// Namespace holds the namespace where cloudsql-postgres-operator is deployed.
	Namespace string `toml:"namespace"`
}

// setDefaults sets default values where necessary.
func (c *Cluster) setDefaults() {
	if c.Namespace == "" {
		c.Namespace = constants.DefaultCloudsqlPostgresOperatorNamespace
	}
}

// Configuration is the root configuration object.
type Configuration struct {
	// Admission holds admission-related configuration options.
	Admission Admission `toml:"admission"`
	// Cluster holds cluster-related configuration options.
	Cluster Cluster `toml:"cluster"`
	// Logging holds logging-related configuration options.
	Logging Logging `toml:"logging"`
	// Project holds project-related configuration options.
	Project Project `toml:"project"`
}

// SetDefaults sets default values where necessary.
func (c *Configuration) setDefaults() {
	c.Admission.setDefaults()
	c.Cluster.setDefaults()
	c.Logging.setDefaults()
	c.Project.setDefaults()
}

// Logging holds logging-related configuration options.
type Logging struct {
	// Level holds the log level to use.
	Level string `toml:"level"`
}

// setDefaults sets default values where necessary.
func (l *Logging) setDefaults() {
	if l.Level == "" {
		l.Level = log.InfoLevel.String()
	}
}

// Project holds project-related configuration options.
type Project struct {
	// ProjectID holds the name of the Google Cloud Platform project where cloudsql-postgres-operator is managing Cloud SQL instances.
	ProjectID string `toml:"project_id"`
}

// setDefaults sets default values where necessary.
func (p *Project) setDefaults() {
	// Nothing to do.
}

// MustNewConfigurationFromFile attempts to parse the specified configuration file, exiting the application if it cannot be parsed.
func MustNewConfigurationFromFile(path string) Configuration {
	if path == "" {
		log.Fatalf("the path to the configuration file must not be empty")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read the configuration file: %v", err)
	}
	var r Configuration
	if err := toml.Unmarshal(b, &r); err != nil {
		log.Fatalf("failed to read the configuration file: %v", err)
	}
	r.setDefaults()
	return r
}
