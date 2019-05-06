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

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
)

// GetPostgresqlInstanceCondition returns the condition of the provided type associated with the provided PostgresqlInstance resource.
func GetPostgresqlInstanceCondition(postgresqlInstance *v1alpha1api.PostgresqlInstance, conditionType v1alpha1api.PostgresqlInstanceStatusConditionType) (v1alpha1api.PostgresqlInstanceStatusCondition, error) {
	for _, cnd := range postgresqlInstance.Status.Conditions {
		if cnd.Type == conditionType {
			return cnd, nil
		}
	}
	return v1alpha1api.PostgresqlInstanceStatusCondition{}, fmt.Errorf("condition of type %q not found in postgresqlinstance %q", conditionType, postgresqlInstance.Name)
}
