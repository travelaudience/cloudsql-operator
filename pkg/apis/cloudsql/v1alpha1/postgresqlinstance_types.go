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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresqlInstance represents a CSQLP instance.
type PostgresqlInstance struct {
	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec represents the specification of the CSQLP instance.
	Spec PostgresqlInstanceSpec `json:"spec"`
	// Status represents the status of the CSQLP instance.
	Status PostgresqlInstanceStatus `json:"status"`
}

// PostgresqlInstanceSpec represents the specification of a CSQLP instance.
type PostgresqlInstanceSpec struct {
}

// PostgresqlInstanceStatus represents the status of a CSQLP instance.
type PostgresqlInstanceStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresqlInstanceList is a list of PostgresqlInstance resources.
type PostgresqlInstanceList struct {
	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`
	// Items is the set of PostgresqlInstance resources in the list.
	Items []PostgresqlInstance `json:"items"`
}
