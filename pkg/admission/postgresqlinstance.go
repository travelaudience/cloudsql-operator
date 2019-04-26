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
	"regexp"
	"strings"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	googleutil "github.com/travelaudience/cloudsql-postgres-operator/pkg/util/google"
)

const (
	// postgresqlInstanceSpecResourcesDiskSizeMinimumGbLowerBound is the lower bound on the value of the ".spec.resources.disk.sizeMinimumGb" field of a PostgresqlInstance resource.
	postgresqlInstanceSpecResourcesDiskSizeMinimumGbLowerBound = 10
	// postgresqlInstanceSpecFlagsSeparator is the separator that must be used between "<name>" and "<value>" in each item present in the ".spec.flags" field of a PostgresqlInstance resource.
	postgresqlInstanceSpecFlagsSeparator = "="
	// postgresqlInstanceSpecNameProjectIDMaxLength is the maximum length that the ".spec.name" field of a PostgresqlInstance resource may have when concatenated with the Google Cloud Platform project's ID.
	postgresqlInstanceSpecNameProjectIDMaxLength = 97
	// PostgresqlInstanceSpecLabelsOwnerName is the name of the "owner" label set injected on the ".spec.labels" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecLabelsOwnerName = "owner"
)

var (
	// hourOfTheDayRegex is the regular expression used to match hours of the day in 24-hour format.
	hourOfTheDayRegex = regexp.MustCompile(`^([01][0-9]|2[03]):00$`)
	// postgresqlInstanceSpecNameRegex is the regular expression used to validate the ".spec.name" field of a PostgresqlInstance resource.
	postgresqlInstanceSpecNameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]+[a-z0-9]$`)
)

var (
	// PostgresqlInstanceSpecAvailabilityTypeDefault is the default value for the ".spec.availability.type" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecAvailabilityTypeDefault = v1alpha1.PostgresqlInstanceSpecAvailabilityTypeZonal
	// PostgresqlInstanceSpecBackupsDailyEnabledDefault is the default value for the ".spec.backups.daily.enabled" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecBackupsDailyEnabledDefault = true
	// PostgresqlInstanceSpecBackupsDailyStartTimeDefault is the default value for the ".spec.backups.daily.startTime" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecBackupsDailyStartTimeDefault = "00:00"
	// PostgresqlInstanceSpecLocationRegionDefault is the default value for the ".spec.location.region" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecLocationRegionDefault = "europe-west1"
	// PostgresqlInstanceSpecLocationZoneDefault is the default value for the ".spec.location.zone" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecLocationZoneDefault = v1alpha1.PostgresqlInstanceSpecLocationZoneAny
	// PostgresqlInstanceSpecMaintenanceDayDefault is the default value for the ".spec.maintenance.day" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecMaintenanceDayDefault = v1alpha1.PostgresqlInstanceSpecMaintenanceDayAny
	// PostgresqlInstanceSpecMaintenanceHourDefault is the default value for the ".spec.maintenance.hour" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecMaintenanceHourDefault = "04:00"
	// PostgresqlInstanceSpecNetworkingPrivateIPEnabledDefault is the default value for the ".spec.networking.privateIp.enabled" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecNetworkingPrivateIPEnabledDefault = false
	// PostgresqlInstanceSpecNetworkingPrivateIPNetworkDefault is the default value for the ".spec.networking.privateIp.network" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecNetworkingPrivateIPNetworkDefault = ""
	// PostgresqlInstanceSpecNetworkingPublicIPEnabledDefault is the default value for the ".spec.networking.publicIp.enabled" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecNetworkingPublicIPEnabledDefault = false
	// PostgresqlInstanceSpecResourcesDiskSizeMaximumGbDefault is the default value for the ".spec.resources.disk.sizeMaximumGb" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecResourcesDiskSizeMaximumGbDefault = int32(0)
	// PostgresqlInstanceSpecResourcesDiskSizeMinimumGbDefault is the default value for the ".spec.resources.disk.sizeMinimumGb" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecResourcesDiskSizeMinimumGbDefault = int32(10)
	// PostgresqlInstanceSpecResourcesDiskTypeDefault is the default value for the ".spec.resources.disk.type" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecResourcesDiskTypeDefault = v1alpha1.PostgresqlInstanceSpecResourceDiskTypeSSD
	// PostgresqlInstanceSpecResourcesInstanceTypeDefault is the default value for the ".spec.resources.instanceType" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecResourcesInstanceTypeDefault = "db-custom-1-3840"
	// PostgresqlInstanceSpecVersionDefault is the default value for the ".spec.version" field of a PostgresqlInstance resource.
	PostgresqlInstanceSpecVersionDefault = v1alpha1.PostgresqlInstanceSpecVersion96
)

// postgresqlInstanceWebhookOperation represents a validation/mutation operation performed by the admission webhook on PostgresqlInstance resources.
type postgresqlInstanceWebhookOperation func(mutatedObj, previousObj *v1alpha1.PostgresqlInstance) error

// validateAndMutatePostgresqlInstance validates and mutates the provided PostgresqlInstance object.
// If the current request is a CREATE request, only currentObj is populated.
// If the current request is an UPDATE request, both currentObj and previousObj are populated.
// If the current request is a DELETE request, only previousObj is populated.
func (w *Webhook) validateAndMutatePostgresqlInstance(currentObj, previousObj *v1alpha1.PostgresqlInstance) (*v1alpha1.PostgresqlInstance, error) {
	// Check whether the current request is a DELETE request and act accordingly.
	// In this case, we allow the request if and only if the "cloudsql.travelaudience.com/allow-deletion" annotation is present on the resource and set to "true".
	if currentObj == nil && previousObj != nil {
		if v, exists := previousObj.Annotations[constants.AllowDeletionAnnotationKey]; !exists || v != v1alpha1.True {
			return nil, fmt.Errorf("the resource cannot be deleted unless the %q annotation is set to %q", constants.AllowDeletionAnnotationKey, v1alpha1.True)
		}
		return nil, nil
	}

	// At this point we know the current request is either a CREATE or UPDATE request.

	// Clone the current object so that we can safely mutate it if necessary.
	mutatedObj := currentObj.DeepCopy()

	// Perform the required validation/mutation steps.
	for _, fn := range []postgresqlInstanceWebhookOperation{
		mutatePostgresqlInstanceMetadataAnnotations,
		validateAndMutatePostgresqlInstanceSpecAvailability,
		validateAndMutatePostgresqlInstanceSpecDailyBackups,
		validateAndMutatePostgresqlInstanceSpecFlags,
		validateAndMutatePostgresqlInstanceSpecLabels,
		validateAndMutatePostgresqlInstanceSpecLocation,
		validateAndMutatePostgresqlInstanceSpecMaintenance,
		w.validatePostgresqlInstanceSpecName,
		w.validateAndMutatePostgresqlInstanceSpecNetworking,
		validateAndMutatePostgresqlInstanceSpecResources,
		validateAndMutatePostgresqlInstanceSpecVersion,
	} {
		if err := fn(mutatedObj, previousObj); err != nil {
			return nil, err
		}
	}

	// Return the (possibly) mutated object so a patch can be created if necessary.
	return mutatedObj, nil
}

// mutatePostgresqlInstanceMetadataAnnotations injects annotations on the specified PostgresqlInstance resource.
func mutatePostgresqlInstanceMetadataAnnotations(mutatedObj, _ *v1alpha1.PostgresqlInstance) error {
	// Make sure that the map of annotations is initialized on the cloned object.
	if mutatedObj.Annotations == nil {
		mutatedObj.Annotations = make(map[string]string, 1)
	}
	// Inject the "cloudsql.travelaudience.com/allow-deletion" annotation with a value of "false" if the annotation is not present or is empty.
	if v, exists := mutatedObj.Annotations[constants.AllowDeletionAnnotationKey]; !exists || v == "" {
		mutatedObj.Annotations[constants.AllowDeletionAnnotationKey] = v1alpha1.False
	}
	return nil
}

// validateAndMutatePostgresqlInstanceSpecAvailability validates and mutates the value of ".spec.availability.type".
func validateAndMutatePostgresqlInstanceSpecAvailability(mutatedObj, _ *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.availability" is initialized.
	if mutatedObj.Spec.Availability == nil {
		mutatedObj.Spec.Availability = &v1alpha1.PostgresqlInstanceSpecAvailability{}
	}
	// If no value for ".spec.availability.type" has been provided, use the default one.
	if mutatedObj.Spec.Availability.Type == nil {
		mutatedObj.Spec.Availability.Type = &PostgresqlInstanceSpecAvailabilityTypeDefault
	}
	// Make sure that ".spec.availability.type" contains a valid value.
	switch *mutatedObj.Spec.Availability.Type {
	case v1alpha1.PostgresqlInstanceSpecAvailabilityTypeRegional, v1alpha1.PostgresqlInstanceSpecAvailabilityTypeZonal:
		// The value is valid.
	default:
		return fmt.Errorf("the availability type of the instance must be one of %q or %q (got %q)", v1alpha1.PostgresqlInstanceSpecAvailabilityTypeRegional, v1alpha1.PostgresqlInstanceSpecAvailabilityTypeZonal, *mutatedObj.Spec.Availability.Type)
	}
	return nil
}

// validateAndMutatePostgresqlInstanceSpecDailyBackups validates and mutates the value of ".spec.backups.daily".
func validateAndMutatePostgresqlInstanceSpecDailyBackups(mutatedObj, _ *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.backups" is initialized.
	if mutatedObj.Spec.Backups == nil {
		mutatedObj.Spec.Backups = &v1alpha1.PostgresqlInstanceSpecBackups{}
	}
	// Make sure that ".spec.backups.daily" is initialized.
	if mutatedObj.Spec.Backups.Daily == nil {
		mutatedObj.Spec.Backups.Daily = &v1alpha1.PostgresqlInstancSpecBackupsDaily{}
	}
	// If no value for ".spec.backups.daily.enabled" has been provided, use the default one.
	if mutatedObj.Spec.Backups.Daily.Enabled == nil {
		mutatedObj.Spec.Backups.Daily.Enabled = &PostgresqlInstanceSpecBackupsDailyEnabledDefault
	}
	// If no value for ".spec.backups.daily.startTime" has been provided, use the default one.
	if mutatedObj.Spec.Backups.Daily.StartTime == nil {
		mutatedObj.Spec.Backups.Daily.StartTime = &PostgresqlInstanceSpecBackupsDailyStartTimeDefault
	}
	// Make sure that the value of ".spec.backups.daily.startTime" matches the required format.
	if !hourOfTheDayRegex.MatchString(*mutatedObj.Spec.Backups.Daily.StartTime) {
		return fmt.Errorf("the start time for daily backups of the instance must be a valid hour of the day in 24-hour format (got %q)", *mutatedObj.Spec.Backups.Daily.StartTime)
	}
	return nil
}

// validateAndMutateInstanceSpecFlags validates and mutates the value of ".spec.flags".
func validateAndMutatePostgresqlInstanceSpecFlags(mutatedObj, _ *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.flags" is initialized.
	if mutatedObj.Spec.Flags == nil {
		mutatedObj.Spec.Flags = make([]string, 0)
	}
	// Iterate over the list of flags and validate the format of each item.
	for _, flag := range mutatedObj.Spec.Flags {
		parts := strings.Split(flag, postgresqlInstanceSpecFlagsSeparator)
		if len(parts) != 2 {
			return fmt.Errorf("flags must be specified in the \"<name>=<value>\" format (got %q)", flag)
		}
	}
	return nil
}

// validateAndMutatePostgresqlInstanceSpecLabels validates and mutates the value of ".spec.labels".
func validateAndMutatePostgresqlInstanceSpecLabels(mutatedObj, _ *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.labels" is initialized.
	if mutatedObj.Spec.Labels == nil {
		mutatedObj.Spec.Labels = make(map[string]string, 1)
	}
	// Make sure that the "owner" label is set to "cloudsql-postgres-operator".
	mutatedObj.Spec.Labels[PostgresqlInstanceSpecLabelsOwnerName] = constants.ApplicationName
	return nil
}

// validateAndMutatePostgresqlInstanceSpecLocation validates and mutates the value of ".spec.location".
func validateAndMutatePostgresqlInstanceSpecLocation(mutatedObj, previousObj *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.location" is initialized.
	if mutatedObj.Spec.Location == nil {
		mutatedObj.Spec.Location = &v1alpha1.PostgresqlInstanceSpecLocation{}
	}
	// If no value for ".spec.location.region" has been provided, use the default one.
	if mutatedObj.Spec.Location.Region == nil {
		mutatedObj.Spec.Location.Region = &PostgresqlInstanceSpecLocationRegionDefault
	}
	// If no value for ".spec.location.zone" has been provided, use the default one.
	if mutatedObj.Spec.Location.Zone == nil {
		mutatedObj.Spec.Location.Zone = &PostgresqlInstanceSpecLocationZoneDefault
	}
	// If the current request is an UPDATE request, make sure that ".spec.location.region" is not being changed/removed.
	if previousObj != nil && previousObj.Spec.Location != nil && *mutatedObj.Spec.Location.Region != *previousObj.Spec.Location.Region {
		return fmt.Errorf("the region where the instance is located cannot be changed (had %q, got %q)", *previousObj.Spec.Location.Region, *mutatedObj.Spec.Location.Region)
	}
	return nil
}

// validateAndMutatePostgresqlInstanceSpecMaintenance validates and mutates the value of ".spec.maintenance".
func validateAndMutatePostgresqlInstanceSpecMaintenance(mutatedObj, _ *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.maintenance" is initialized.
	if mutatedObj.Spec.Maintenance == nil {
		mutatedObj.Spec.Maintenance = &v1alpha1.PostgresqlInstanceSpecMaintenance{}
	}
	// If no value for ".spec.maintenance.day" has been provided, use the default one.
	if mutatedObj.Spec.Maintenance.Day == nil {
		mutatedObj.Spec.Maintenance.Day = &PostgresqlInstanceSpecMaintenanceDayDefault
	}
	// If no value for ".spec.maintenance.hour" has been provided, use the default one.
	if mutatedObj.Spec.Maintenance.Hour == nil {
		mutatedObj.Spec.Maintenance.Hour = &PostgresqlInstanceSpecMaintenanceHourDefault
	}
	// Make sure that ".spec.maintenance.day" contains a valid value.
	switch *mutatedObj.Spec.Maintenance.Day {
	case v1alpha1.PostgresqlInstanceSpecMaintenanceDayAny, v1alpha1.PostgresqlInstanceSpecMaintenanceDayMonday, v1alpha1.PostgresqlInstanceSpecMaintenanceDayTuesday, v1alpha1.PostgresqlInstanceSpecMaintenanceDayWednesday, v1alpha1.PostgresqlInstanceSpecMaintenanceDayThursday, v1alpha1.PostgresqlInstanceSpecMaintenanceDayFriday, v1alpha1.PostgresqlInstanceSpecMaintenanceDaySaturday, v1alpha1.PostgresqlInstanceSpecMaintenanceDaySunday:
		// Nothing to do.
	default:
		return fmt.Errorf("the day of the week for periodic maintenance must be %q or a valid weekday (got %q)", v1alpha1.PostgresqlInstanceSpecMaintenanceDayAny, *mutatedObj.Spec.Maintenance.Day)
	}
	// Make sure that ".spec.maintenance.hour" contains a valid value.
	switch {
	case *mutatedObj.Spec.Maintenance.Hour == v1alpha1.PostgresqlInstanceSpecMaintenanceHourAny:
		// The value is valid.
	case hourOfTheDayRegex.MatchString(*mutatedObj.Spec.Maintenance.Hour):
		// The value is valid.
	default:
		return fmt.Errorf("the hour of the day for periodic maintenance must be %q or a valid hour of the day in 24-hour format (got %q)", v1alpha1.PostgresqlInstanceSpecMaintenanceDayAny, *mutatedObj.Spec.Maintenance.Hour)
	}
	return nil
}

// validatePostgresqlInstanceSpecName validates the value of ".spec.name".
func (w *Webhook) validatePostgresqlInstanceSpecName(mutatedObj, previousObj *v1alpha1.PostgresqlInstance) error {
	// If the current request is an UPDATE request, make sure that ".spec.name" is not being changed/removed.
	if previousObj != nil && mutatedObj.Spec.Name != previousObj.Spec.Name {
		return fmt.Errorf("the name of the instance cannot be changed (had %q, got %q)", previousObj.Spec.Name, mutatedObj.Spec.Name)
	}
	// Make sure that ".spec.name" is not empty.
	if mutatedObj.Spec.Name == "" {
		return fmt.Errorf("the name of the instance cannot be empty")
	}
	// Make sure that ".spec.name" matches the required format.
	if !postgresqlInstanceSpecNameRegex.MatchString(mutatedObj.Spec.Name) {
		return fmt.Errorf("the name of the instance must match the %q regular expression (got %q)", postgresqlInstanceSpecNameRegex, mutatedObj.Spec.Name)
	}
	// Make sure that ".spec.name" does not exceed the maximum length.
	if len(mutatedObj.Spec.Name)+len(w.projectID) > postgresqlInstanceSpecNameProjectIDMaxLength {
		return fmt.Errorf("the name of the instance must not exceed %d characters (got %q)", postgresqlInstanceSpecNameProjectIDMaxLength-len(w.projectID), mutatedObj.Spec.Name)
	}
	// If the current request is a CREATE request, make sure that ".spec.name" does not clash with the name of a pre-existing CSQLP instance.
	if previousObj == nil {
		_, err := w.cloudsqlClient.Instances.Get(w.projectID, mutatedObj.Spec.Name).Do()
		if err == nil {
			// No error has been returned, which means that ".spec.name" is already being used.
			return fmt.Errorf("the name %q is already in use by an instance", mutatedObj.Spec.Name)
		}
		if !googleutil.IsNotFound(err) {
			// An error has been returned, but it is not a "404 NOT FOUND" one.
			return fmt.Errorf("failed to check whether %q can be used as an instance name: %v", mutatedObj.Spec.Name, err)
		}
		return nil
	}
	return nil
}

// validateAndMutatePostgresqlInstanceSpecNetworking validates and mutates the value of ".spec.networking".
func (w *Webhook) validateAndMutatePostgresqlInstanceSpecNetworking(mutatedObj, previousObj *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.networking" is initialized.
	if mutatedObj.Spec.Networking == nil {
		mutatedObj.Spec.Networking = &v1alpha1.PostgresqlInstanceSpecNetworking{}
	}
	// Make sure that ".spec.networking.privateIp" is initialized.
	if mutatedObj.Spec.Networking.PrivateIP == nil {
		mutatedObj.Spec.Networking.PrivateIP = &v1alpha1.PostgresqlInstanceSpecNetworkingPrivateIP{}
	}
	// If no value for ".spec.networking.privateIp.enabled" has been provided, use the default one.
	if mutatedObj.Spec.Networking.PrivateIP.Enabled == nil {
		mutatedObj.Spec.Networking.PrivateIP.Enabled = &PostgresqlInstanceSpecNetworkingPrivateIPEnabledDefault
	}
	// If no value for ".spec.networking.privateIp.network" has been provided, use the default one.
	if mutatedObj.Spec.Networking.PrivateIP.Network == nil {
		mutatedObj.Spec.Networking.PrivateIP.Network = &PostgresqlInstanceSpecNetworkingPrivateIPNetworkDefault
	}
	// Make sure that ".spec.networking.publicIp" is initialized.
	if mutatedObj.Spec.Networking.PublicIP == nil {
		mutatedObj.Spec.Networking.PublicIP = &v1alpha1.PostgresqlInstanceSpecNetworkingPublicIP{}
	}
	// If no value for ".spec.networking.publicIp.enabled" has been provided, use the default one.
	if mutatedObj.Spec.Networking.PublicIP.Enabled == nil {
		mutatedObj.Spec.Networking.PublicIP.Enabled = &PostgresqlInstanceSpecNetworkingPublicIPEnabledDefault
	}
	// If no value for ".spec.networking.publicIp.authorizedNetworks" has been provided, use the default one.
	if mutatedObj.Spec.Networking.PublicIP.AuthorizedNetworks == nil {
		mutatedObj.Spec.Networking.PublicIP.AuthorizedNetworks = make([]v1alpha1.PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetwork, 0)
	}
	// If the current request is an UPDATE request, make sure that ".spec.networking.privateIp.enabled" is not being changed from true to false.
	if previousObj != nil && *previousObj.Spec.Networking.PrivateIP.Enabled && !*mutatedObj.Spec.Networking.PrivateIP.Enabled {
		return fmt.Errorf("private ip access to the instance cannot be disabled after having been enabled")
	}
	// If the current request is an UPDATE request, make sure that ".spec.networking.privateIp.networking" is not being removed.
	if previousObj != nil && *previousObj.Spec.Networking.PrivateIP.Network != "" && *mutatedObj.Spec.Networking.PrivateIP.Network == "" {
		return fmt.Errorf("the resource link of the vpc network for the instance cannot be removed")
	}
	// Make sure that at least one of ".spec.networking.privateIp.enabled" and ".spec.networking.publicIp.enabled" are set to true.
	if !*mutatedObj.Spec.Networking.PrivateIP.Enabled && !*mutatedObj.Spec.Networking.PublicIP.Enabled {
		return fmt.Errorf("at least one of private or public ip access to the instance must be enabled")
	}
	// If ".spec.networking.privateIp.enabled" is true, validate the value of the ".spec.networking.privateIp.network" field.
	if *mutatedObj.Spec.Networking.PrivateIP.Enabled && *mutatedObj.Spec.Networking.PrivateIP.Network == "" {
		return fmt.Errorf("the resource link of the vpc network for the instance cannot be empty")
	}
	return nil
}

// validateAndMutatePostgresqlInstanceSpecResources validates and mutates the value of ".spec.resources".
func validateAndMutatePostgresqlInstanceSpecResources(mutatedObj, previousObj *v1alpha1.PostgresqlInstance) error {
	// Make sure that ".spec.resources" is initialized.
	if mutatedObj.Spec.Resources == nil {
		mutatedObj.Spec.Resources = &v1alpha1.PostgresqlInstanceSpecResources{}
	}
	// Make sure that ".spec.resources.disk" is initialized.
	if mutatedObj.Spec.Resources.Disk == nil {
		mutatedObj.Spec.Resources.Disk = &v1alpha1.PostgresqlInstanceSpecResourcesDisk{}
	}
	// If no value for ".spec.resources.disk.sizeMaximumGb" has been provided, use the default one.
	if mutatedObj.Spec.Resources.Disk.SizeMaximumGb == nil {
		mutatedObj.Spec.Resources.Disk.SizeMaximumGb = &PostgresqlInstanceSpecResourcesDiskSizeMaximumGbDefault
	}
	// If no value for ".spec.resources.disk.sizeMinimumGb" has been provided, use the default one.
	if mutatedObj.Spec.Resources.Disk.SizeMinimumGb == nil {
		mutatedObj.Spec.Resources.Disk.SizeMinimumGb = &PostgresqlInstanceSpecResourcesDiskSizeMinimumGbDefault
	}
	// If no value for ".spec.resources.disk.type" has been provided, use the default one.
	if mutatedObj.Spec.Resources.Disk.Type == nil {
		mutatedObj.Spec.Resources.Disk.Type = &PostgresqlInstanceSpecResourcesDiskTypeDefault
	}
	// If no value for  ".spec.resources.instanceType" has been provided, use the default one.
	if mutatedObj.Spec.Resources.InstanceType == nil {
		mutatedObj.Spec.Resources.InstanceType = &PostgresqlInstanceSpecResourcesInstanceTypeDefault
	}
	// If the current operation is an UPDATE request, make sure that ".spec.resources.disk.sizeMinimumGb" is not being decreased.
	if previousObj != nil && *mutatedObj.Spec.Resources.Disk.SizeMinimumGb < *previousObj.Spec.Resources.Disk.SizeMinimumGb {
		return fmt.Errorf("the minimum disk size for the instance cannot be decreased (had \"%d\", got \"%d\")", *previousObj.Spec.Resources.Disk.SizeMinimumGb, *mutatedObj.Spec.Resources.Disk.SizeMinimumGb)
	}
	// If the current operation is an UPDATE request, make sure that ".spec.resources.disk.type" is not being changed.
	if previousObj != nil && *previousObj.Spec.Resources.Disk.Type != *mutatedObj.Spec.Resources.Disk.Type {
		return fmt.Errorf("the disk type for the instance cannot be changed (had %q, got %q)", *previousObj.Spec.Resources.Disk.Type, *mutatedObj.Spec.Resources.Disk.Type)
	}
	// Make sure that ".spec.resources.disk.sizeMinimumGb" is greater than or equal to the minimum allowed disk size request.
	if *mutatedObj.Spec.Resources.Disk.SizeMinimumGb < postgresqlInstanceSpecResourcesDiskSizeMinimumGbLowerBound {
		return fmt.Errorf("the minimum disk size in gb for the instance is %d (got \"%d\")", postgresqlInstanceSpecResourcesDiskSizeMinimumGbLowerBound, *mutatedObj.Spec.Resources.Disk.SizeMinimumGb)
	}
	// Make sure that ".spec.resources.disk.sizeMaximumGb" is either 0 or greater than or equal to ".spec.resources.disk.sizeMinimumGb".
	if *mutatedObj.Spec.Resources.Disk.SizeMaximumGb != 0 && *mutatedObj.Spec.Resources.Disk.SizeMaximumGb < *mutatedObj.Spec.Resources.Disk.SizeMinimumGb {
		return fmt.Errorf("the maximum disk size in gb for the instance must be 0 or at least %d (got \"%d\")", *mutatedObj.Spec.Resources.Disk.SizeMinimumGb, *mutatedObj.Spec.Resources.Disk.SizeMaximumGb)
	}
	// Make sure that ".spec.resources.disk.type" contains a valid value.
	switch *mutatedObj.Spec.Resources.Disk.Type {
	case v1alpha1.PostgresqlInstanceSpecResourceDiskTypeHDD, v1alpha1.PostgresqlInstanceSpecResourceDiskTypeSSD:
		// The value is valid.
	default:
		return fmt.Errorf("the disk type for the instance must be one of %q or %q (got %q)", v1alpha1.PostgresqlInstanceSpecResourceDiskTypeHDD, v1alpha1.PostgresqlInstanceSpecResourceDiskTypeSSD, *mutatedObj.Spec.Resources.Disk.Type)
	}
	return nil
}

// validateAndMutatePostgresqlInstanceSpecVersion validates and mutates the value of ".spec.version".
func validateAndMutatePostgresqlInstanceSpecVersion(mutatedObj, previousObj *v1alpha1.PostgresqlInstance) error {
	// If the current operation is an UPDATE request, make sure that ".spec.version" is not being changed/removed.
	if previousObj != nil && *mutatedObj.Spec.Version != *previousObj.Spec.Version {
		return fmt.Errorf("the version of the instance cannot be changed")
	}
	// If no value for ".spec.version" has been provided, use the default one.
	if mutatedObj.Spec.Version == nil {
		mutatedObj.Spec.Version = &PostgresqlInstanceSpecVersionDefault
	}
	// Make sure that ".spec.version" contains a valid value.
	switch *mutatedObj.Spec.Version {
	case v1alpha1.PostgresqlInstanceSpecVersion96:
		// The value is valid.
	default:
		return fmt.Errorf("the version of the instance must be %q (got %q)", v1alpha1.PostgresqlInstanceSpecVersion96, *mutatedObj.Spec.Version)
	}
	return nil
}
