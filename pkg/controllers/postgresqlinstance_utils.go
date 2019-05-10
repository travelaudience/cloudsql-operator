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

package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	cloudsqladmin "google.golang.org/api/sqladmin/v1beta4"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/pointers"
)

// buildDatabaseInstance builds the DatabaseInstance object that corresponds to the specified PostgresqlInstance resource.
func buildDatabaseInstance(postgresqlInstance *v1alpha1api.PostgresqlInstance) *cloudsqladmin.DatabaseInstance {
	// Build the DatabaseInstance object.
	databaseInstance := &cloudsqladmin.DatabaseInstance{
		DatabaseVersion: postgresqlInstance.Spec.Version.APIValue(),
		Name:            postgresqlInstance.Spec.Name,
		Region:          *postgresqlInstance.Spec.Location.Region,
		Settings:        buildDatabaseInstanceSettings(postgresqlInstance),
	}
	// Force sending fields as required.
	setForceSendFields(databaseInstance)
	// Return the DatabaseInstance object.
	return databaseInstance
}

// buildDatabaseInstanceSettings builds the Settings field of the DatabaseInstance object that corresponds to the specified PostgresqlInstance resource.
func buildDatabaseInstanceSettings(postgresqlInstance *v1alpha1api.PostgresqlInstance) *cloudsqladmin.Settings {
	r := &cloudsqladmin.Settings{
		AvailabilityType: postgresqlInstance.Spec.Availability.Type.APIValue(),
		BackupConfiguration: &cloudsqladmin.BackupConfiguration{
			Enabled:   *postgresqlInstance.Spec.Backups.Daily.Enabled,
			StartTime: *postgresqlInstance.Spec.Backups.Daily.StartTime,
		},
		DatabaseFlags:  postgresqlInstance.Spec.Flags.APIValue(),
		DataDiskSizeGb: int64(*postgresqlInstance.Spec.Resources.Disk.SizeMinimumGb),
		DataDiskType:   postgresqlInstance.Spec.Resources.Disk.Type.APIValue(),
		IpConfiguration: &cloudsqladmin.IpConfiguration{
			AuthorizedNetworks: postgresqlInstance.Spec.Networking.PublicIP.AuthorizedNetworks.APIValue(),
			Ipv4Enabled:        *postgresqlInstance.Spec.Networking.PublicIP.Enabled,
		},
		LocationPreference: &cloudsqladmin.LocationPreference{
			Zone: postgresqlInstance.Spec.Location.Zone.APIValue(),
		},
		Tier:       *postgresqlInstance.Spec.Resources.InstanceType,
		UserLabels: postgresqlInstance.Spec.Labels,
	}
	if *postgresqlInstance.Spec.Maintenance.Day == v1alpha1api.PostgresqlInstanceSpecMaintenanceDayAny {
		r.MaintenanceWindow = &cloudsqladmin.MaintenanceWindow{}
	} else {
		r.MaintenanceWindow = &cloudsqladmin.MaintenanceWindow{
			Day:  postgresqlInstance.Spec.Maintenance.Day.APIValue(),
			Hour: postgresqlInstance.Spec.Maintenance.Hour.APIValue(),
		}
	}
	if *postgresqlInstance.Spec.Networking.PrivateIP.Enabled {
		r.IpConfiguration.PrivateNetwork = *postgresqlInstance.Spec.Networking.PrivateIP.Network
	}
	if *postgresqlInstance.Spec.Resources.Disk.SizeMaximumGb == *postgresqlInstance.Spec.Resources.Disk.SizeMinimumGb {
		r.StorageAutoResize = pointers.NewBool(false)
		r.StorageAutoResizeLimit = 0
	} else {
		r.StorageAutoResize = pointers.NewBool(true)
		r.StorageAutoResizeLimit = int64(*postgresqlInstance.Spec.Resources.Disk.SizeMaximumGb)
	}
	return r
}

// isOperationInProgressOrFailed indicates whether the last operation performed on the CSQLP instance associated with the provided PostgresqlInstance resource is still in progress, or has failed.
func (c *PostgresqlInstanceController) isOperationInProgressOrFailed(postgresqlInstance *v1alpha1api.PostgresqlInstance) (bool, string, string, string, string, error) {
	// Grab the list of operations for the CSQLP instance.
	ops, err := c.cloudsqlClient.Operations.List(c.projectID, postgresqlInstance.Spec.Name).Do()
	if err != nil {
		return false, "", "", "", "", err
	}
	// If there are no operations, there's nothing else to check.
	if len(ops.Items) == 0 {
		return false, "", "", "", "", nil
	}
	// Operations are sorted by reverse chronological order, so the last (or current) operation is the first item in the slice.
	lastOp := ops.Items[0]
	// If the last operation's status is "DONE" and there are no errors, there's nothing else to check.
	if lastOp.Status == constants.OperationStatusDone && (lastOp.Error == nil || len(lastOp.Error.Errors) == 0) {
		return false, lastOp.Name, lastOp.OperationType, lastOp.Status, "", nil
	}
	// Check whether there are any errors, and, in case there are, build the error message by concatenating them.
	errorMessage := ""
	if lastOp.Error != nil && len(lastOp.Error.Errors) > 0 {
		for _, err := range lastOp.Error.Errors {
			errorMessage += fmt.Sprintf("%s; %q", err.Code, err.Message)
		}
	}
	return true, lastOp.Name, lastOp.OperationType, lastOp.Status, errorMessage, nil
}

// updateDatabaseInstanceSettings updates the provided DatabaseInstance object according to the provided PostgresqlInstance resource.
func (c *PostgresqlInstanceController) updateDatabaseInstanceSettings(postgresqlInstance *v1alpha1api.PostgresqlInstance, databaseInstance *cloudsqladmin.DatabaseInstance) (mustUpdate bool) {
	// Compute the desired settings based on the provided PostgresqlInstance resource.
	desiredSettings := buildDatabaseInstanceSettings(postgresqlInstance)
	// Update each field of the provided CSQLP instance that differs from the desired value.
	if databaseInstance.Settings.AvailabilityType != desiredSettings.AvailabilityType {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.availabilityType must be updated")
		databaseInstance.Settings.AvailabilityType = desiredSettings.AvailabilityType
		mustUpdate = true
	}
	if databaseInstance.Settings.BackupConfiguration.Enabled != desiredSettings.BackupConfiguration.Enabled {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.backupConfiguration.enabled must be updated")
		databaseInstance.Settings.BackupConfiguration.Enabled = desiredSettings.BackupConfiguration.Enabled
		mustUpdate = true
	}
	if databaseInstance.Settings.BackupConfiguration.StartTime != desiredSettings.BackupConfiguration.StartTime {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.backupConfiguration.startTime must be updated")
		databaseInstance.Settings.BackupConfiguration.StartTime = desiredSettings.BackupConfiguration.StartTime
		mustUpdate = true
	}
	if !reflect.DeepEqual(databaseInstance.Settings.DatabaseFlags, desiredSettings.DatabaseFlags) {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.databaseFlags must be updated")
		databaseInstance.Settings.DatabaseFlags = desiredSettings.DatabaseFlags
		mustUpdate = true
	}
	if databaseInstance.Settings.DataDiskSizeGb != desiredSettings.DataDiskSizeGb {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.dataDiskSizeGb must be updated")
		databaseInstance.Settings.DataDiskSizeGb = desiredSettings.DataDiskSizeGb
		mustUpdate = true
	}
	if !reflect.DeepEqual(databaseInstance.Settings.IpConfiguration.AuthorizedNetworks, desiredSettings.IpConfiguration.AuthorizedNetworks) {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.ipConfiguration.authorizedNetworks must be updated")
		databaseInstance.Settings.IpConfiguration.AuthorizedNetworks = desiredSettings.IpConfiguration.AuthorizedNetworks
		mustUpdate = true
	}
	if databaseInstance.Settings.IpConfiguration.Ipv4Enabled != desiredSettings.IpConfiguration.Ipv4Enabled {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.ipConfiguration.ipv4Enabled must be updated")
		databaseInstance.Settings.IpConfiguration.Ipv4Enabled = desiredSettings.IpConfiguration.Ipv4Enabled
		mustUpdate = true
	}
	if databaseInstance.Settings.IpConfiguration.PrivateNetwork != desiredSettings.IpConfiguration.PrivateNetwork {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.ipConfiguration.privateNetwork must be updated")
		databaseInstance.Settings.IpConfiguration.PrivateNetwork = desiredSettings.IpConfiguration.PrivateNetwork
		mustUpdate = true
	}
	if *postgresqlInstance.Spec.Location.Zone != v1alpha1api.PostgresqlInstanceSpecLocationZoneAny && databaseInstance.Settings.LocationPreference.Zone != desiredSettings.LocationPreference.Zone {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.locationPreference.zone must be updated")
		databaseInstance.Settings.LocationPreference.Zone = desiredSettings.LocationPreference.Zone
		mustUpdate = true
	}
	if databaseInstance.Settings.MaintenanceWindow.Day != desiredSettings.MaintenanceWindow.Day {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.maintenanceWindow.day must be updated")
		databaseInstance.Settings.MaintenanceWindow.Day = desiredSettings.MaintenanceWindow.Day
		mustUpdate = true
	}
	if databaseInstance.Settings.MaintenanceWindow.Hour != desiredSettings.MaintenanceWindow.Hour {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.maintenanceWindow.hour must be updated")
		databaseInstance.Settings.MaintenanceWindow.Hour = desiredSettings.MaintenanceWindow.Hour
		mustUpdate = true
	}
	if *databaseInstance.Settings.StorageAutoResize != *desiredSettings.StorageAutoResize {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.storageAutoResize must be updated")
		*databaseInstance.Settings.StorageAutoResize = *desiredSettings.StorageAutoResize
		mustUpdate = true
	}
	if databaseInstance.Settings.StorageAutoResizeLimit != desiredSettings.StorageAutoResizeLimit {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.storageAutoResizeLimit must be updated")
		databaseInstance.Settings.StorageAutoResizeLimit = desiredSettings.StorageAutoResizeLimit
		mustUpdate = true
	}
	if databaseInstance.Settings.Tier != desiredSettings.Tier {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.tier must be updated")
		databaseInstance.Settings.Tier = desiredSettings.Tier
		mustUpdate = true
	}
	if !reflect.DeepEqual(databaseInstance.Settings.UserLabels, desiredSettings.UserLabels) {
		c.logger.WithField(logFieldName, postgresqlInstance.Name).Debug(".settings.userLabels must be updated")
		databaseInstance.Settings.UserLabels = desiredSettings.UserLabels
		mustUpdate = true
	}
	// Force sending fields as required.
	setForceSendFields(databaseInstance)
	return mustUpdate
}

// patchPostgresqlInstance updates the provided PostgresqlInstance using patch semantics.
// If there are no changes to be made, no patch is performed.
func (c *PostgresqlInstanceController) patchPostgresqlInstance(oldObj, newObj *v1alpha1api.PostgresqlInstance, subresources ...string) (*v1alpha1api.PostgresqlInstance, error) {
	// Return if there are no changes to be made.
	if reflect.DeepEqual(oldObj, newObj) {
		return newObj, nil
	}
	// Prepare the patch to apply based on the provided objects.
	oldBytes, err := json.Marshal(oldObj)
	if err != nil {
		return nil, err
	}
	newBytes, err := json.Marshal(newObj)
	if err != nil {
		return nil, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldBytes, newBytes, &v1alpha1api.PostgresqlInstance{})
	if err != nil {
		return nil, err
	}
	// Apply the patch.
	return c.selfClient.CloudsqlV1alpha1().PostgresqlInstances().Patch(oldObj.Name, types.MergePatchType, patchBytes, subresources...)
}

// patchPostgresqlInstanceStatus updates the status of the provided PostgresqlInstance using patch semantics.
// If there are no changes to be made, no patch is performed.
func (c *PostgresqlInstanceController) patchPostgresqlInstanceStatus(oldObj, newObj *v1alpha1api.PostgresqlInstance) (*v1alpha1api.PostgresqlInstance, error) {
	return c.patchPostgresqlInstance(oldObj, newObj, "status")
}

// setForceSendFields updates the provided DatabaseInstance object in order to force sending fields that would otherwise be omitted from the JSON representation due to the presence of the "omitempty" tag.
// This is required in order to, for example, be able to explicitly set ".settings.ipConfiguration.ipv4Enabled" to "false" or ".settings.maintenanceWindow.hour" to "0".
func setForceSendFields(databaseInstance *cloudsqladmin.DatabaseInstance) {
	databaseInstance.Settings.BackupConfiguration.ForceSendFields = []string{
		"Enabled",
	}
	databaseInstance.Settings.IpConfiguration.ForceSendFields = []string{
		"Ipv4Enabled",
		"PrivateNetwork",
	}
	databaseInstance.Settings.MaintenanceWindow.ForceSendFields = []string{
		"Hour",
	}
	databaseInstance.Settings.ForceSendFields = []string{
		"StorageAutoResizeLimit",
	}
}

// setPostgresqlInstanceCondition sets a condition on the provided PostgresqlInstance resource according to the following rules:
// 1. If no condition of the provided type exists, the condition is inserted with its last transition time set to the current time.
// 2. If a condition of the provided type and state exists, the condition is updated but its last transition time is not modified.
// 3. If a condition of the provided type but different state exists, the condition is updated and its last transition time is set to the current time.
func setPostgresqlInstanceCondition(postgresqlInstance *v1alpha1api.PostgresqlInstance, conditionType v1alpha1api.PostgresqlInstanceStatusConditionType, conditionStatus corev1.ConditionStatus, conditionReason string, conditionMessage string) {
	// Create the new condition.
	newCondition := v1alpha1api.PostgresqlInstanceStatusCondition{
		LastTransitionTime: v1.NewTime(time.Now()),
		Message:            conditionMessage,
		Reason:             conditionReason,
		Status:             conditionStatus,
		Type:               conditionType,
	}
	// Search through existing conditions in order to understand if we need to insert the new condition or not.
	for idx, cdn := range postgresqlInstance.Status.Conditions {
		// If the current condition's type is different from the one we will be inserting, skip it.
		if cdn.Type != newCondition.Type {
			continue
		}
		// If the status is the same, we should not update the condition's last transition time.
		if cdn.Status == newCondition.Status {
			newCondition.LastTransitionTime = cdn.LastTransitionTime
		}
		// Overwrite the existing condition and return.
		postgresqlInstance.Status.Conditions[idx] = newCondition
		return
	}
	// At this point we know that there is no existing condition with this type, so we just append it to the set of conditions.
	postgresqlInstance.Status.Conditions = append(postgresqlInstance.Status.Conditions, newCondition)
}

// setPostgresqlInstanceConnectionNameAndIPs sets the connection name and the set of IPs of associated with the provided CSQLP instance.
func setPostgresqlInstanceConnectionNameAndIPs(postgresqlInstance *v1alpha1api.PostgresqlInstance, databaseInstance *cloudsqladmin.DatabaseInstance) {
	postgresqlInstance.Status.IPs = v1alpha1api.PostgresqlInstanceStatusIPAddresses{}
	for _, ip := range databaseInstance.IpAddresses {
		if ip != nil {
			switch ip.Type {
			case constants.DatabaseInstanceIPAddressTypePrivate:
				postgresqlInstance.Status.IPs.PrivateIP = ip.IpAddress
			case constants.DatabaseInstanceIPAddressTypePublic:
				postgresqlInstance.Status.IPs.PublicIP = ip.IpAddress
			default:
				continue
			}
		}
	}
	postgresqlInstance.Status.ConnectionName = databaseInstance.ConnectionName
}
