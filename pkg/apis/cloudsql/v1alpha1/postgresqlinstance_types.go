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
	"strconv"
	"strings"

	cloudsqladmin "google.golang.org/api/sqladmin/v1beta4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// aclEntryKind is the value of the ".kind" field of each element of a CSQLP instance's ".settings.ipConfiguration.authorizedNetworks" field, as returned by the Cloud SQL Admin API.
	// We explicitly set it in the "APIValue()" method in order to make comparisons easier when updating a CSQLP instance.
	aclEntryKind = "sql#aclEntry"
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

const (
	// PostgresqlInstanceStatusConditionTypeCreated indicates that the CSQLP instance represented by a given PostgresqlInstance resource has been created.
	PostgresqlInstanceStatusConditionTypeCreated = PostgresqlInstanceStatusConditionType("Created")
	// PostgresqlInstanceStatusConditionTypeReady indicates that the CSQLP instance represented by a given PostgresqlInstance resource is in a ready state.
	PostgresqlInstanceStatusConditionTypeReady = PostgresqlInstanceStatusConditionType("Ready")
	// PostgresqlInstanceStatusConditionTypeUpToDate indicates that the settings for the CSQLP instance represented by a given PostgresqlInstance resource are up-to-date.
	PostgresqlInstanceStatusConditionTypeUpToDate = PostgresqlInstanceStatusConditionType("UpToDate")
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
	Flags PostgresqlInstanceSpecFlags `json:"flags"`
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

// APIValue returns the Cloud SQL Admin API value that represents the current availability type.
func (v *PostgresqlInstanceSpecAvailabilityType) APIValue() string {
	return strings.ToUpper(string(*v))
}

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

// PostgresqlInstanceSpecFlags allows for customizing the database flags for a CSQLP instance.
type PostgresqlInstanceSpecFlags []string

// APIValue returns the Cloud SQL Admin API value that represents the current set of database flags.
func (v *PostgresqlInstanceSpecFlags) APIValue() []*cloudsqladmin.DatabaseFlags {
	if len(*v) == 0 {
		return nil
	}
	f := make([]*cloudsqladmin.DatabaseFlags, 0, len(*v))
	for _, flag := range *v {
		parts := strings.Split(flag, "=")
		if len(parts) != 2 {
			// If the current flag specifier is malformed, we skip it.
			// This should never happen in practice, as the admission webhook rejects any PostgresqlInstance resources for which this does not hold.
			continue
		}
		f = append(f, &cloudsqladmin.DatabaseFlags{
			Name:  parts[0],
			Value: parts[1],
		})
	}
	return f
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

// APIValue returns the Cloud SQL Admin API value that represents the current choice of zone.
func (v *PostgresqlInstanceSpecLocationZone) APIValue() string {
	switch *v {
	case PostgresqlInstanceSpecLocationZoneAny:
		return ""
	default:
		return string(*v)
	}
}

// PostgresqlInstanceSpecMaintenance allows for customizing the maintenance window of a CSQLP instance.
type PostgresqlInstanceSpecMaintenance struct {
	// Day is the preferred day of the week for periodic maintenance of the CSQLP instance.
	// +optional
	Day *PostgresqlInstanceSpecMaintenanceDay `json:"day"`
	// Hour is the preferred hour of the day (in UTC) for periodic maintenance of the CSQLP instance, in 24-hour format.
	// +optional
	Hour *PostgresqlInstanceSpecMaintenanceHour `json:"hour"`
}

// PostgresqlInstanceSpecMaintenanceDay represents a day of the week for periodic maintenance of a CSQLP instance.
type PostgresqlInstanceSpecMaintenanceDay string

// APIValue returns the Cloud SQL Admin API value that represents the current maintenance day.
func (v *PostgresqlInstanceSpecMaintenanceDay) APIValue() int64 {
	switch *v {
	case PostgresqlInstanceSpecMaintenanceDayMonday:
		return 1
	case PostgresqlInstanceSpecMaintenanceDayTuesday:
		return 2
	case PostgresqlInstanceSpecMaintenanceDayWednesday:
		return 3
	case PostgresqlInstanceSpecMaintenanceDayThursday:
		return 4
	case PostgresqlInstanceSpecMaintenanceDayFriday:
		return 5
	case PostgresqlInstanceSpecMaintenanceDaySaturday:
		return 6
	case PostgresqlInstanceSpecMaintenanceDaySunday:
		fallthrough
	default:
		return 7
	}
}

// PostgresqlInstanceSpecMaintenanceHour represents a hour of the day for periodic maintenance of a CSQLP instance.
type PostgresqlInstanceSpecMaintenanceHour string

// APIValue returns the Cloud SQL Admin API value that represents the current maintenance hour.
func (v *PostgresqlInstanceSpecMaintenanceHour) APIValue() int64 {
	i, err := strconv.ParseInt(strings.TrimSuffix(string(*v), ":00"), 10, 64)
	if err != nil {
		return 0
	}
	return i
}

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
	AuthorizedNetworks PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetworkList `json:"authorizedNetworks"`
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

// PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetworkList allows for specifying a list of subnets for which to authorize access to a CSQLP instance via a public IP.
type PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetworkList []PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetwork

// APIValue returns the Cloud SQL Admin API value that represents the current list of authorized networks.
func (v *PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetworkList) APIValue() []*cloudsqladmin.AclEntry {
	r := make([]*cloudsqladmin.AclEntry, 0, len(*v))
	for _, an := range *v {
		ae := &cloudsqladmin.AclEntry{
			Kind:  aclEntryKind,
			Value: an.Cidr,
		}
		if an.Name != nil {
			ae.Name = *an.Name
		}
		r = append(r, ae)
	}
	return r
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

// APIValue returns the Cloud SQL Admin API value that represents the current disk type.
func (v *PostgresqlInstanceSpecResourcesDiskType) APIValue() string {
	switch *v {
	case PostgresqlInstanceSpecResourceDiskTypeHDD:
		return "PD_HDD"
	default:
		return "PD_SSD"
	}
}

// PostgresqlInstanceSpecVersion represents a supported Cloud SQL for PostgreSQL version.
type PostgresqlInstanceSpecVersion string

// APIValue returns the Cloud SQL Admin API value that represents the current Cloud SQL for PostgreSQL version.
func (v *PostgresqlInstanceSpecVersion) APIValue() string {
	switch *v {
	case PostgresqlInstanceSpecVersion96:
		fallthrough
	default:
		return "POSTGRES_9_6"
	}
}

// PostgresqlInstanceStatus represents the status of a CSQLP instance.
type PostgresqlInstanceStatus struct {
	// Conditions is the set of conditions associated with the current PostgresqlInstance resource.
	// +optional
	Conditions []PostgresqlInstanceStatusCondition `json:"conditions,omitempty"`
	// IPs is the set of IPs associated with the current PostgresqlInstance resource.
	// +optional
	IPs PostgresqlInstanceStatusIPAddresses `json:"ips,omitempty"`
}

// PostgresqlInstanceStatusCondition represents a condition associated with a PostgresqlInstance resource.
type PostgresqlInstanceStatusCondition struct {
	// LastTransitionTime is the timestamp corresponding to the last status change of this condition.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Message is a human readable description of the details of the condition's last transition.
	// +optional
	Message string `json:"message,omitempty"`
	// Reason is a brief machine readable explanation for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Status is the status of the condition (one of "True", "False" or "Unknown").
	Status corev1.ConditionStatus `json:"status"`
	// Type is the type of the condition.
	Type PostgresqlInstanceStatusConditionType `json:"type"`
}

// PostgresqlInstanceStatusConditionType represents the type of a condition associated with a PostgresqlInstance resource.
type PostgresqlInstanceStatusConditionType string

// PostgresqlInstanceStatusIPAddresses holds the reported IP addresses of a CSQLP instance.
type PostgresqlInstanceStatusIPAddresses struct {
	// PrivateIP is the private IP associated with the CSQLP instance (if any).
	PrivateIP string `json:"privateIp,omitempty"`
	// PublicIP is the public IP associated with the CSQLP instance (if any).
	PublicIP string `json:"publicIp,omitempty"`
}
