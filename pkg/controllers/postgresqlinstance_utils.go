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
	cloudsqladmin "google.golang.org/api/sqladmin/v1beta4"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/pointers"
)

// buildDatabaseInstance builds the DatabaseInstance object that corresponds to the specified PostgresqlInstance resource.
func buildDatabaseInstance(postgresqlInstance *v1alpha1api.PostgresqlInstance) *cloudsqladmin.DatabaseInstance {
	// Create the base DatabaseInstance object.
	instance := &cloudsqladmin.DatabaseInstance{
		Settings: &cloudsqladmin.Settings{
			BackupConfiguration: &cloudsqladmin.BackupConfiguration{},
			IpConfiguration:     &cloudsqladmin.IpConfiguration{},
			LocationPreference:  &cloudsqladmin.LocationPreference{},
			MaintenanceWindow:   &cloudsqladmin.MaintenanceWindow{},
		},
	}
	// Observe ".spec.availability".
	instance.Settings.AvailabilityType = postgresqlInstance.Spec.Availability.Type.APIValue()
	// Observe ".spec.backups".
	instance.Settings.BackupConfiguration.Enabled = *postgresqlInstance.Spec.Backups.Daily.Enabled
	instance.Settings.BackupConfiguration.StartTime = *postgresqlInstance.Spec.Backups.Daily.StartTime
	// Observe ".spec.flags".
	instance.Settings.DatabaseFlags = postgresqlInstance.Spec.Flags.APIValue()
	// Observe ".spec.labels".
	instance.Settings.UserLabels = postgresqlInstance.Spec.Labels
	// Observe ".spec.location".
	instance.Region = *postgresqlInstance.Spec.Location.Region
	instance.Settings.LocationPreference.Zone = postgresqlInstance.Spec.Location.Zone.APIValue()
	// Observe ".spec.maintenance".
	if *postgresqlInstance.Spec.Maintenance.Day != v1alpha1api.PostgresqlInstanceSpecMaintenanceDayAny {
		instance.Settings.MaintenanceWindow.Day = postgresqlInstance.Spec.Maintenance.Day.APIValue()
		instance.Settings.MaintenanceWindow.Hour = postgresqlInstance.Spec.Maintenance.Hour.APIValue()
	}
	// Observe ".spec.name".
	instance.Name = postgresqlInstance.Spec.Name
	// Observe ".spec.network".
	if *postgresqlInstance.Spec.Networking.PrivateIP.Enabled {
		instance.Settings.IpConfiguration.PrivateNetwork = *postgresqlInstance.Spec.Networking.PrivateIP.Network
	}
	instance.Settings.IpConfiguration.Ipv4Enabled = *postgresqlInstance.Spec.Networking.PublicIP.Enabled
	instance.Settings.IpConfiguration.AuthorizedNetworks = postgresqlInstance.Spec.Networking.PublicIP.AuthorizedNetworks.APIValue()
	// Observe ".spec.resources".
	if *postgresqlInstance.Spec.Resources.Disk.SizeMaximumGb == *postgresqlInstance.Spec.Resources.Disk.SizeMinimumGb {
		instance.Settings.StorageAutoResize = pointers.NewBool(false)
		instance.Settings.StorageAutoResizeLimit = 0
	} else {
		instance.Settings.StorageAutoResize = pointers.NewBool(true)
		instance.Settings.StorageAutoResizeLimit = int64(*postgresqlInstance.Spec.Resources.Disk.SizeMaximumGb)
	}
	instance.Settings.DataDiskSizeGb = int64(*postgresqlInstance.Spec.Resources.Disk.SizeMinimumGb)
	instance.Settings.DataDiskType = postgresqlInstance.Spec.Resources.Disk.Type.APIValue()
	instance.Settings.Tier = *postgresqlInstance.Spec.Resources.InstanceType
	// Observe ".spec.version".
	instance.DatabaseVersion = postgresqlInstance.Spec.Version.APIValue()
	// Return the DatabaseInstance object.
	return instance
}
