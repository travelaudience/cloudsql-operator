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
	// annotationKeyPrefix is the prefix used in all annotations supported by cloudsql-postgres-operator.
	annotationKeyPrefix = "cloudsql.travelaudience.com/"
)

const (
	// AllowDeletionAnnotationKey is the key of the annotation that specifies whether deletion of a given resource is allowed.
	AllowDeletionAnnotationKey = annotationKeyPrefix + "allow-deletion"
)
