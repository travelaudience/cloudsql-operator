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

package constants

const (
	// ApplicationName holds the name of the application.
	ApplicationName = "cloudsql-postgres-operator"
	// DefaultCloudsqlPostgresOperatorNamespace is the namespace where cloudsql-postgres-operator is deployed by default.
	DefaultCloudsqlPostgresOperatorNamespace = "cloudsql-postgres-operator"
	// DefaultWebhookBindAddress is the address to which the webhook binds by default.
	DefaultWebhookBindAddress = "0.0.0.0:443"
)
