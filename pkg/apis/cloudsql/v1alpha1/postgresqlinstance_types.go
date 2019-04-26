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

const (
	// Any represents an arbitrary choice of a value.
	Any = "Any"
	// False represents the "false" boolean value.
	False = "false"
	// True represents the "true" boolean value.
	True = "true"
)

const (
	// PostgresqlInstanceSpecAvailabilityTypeRegional represents the "REGIONAL" availability type for CSQLP instances.
	PostgresqlInstanceSpecAvailabilityTypeRegional = PostgresqlInstanceSpecAvailabilityType("Regional")
	// PostgresqlInstanceSpecAvailabilityTypeZonal represents the "ZONAL" availability type for CSQLP instances.
	PostgresqlInstanceSpecAvailabilityTypeZonal = PostgresqlInstanceSpecAvailabilityType("Zonal")
)

const (
	// PostgresqlInstanceSpecLocationZoneAny represents an arbitrary choice of a zone for a CSQLP instance.
	PostgresqlInstanceSpecLocationZoneAny = PostgresqlInstanceSpecLocationZone(Any)
)

const (
	// PostgresqlInstanceSpecMaintenanceDayAny represents an arbitrary choice of a day of the week for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDayAny = PostgresqlInstanceSpecMaintenanceDay(Any)
	// PostgresqlInstanceSpecMaintenanceDayMonday represents the choice of Mondays for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDayMonday = PostgresqlInstanceSpecMaintenanceDay("Monday")
	// PostgresqlInstanceSpecMaintenanceDayTuesday represents the choice of Tuesdays for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDayTuesday = PostgresqlInstanceSpecMaintenanceDay("Tuesday")
	// PostgresqlInstanceSpecMaintenanceDayWednesday represents the choice of Wednesdays for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDayWednesday = PostgresqlInstanceSpecMaintenanceDay("Wednesday")
	// PostgresqlInstanceSpecMaintenanceDayThursday represents the choice of Thursdays for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDayThursday = PostgresqlInstanceSpecMaintenanceDay("Thursday")
	// PostgresqlInstanceSpecMaintenanceDayFriday represents the choice of Fridays for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDayFriday = PostgresqlInstanceSpecMaintenanceDay("Friday")
	// PostgresqlInstanceSpecMaintenanceDaySaturday represents the choice of Saturdays for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDaySaturday = PostgresqlInstanceSpecMaintenanceDay("Saturday")
	// PostgresqlInstanceSpecMaintenanceDaySunday represents the choice of Sundays for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceDaySunday = PostgresqlInstanceSpecMaintenanceDay("Sunday")
)

const (
	// PostgresqlInstanceSpecMaintenanceHourAny represents an arbitrary choice of an hour of the day for periodic maintenance of a CSQLP instance.
	PostgresqlInstanceSpecMaintenanceHourAny = Any
)

const (
	// PostgresqlInstanceSpecResourceDiskTypeHDD represents the "HDD" disk type for CSQLP instances.
	PostgresqlInstanceSpecResourceDiskTypeHDD = PostgresqlInstanceSpecResourcesDiskType("HDD")
	// PostgresqlInstanceSpecResourceDiskTypeSSD represents the "SSD" disk type for CSQLP instances.
	PostgresqlInstanceSpecResourceDiskTypeSSD = PostgresqlInstanceSpecResourcesDiskType("SSD")
)

const (
	// PostgresqlInstanceSpecVersion96 represents the "POSTGRES_9_6" version of CSQLP instances.
	PostgresqlInstanceSpecVersion96 = PostgresqlInstanceSpecVersion("9.6")
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

// PostgresqlInstanceSpec represents the specification of a CSQLP instance.
type PostgresqlInstanceSpec struct {
	// Availability allows for customizing the availability of the CSQLP instance.
	// +optional
	Availability *PostgresqlInstanceSpecAvailability `json:"availability"`
	// Backups allows for customizing the backup strategy for the CSQLP instance.
	// +optional
	Backups *PostgresqlInstanceSpecBackups `json:"backups"`
	// Flags is a list of flags passed to the CSQLP instance.
	// +optional
	Flags []string `json:"flags"`
	// Labels is a map of user-defined labels to be set on the CSQLP instance.
	// +optional
	Labels map[string]string `json:"labels"`
	// Location allows for customizing the geographical location of the CSQLP instance.
	// +optional
	Location *PostgresqlInstanceSpecLocation `json:"location"`
	// Maintenance allows for customizing the maintenance window of the CSQLP instance.
	// +optional
	Maintenance *PostgresqlInstanceSpecMaintenance `json:"maintenance"`
	// Name is the name of the CSQLP instance.
	Name string `json:"name"`
	// Networking allows for customizing the networking aspects of the CSQLP instance.
	// +optional
	Networking *PostgresqlInstanceSpecNetworking `json:"networking"`
	// Paused indicates whether reconciliation of the CSQLP instance is paused.
	// Meant only to facilitate end-to-end testing.
	// TODO Find a way to not leak this testing implementation detail into the API.
	Paused bool `json:"paused,omitempty"`
	// Resources allows for customizing the resource requests for the CSQLP instance.
	// +optional
	Resources *PostgresqlInstanceSpecResources `json:"resources"`
	// Version is the version of the CSQLP instance.
	// +optional
	Version *PostgresqlInstanceSpecVersion `json:"version"`
}

// PostgresqlInstanceSpecAvailability allows for customizing the availability of a CSQLP instance.
type PostgresqlInstanceSpecAvailability struct {
	// Type is the availability type of the CSQLP instance.
	Type *PostgresqlInstanceSpecAvailabilityType `json:"type"`
}

// PostgresqlInstanceSpecAvailabilityType represents availability types for CSQLP instances.
type PostgresqlInstanceSpecAvailabilityType string

// PostgresqlInstanceSpecBackups allows for customizing the backup strategy for a CSQLP instance.
type PostgresqlInstanceSpecBackups struct {
	// Daily allows for customizing the daily backup strategy for the CSQLP instance.
	// +optional
	Daily *PostgresqlInstancSpecBackupsDaily `json:"daily"`
}

// PostgresqlInstancSpecBackupsDaily allows for customizing the daily backup strategy for a CSQLP instance.
type PostgresqlInstancSpecBackupsDaily struct {
	// Enabled specifies whether daily backups are enabled for the CSQLP instance.
	// +optional
	Enabled *bool `json:"enabled"`
	// StartTime is the start time (in UTC) for the daily backups of the CSQLP instance, in 24-hour format.
	// +optional
	StartTime *string `json:"startTime"`
}

// PostgresqlInstanceSpecLocation allows for customizing the geographical location of a CSQLP instance.
type PostgresqlInstanceSpecLocation struct {
	// Region is the region where the CSQLP instance is located.
	// +optional
	Region *string `json:"region"`
	// Zone is the zone where the CSQLP instance is located.
	// +optional
	Zone *PostgresqlInstanceSpecLocationZone `json:"zone"`
}

// PostgresqlInstanceSpecLocationZone represents a Google Cloud Platform zone where CSQLP instances may be located.
type PostgresqlInstanceSpecLocationZone string

// PostgresqlInstanceSpecMaintenance allows for customizing the maintenance window of a CSQLP instance.
type PostgresqlInstanceSpecMaintenance struct {
	// Day is the preferred day of the week for periodic maintenance of the CSQLP instance.
	// +optional
	Day *PostgresqlInstanceSpecMaintenanceDay `json:"day"`
	// Hour is the preferred hour of the day (in UTC) for periodic maintenance of the CSQLP instance, in 24-hour format.
	// +optional
	Hour *string `json:"hour"`
}

// PostgresqlInstanceSpecMaintenanceDay represents a day of the week for periodic maintenance of a CSQLP instance.
type PostgresqlInstanceSpecMaintenanceDay string

// PostgresqlInstanceSpecNetworking allows for customizing the networking aspects of a CSQLP instance.
type PostgresqlInstanceSpecNetworking struct {
	// PrivateIP allows for customizing access to the CSQLP instance via a private IP address.
	// +optional
	PrivateIP *PostgresqlInstanceSpecNetworkingPrivateIP `json:"privateIp"`
	// PublicIP allows for customizing access to the CSQLP instance via a public IP address.
	// +optional
	PublicIP *PostgresqlInstanceSpecNetworkingPublicIP `json:"publicIp"`
}

// PostgresqlInstanceSpecNetworkingPrivateIP allows for customizing access to a CSQLP instance via a private IP.
type PostgresqlInstanceSpecNetworkingPrivateIP struct {
	// Enabled specifies whether the Cloud SQL for Postgresql Instance is accessible via a private IP address.
	// +optional
	Enabled *bool `json:"enabled"`
	// Network is resource link of the VPC network from which the CSQLP instance is accessible via a private IP address.
	// +optional
	Network *string `json:"network"`
}

// PostgresqlInstanceSpecNetworkingPublicIP allows for customizing access to a CSQLP instance via a public IP.
type PostgresqlInstanceSpecNetworkingPublicIP struct {
	// AuthorizedNetworks is a list of rules that authorize access to the CSQLP instance via a public IP address.
	// +optional
	AuthorizedNetworks []PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetwork `json:"authorizedNetworks"`
	// Enabled specifies whether the Cloud SQL for Postgresql Instance is accessible via a public IP address.
	// +optional
	Enabled *bool `json:"enabled"`
}

// PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetwork allows for specifying a subnet for which to authorize access to a CSQLP instance via a public IP.
type PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetwork struct {
	// Cidr is the subnet which to authorize by the current rule.
	Cidr string `json:"cidr"`
	// Name is the name of the current rule.
	// +optional
	Name *string `json:"name"`
}

// PostgresqlInstanceSpecResources allows for customizing the resource requests for a CSQLP instance.
type PostgresqlInstanceSpecResources struct {
	// Disk allows for customizing the storage of the CSQLP instance.
	// +optional
	Disk *PostgresqlInstanceSpecResourcesDisk `json:"disk"`
	// InstanceType is the instance type to use for the CSQLP instance.
	// +optional
	InstanceType *string `json:"instanceType"`
}

// PostgresqlInstanceSpecResourcesDisk allows for customizing the storage of the CSQLP instance.
type PostgresqlInstanceSpecResourcesDisk struct {
	// SizeMaximumGb is the maximum size (in GB) to which the storage capacity of the CSQLP instance can be automatically increased.
	// +optional
	SizeMaximumGb *int32 `json:"sizeMaximumGb"`
	// SizeMinimumGb is the minimum size (in GB) requested for the storage capacity of the CSQLP instance.
	// +optional
	SizeMinimumGb *int32 `json:"sizeMinimumGb"`
	// Type is the type of disk used for storage by the CSQLP instance.
	// +optional
	Type *PostgresqlInstanceSpecResourcesDiskType `json:"type"`
}

// PostgresqlInstanceSpecResourcesDiskType represents the types of disk that can be used for storage by the Cloud SQL for Postgresql instance.
type PostgresqlInstanceSpecResourcesDiskType string

// PostgresqlInstanceSpecVersion represents a supported Cloud SQL for PostgreSQL version.
type PostgresqlInstanceSpecVersion string

// PostgresqlInstanceStatus represents the status of a CSQLP instance.
type PostgresqlInstanceStatus struct {
}
