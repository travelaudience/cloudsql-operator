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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
)

const (
	// PostgresqlInstanceMetadataNamePrefix is the prefix used when generating random values for the ".metadata.name" field of PostgresqlInstance objects.
	PostgresqlInstanceMetadataNamePrefix = "postgresqlinstance-"
	// postgresqlInstanceSpecNamePrefix is the prefix used when generating random values for the ".spec.name" field of PostgresqlInstance objects.
	postgresqlInstanceSpecNamePrefix = "cloudsql-postgres-operator-"
	// postgresqlInstanceSpecNameSuffixLength is the length of the suffix used when generating random values for the ".spec.name" field of PostgresqlInstance objects.
	postgresqlInstanceSpecNameSuffixLength = 6
)

// NewRandomPostgresqlInstanceSpecName returns a random string that can be used as the value of ".spec.name" for a PostgresqlInstance resource.
func (f *Framework) NewRandomPostgresqlInstanceSpecName() string {
	return postgresqlInstanceSpecNamePrefix + f.RandomStringWithLength(postgresqlInstanceSpecNameSuffixLength)
}

// DeletePostgresqlInstanceByName deletes the provided PostgresqlInstance resource.
func (f *Framework) DeletePostgresqlInstanceByName(metadataName string) error {
	t, err := f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Get(metadataName, metav1.GetOptions{})
	if err != nil {
		return nil
	}
	t.Annotations[constants.AllowDeletionAnnotationKey] = constants.True
	if _, err := f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Update(t); err != nil {
		return err
	}
	return f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Delete(t.Name, metav1.NewDeleteOptions(0))
}
