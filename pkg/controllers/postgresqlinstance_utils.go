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
	"reflect"

	log "github.com/sirupsen/logrus"
	cloudsqladmin "google.golang.org/api/sqladmin/v1beta4"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/pointers"
)

// buildDatabaseInstance builds the DatabaseInstance object that corresponds to the specified PostgresqlInstance resource.
func buildDatabaseInstance(postgresqlInstance *v1alpha1api.PostgresqlInstance) *cloudsqladmin.DatabaseInstance {
	return &cloudsqladmin.DatabaseInstance{
		DatabaseVersion: postgresqlInstance.Spec.Version.APIValue(),
		Name:            postgresqlInstance.Spec.Name,
		Region:          *postgresqlInstance.Spec.Location.Region,
		Settings:        buildDatabaseInstanceSettings(postgresqlInstance),
	}
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

// updateDatabaseInstanceSettings updates the provided DatabaseInstance object according to the provided PostgresqlInstance resource.
func updateDatabaseInstanceSettings(postgresqlInstance *v1alpha1api.PostgresqlInstance, databaseInstance *cloudsqladmin.DatabaseInstance) (mustUpdate bool) {
	// Compute the desired settings based on the provided PostgresqlInstance resource.
	desiredSettings := buildDatabaseInstanceSettings(postgresqlInstance)
	// Update each field of the provided CSQLP instance that differs from the desired value.
	if databaseInstance.Settings.AvailabilityType != desiredSettings.AvailabilityType {
		log.Debug(".settings.availabilityType must be updated")
		databaseInstance.Settings.AvailabilityType = desiredSettings.AvailabilityType
		mustUpdate = true
	}
	if databaseInstance.Settings.BackupConfiguration.Enabled != desiredSettings.BackupConfiguration.Enabled {
		log.Debug(".settings.backupConfiguration.enabled must be updated")
		databaseInstance.Settings.BackupConfiguration.Enabled = desiredSettings.BackupConfiguration.Enabled
		mustUpdate = true
	}
	if databaseInstance.Settings.BackupConfiguration.StartTime != desiredSettings.BackupConfiguration.StartTime {
		log.Debug(".settings.backupConfiguration.startTime must be updated")
		databaseInstance.Settings.BackupConfiguration.StartTime = desiredSettings.BackupConfiguration.StartTime
		mustUpdate = true
	}
	if !reflect.DeepEqual(databaseInstance.Settings.DatabaseFlags, desiredSettings.DatabaseFlags) {
		log.Debug(".settings.databaseFlags must be updated")
		databaseInstance.Settings.DatabaseFlags = desiredSettings.DatabaseFlags
		mustUpdate = true
	}
	if databaseInstance.Settings.DataDiskSizeGb != desiredSettings.DataDiskSizeGb {
		log.Debug(".settings.dataDiskSizeGb must be updated")
		databaseInstance.Settings.DataDiskSizeGb = desiredSettings.DataDiskSizeGb
		mustUpdate = true
	}
	if !reflect.DeepEqual(databaseInstance.Settings.IpConfiguration.AuthorizedNetworks, desiredSettings.IpConfiguration.AuthorizedNetworks) {
		log.Debug(".settings.ipConfiguration.authorizedNetworks must be updated")
		databaseInstance.Settings.IpConfiguration.AuthorizedNetworks = desiredSettings.IpConfiguration.AuthorizedNetworks
		mustUpdate = true
	}
	if databaseInstance.Settings.IpConfiguration.Ipv4Enabled != desiredSettings.IpConfiguration.Ipv4Enabled {
		log.Debug(".settings.ipConfiguration.ipv4Enabled must be updated")
		databaseInstance.Settings.IpConfiguration.Ipv4Enabled = desiredSettings.IpConfiguration.Ipv4Enabled
		mustUpdate = true
	}
	if databaseInstance.Settings.IpConfiguration.PrivateNetwork != desiredSettings.IpConfiguration.PrivateNetwork {
		log.Debug(".settings.ipConfiguration.privateNetwork must be updated")
		databaseInstance.Settings.IpConfiguration.PrivateNetwork = desiredSettings.IpConfiguration.PrivateNetwork
		mustUpdate = true
	}
	if *postgresqlInstance.Spec.Location.Zone != v1alpha1api.PostgresqlInstanceSpecLocationZoneAny && databaseInstance.Settings.LocationPreference.Zone != desiredSettings.LocationPreference.Zone {
		log.Debug(".settings.locationPreference.zone must be updated")
		databaseInstance.Settings.LocationPreference.Zone = desiredSettings.LocationPreference.Zone
		mustUpdate = true
	}
	if databaseInstance.Settings.MaintenanceWindow.Day != desiredSettings.MaintenanceWindow.Day {
		log.Debug(".settings.maintenanceWindow.day must be updated")
		databaseInstance.Settings.MaintenanceWindow.Day = desiredSettings.MaintenanceWindow.Day
		mustUpdate = true
	}
	if databaseInstance.Settings.MaintenanceWindow.Hour != desiredSettings.MaintenanceWindow.Hour {
		log.Debug(".settings.maintenanceWindow.hour must be updated")
		databaseInstance.Settings.MaintenanceWindow.Hour = desiredSettings.MaintenanceWindow.Hour
		mustUpdate = true
	}
	if *databaseInstance.Settings.StorageAutoResize != *desiredSettings.StorageAutoResize {
		log.Debug(".settings.storageAutoResize must be updated")
		*databaseInstance.Settings.StorageAutoResize = *desiredSettings.StorageAutoResize
		mustUpdate = true
	}
	if databaseInstance.Settings.StorageAutoResizeLimit != desiredSettings.StorageAutoResizeLimit {
		log.Debug(".settings.storageAutoResizeLimit must be updated")
		databaseInstance.Settings.StorageAutoResizeLimit = desiredSettings.StorageAutoResizeLimit
		mustUpdate = true
	}
	if databaseInstance.Settings.Tier != desiredSettings.Tier {
		log.Debug(".settings.tier must be updated")
		databaseInstance.Settings.Tier = desiredSettings.Tier
		mustUpdate = true
	}
	return mustUpdate
}
