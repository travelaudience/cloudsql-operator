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

package crds

import (
	"fmt"

	extsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
)

const (
	// PostgresqlInstanceKind is the value used as ".spec.names.kind" when registering the PostgresqlInstance CRD.
	PostgresqlInstanceKind = "PostgresqlInstance"
	// PostgresqlInstancePlural is the value used as ".spec.names.plural" when registering the PostgresqlInstance CRD.
	PostgresqlInstancePlural = "postgresqlinstances"
)

var (
	// postgresqlInstanceCRDName is the value used as ".metadata.name" when registering the PostgresqlInstance CRD.
	postgresqlInstanceCRDName = fmt.Sprintf("%s.%s", PostgresqlInstancePlural, v1alpha1.SchemeGroupVersion.Group)
)

var (
	// crds is a mapping between kinds and actual CustomResourceDefinition resources.
	crds = map[string]*extsv1beta1.CustomResourceDefinition{
		PostgresqlInstanceKind: {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					constants.LabelAppKey: constants.ApplicationName,
				},
				Name: postgresqlInstanceCRDName,
			},
			Spec: extsv1beta1.CustomResourceDefinitionSpec{
				Group: v1alpha1.SchemeGroupVersion.Group,
				Versions: []extsv1beta1.CustomResourceDefinitionVersion{
					{
						Name:    v1alpha1.SchemeGroupVersion.Version,
						Served:  true,
						Storage: true,
					},
				},
				Scope: extsv1beta1.ClusterScoped,
				Names: extsv1beta1.CustomResourceDefinitionNames{
					Plural: PostgresqlInstancePlural,
					Kind:   PostgresqlInstanceKind,
				},
				Subresources: &extsv1beta1.CustomResourceSubresources{
					Status: &extsv1beta1.CustomResourceSubresourceStatus{},
				},
			},
		},
	}
)
