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

package admission

import (
	"fmt"
	"strconv"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
)

// validateAndMutatePostgresqlInstance validates and mutates the provided PostgresqlInstance object.
// If the current request is a CREATE request, only currentObj is populated.
// If the current request is an UPDATE request, both currentObj and previousObj are populated.
// If the current request is a DELETE request, only previousObj is populated.
func (w *Webhook) validateAndMutatePostgresqlInstance(currentObj, previousObj *v1alpha1.PostgresqlInstance) (*v1alpha1.PostgresqlInstance, error) {
	// Check which type of request we're dealing with and act accordingly.
	switch {
	case currentObj == nil && previousObj != nil:
		// The current request is a DELETE request.
		// Hence, we must allow it if and only if the "cloudsql.travelaudience.com/allow-deletion" annotation is present on the resource and set to "true".
		if v, exists := previousObj.Annotations[constants.AllowDeletionAnnotationKey]; !exists || v != constants.AllowDeletionAnnotationValue {
			return nil, fmt.Errorf("%q cannot be deleted unless the %q annotation is set to %q", previousObj.Name, constants.AllowDeletionAnnotationKey, constants.AllowDeletionAnnotationValue)
		}
		return nil, nil
	default:
		// The current request is either a CREATE or UPDATE request.
		// Clone the current object so that we can safely mutate it if necessary.
		mutatedObj := currentObj.DeepCopy()
		// Make sure that the map of annotations is initialized on the cloned object.
		if mutatedObj.Annotations == nil {
			mutatedObj.Annotations = make(map[string]string, 1)
		}
		// Inject the "cloudsql.travelaudience.com/allow-deletion" annotation with a value of "false" if it is not present.
		if _, exists := mutatedObj.Annotations[constants.AllowDeletionAnnotationKey]; !exists {
			mutatedObj.Annotations[constants.AllowDeletionAnnotationKey] = strconv.FormatBool(false)
		}
		// Return the (possibly) mutated object so a patch can be created if necessary.
		return mutatedObj, nil
	}
}
