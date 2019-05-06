// +build e2e

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

package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudsqladmin "google.golang.org/api/sqladmin/v1beta4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1api "github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-postgres-operator/test/e2e/framework"
)

const (
	// postgresqlDriverName is the name of the SQL driver to use when connecting to PostgreSQL.
	postgresqlDriverName = "postgres"
	// postgresqlConnectionStringFormat is a format string used to build the connection string to use when connecting to PostgreSQL.
	// The formatting directives correspond to:
	// 1. Username.
	// 2. Password (URL encoded).
	// 3. Host.
	// 4. Database.
	postgresqlConnectionStringFormat = postgresqlDriverName + "://%s:%s@%s:5432/%s"
	// selectCurrentUserQuery is the SQL query that returns the name of the current user.
	selectCurrentUserQuery = "SELECT current_user;"
	// waitUntilPostgresqlInstanceStatusConditionTimeout is the timeout used while waiting for a given condition to be reported on a PostgresqlInstance resource.
	waitUntilPostgresqlInstanceStatusConditionTimeout = 10 * time.Minute
)

var _ = Describe("CSQLP instances", func() {
	framework.LifecycleIt("are created, updated and deleted as expected", func() {
		var (
			databaseInstance         *cloudsqladmin.DatabaseInstance
			db                       *sql.DB
			err                      error
			username                 string
			password                 string
			postgresqlInstance       *v1alpha1api.PostgresqlInstance
			postgresqlInstanceSecret *corev1.Secret
			publicIp                 string
			selectCurrentUserValue   string
		)

		By("Creating a PostgresqlInstance resource with public IP disabled")

		// Create a PostgresqlInstance resource.
		var (
			availabilityType      = v1alpha1api.PostgresqlInstanceSpecAvailabilityTypeRegional
			dailyBackupsEnabled   = false
			dailyBackupsStartTime = "06:00"
			diskSizeMaximumGb     = int32(0)
			diskSizeMinimumGb     = int32(10)
			diskType              = v1alpha1api.PostgresqlInstanceSpecResourceDiskTypeHDD
			flags                 = []string{
				"log_connections=on",
				"log_disconnections=on",
			}
			instanceType = "db-custom-2-7680"
			labels       = map[string]string{
				"e2e": "true",
			}
			maintenanceDay             = v1alpha1api.PostgresqlInstanceSpecMaintenanceDayMonday
			maintenanceHour            = v1alpha1api.PostgresqlInstanceSpecMaintenanceHour("00:00")
			privateIpEnabled           = true
			privateIpNetwork           = f.BuildPrivateNetworkResourceLink("default")
			publicIpAuthorizedNetworks = v1alpha1api.PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetworkList{}
			publicIpEnabled            = false
			region                     = "europe-west4"
			zone                       = v1alpha1api.PostgresqlInstanceSpecLocationZone("europe-west4-b")
			version                    = v1alpha1api.PostgresqlInstanceSpecVersion96
		)
		postgresqlInstance = &v1alpha1api.PostgresqlInstance{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: framework.PostgresqlInstanceMetadataNamePrefix,
			},
			Spec: v1alpha1api.PostgresqlInstanceSpec{
				Availability: &v1alpha1api.PostgresqlInstanceSpecAvailability{
					Type: &availabilityType,
				},
				Backups: &v1alpha1api.PostgresqlInstanceSpecBackups{
					Daily: &v1alpha1api.PostgresqlInstancSpecBackupsDaily{
						Enabled:   &dailyBackupsEnabled,
						StartTime: &dailyBackupsStartTime,
					},
				},
				Flags:  flags,
				Labels: labels,
				Location: &v1alpha1api.PostgresqlInstanceSpecLocation{
					Region: &region,
					Zone:   &zone,
				},
				Maintenance: &v1alpha1api.PostgresqlInstanceSpecMaintenance{
					Day:  &maintenanceDay,
					Hour: &maintenanceHour,
				},
				Name: f.NewRandomPostgresqlInstanceSpecName(),
				Networking: &v1alpha1api.PostgresqlInstanceSpecNetworking{
					PrivateIP: &v1alpha1api.PostgresqlInstanceSpecNetworkingPrivateIP{
						Enabled: &privateIpEnabled,
						Network: &privateIpNetwork,
					},
					PublicIP: &v1alpha1api.PostgresqlInstanceSpecNetworkingPublicIP{
						AuthorizedNetworks: publicIpAuthorizedNetworks,
						Enabled:            &publicIpEnabled,
					},
				},
				Resources: &v1alpha1api.PostgresqlInstanceSpecResources{
					Disk: &v1alpha1api.PostgresqlInstanceSpecResourcesDisk{
						SizeMaximumGb: &diskSizeMaximumGb,
						SizeMinimumGb: &diskSizeMinimumGb,
						Type:          &diskType,
					},
					InstanceType: &instanceType,
				},
				Version: &version,
			},
		}
		postgresqlInstance, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Create(postgresqlInstance)
		Expect(err).NotTo(HaveOccurred())
		Expect(postgresqlInstance).NotTo(BeNil())

		By(`waiting for the "Created" condition to be "True"`)

		ctx1, fn1 := context.WithTimeout(context.Background(), waitUntilPostgresqlInstanceStatusConditionTimeout)
		defer fn1()
		err = f.WaitUntilPostgresqlInstanceStatusCondition(ctx1, postgresqlInstance, v1alpha1api.PostgresqlInstanceStatusConditionTypeCreated, corev1.ConditionTrue)
		Expect(err).NotTo(HaveOccurred())

		By(`waiting for the "Ready" condition to be "True"`)

		ctx2, fn2 := context.WithTimeout(context.Background(), waitUntilPostgresqlInstanceStatusConditionTimeout)
		defer fn2()
		err = f.WaitUntilPostgresqlInstanceStatusCondition(ctx2, postgresqlInstance, v1alpha1api.PostgresqlInstanceStatusConditionTypeReady, corev1.ConditionTrue)
		Expect(err).NotTo(HaveOccurred())

		By(`checking that the CSQLP instance has all fields correctly set`)

		databaseInstance, err = f.CloudSQLClient.Instances.Get(f.ProjectId, postgresqlInstance.Spec.Name).Do()
		Expect(err).NotTo(HaveOccurred())
		Expect(databaseInstance).NotTo(BeNil())
		Expect(databaseInstance.Settings.AvailabilityType).To(Equal(postgresqlInstance.Spec.Availability.Type.APIValue()))
		Expect(databaseInstance.Settings.BackupConfiguration.Enabled).To(Equal(*postgresqlInstance.Spec.Backups.Daily.Enabled))
		Expect(databaseInstance.Settings.BackupConfiguration.StartTime).To(Equal(*postgresqlInstance.Spec.Backups.Daily.StartTime))
		Expect(databaseInstance.Settings.DatabaseFlags).To(Equal(postgresqlInstance.Spec.Flags.APIValue()))
		Expect(databaseInstance.Settings.DataDiskSizeGb).To(Equal(int64(*postgresqlInstance.Spec.Resources.Disk.SizeMinimumGb)))
		Expect(databaseInstance.Settings.DataDiskType).To(Equal(postgresqlInstance.Spec.Resources.Disk.Type.APIValue()))
		Expect(databaseInstance.Settings.IpConfiguration.Ipv4Enabled).To(Equal(*postgresqlInstance.Spec.Networking.PublicIP.Enabled))
		Expect(databaseInstance.Settings.IpConfiguration.AuthorizedNetworks).To(Equal(postgresqlInstance.Spec.Networking.PublicIP.AuthorizedNetworks.APIValue()))
		Expect(databaseInstance.Settings.IpConfiguration.PrivateNetwork).To(Equal(*postgresqlInstance.Spec.Networking.PrivateIP.Network))
		Expect(databaseInstance.Settings.MaintenanceWindow.Day).To(Equal(postgresqlInstance.Spec.Maintenance.Day.APIValue()))
		Expect(databaseInstance.Settings.MaintenanceWindow.Hour).To(Equal(postgresqlInstance.Spec.Maintenance.Hour.APIValue()))
		Expect(databaseInstance.Settings.LocationPreference.Zone).To(Equal(postgresqlInstance.Spec.Location.Zone.APIValue()))
		Expect(databaseInstance.Settings.Tier).To(Equal(*postgresqlInstance.Spec.Resources.InstanceType))
		Expect(databaseInstance.Settings.StorageAutoResizeLimit).To(Equal(int64(*postgresqlInstance.Spec.Resources.Disk.SizeMaximumGb)))
		Expect(*databaseInstance.Settings.StorageAutoResize).To(Equal(*postgresqlInstance.Spec.Resources.Disk.SizeMaximumGb == int32(0)))
		Expect(databaseInstance.Settings.UserLabels).To(Equal(postgresqlInstance.Spec.Labels))

		By(`checking that the private IP of the CSQLP instance has been reported (and that no public IP has)`)

		postgresqlInstance, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Get(postgresqlInstance.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(postgresqlInstance.Status.IPs.PrivateIP).NotTo(BeEmpty())
		Expect(postgresqlInstance.Status.IPs.PublicIP).To(BeEmpty())

		By(`checking that the secret containing the password for the "postgres" user has been created and contains the PGUSER and PGPASS keys`)

		postgresqlInstanceSecret, err = f.KubeClient.CoreV1().Secrets(f.Namespace).Get(postgresqlInstance.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(postgresqlInstanceSecret.Data).To(HaveKey(constants.PostgresqlInstanceUsernameKey))
		Expect(postgresqlInstanceSecret.Data).To(HaveKey(constants.PostgresqlInstancePasswordKey))

		By(`checking that the PGUSER and PGPASS keys have adequate values`)

		username = string(postgresqlInstanceSecret.Data[constants.PostgresqlInstanceUsernameKey])
		Expect(err).NotTo(HaveOccurred())
		Expect(username).To(Equal(constants.PostgresqlInstanceUsernameValue))
		password = string(postgresqlInstanceSecret.Data[constants.PostgresqlInstancePasswordKey])
		Expect(err).NotTo(HaveOccurred())
		Expect(password).NotTo(BeNil())

		By(`enabling public IP networking and adding the external IP of the current host to the list of authorized networks`)

		publicIpEnabled = true
		publicIpAuthorizedNetworks = append(publicIpAuthorizedNetworks, v1alpha1api.PostgresqlInstanceSpecNetworkingPublicIPAuthorizedNetwork{
			Cidr: fmt.Sprintf("%s/32", f.ExternalIP),
		})
		postgresqlInstance.Spec.Networking.PublicIP.Enabled = &publicIpEnabled
		postgresqlInstance.Spec.Networking.PublicIP.AuthorizedNetworks = publicIpAuthorizedNetworks
		postgresqlInstance, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Update(postgresqlInstance)
		Expect(err).NotTo(HaveOccurred())

		By(`waiting for the "Ready" condition to switch to "False" and back to "True"`)

		ctx3, fn3 := context.WithTimeout(context.Background(), waitUntilPostgresqlInstanceStatusConditionTimeout)
		defer fn3()
		err = f.WaitUntilPostgresqlInstanceStatusCondition(ctx3, postgresqlInstance, v1alpha1api.PostgresqlInstanceStatusConditionTypeReady, corev1.ConditionFalse)
		Expect(err).NotTo(HaveOccurred())

		ctx4, fn4 := context.WithTimeout(context.Background(), waitUntilPostgresqlInstanceStatusConditionTimeout)
		defer fn4()
		err = f.WaitUntilPostgresqlInstanceStatusCondition(ctx4, postgresqlInstance, v1alpha1api.PostgresqlInstanceStatusConditionTypeReady, corev1.ConditionTrue)
		Expect(err).NotTo(HaveOccurred())

		By(`checking that a public IP for the CSQLP instance has now been reported, and that the private IP keeps being reported`)

		postgresqlInstance, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Get(postgresqlInstance.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		publicIp = postgresqlInstance.Status.IPs.PublicIP
		Expect(publicIp).NotTo(BeEmpty())
		Expect(postgresqlInstance.Status.IPs.PrivateIP).NotTo(BeEmpty())

		By(`attempting to connect to the instance via its public IP using the values of PGUSER and PGPASS`)

		db, err = sql.Open(postgresqlDriverName, fmt.Sprintf(postgresqlConnectionStringFormat, username, url.QueryEscape(password), publicIp, username))
		Expect(err).NotTo(HaveOccurred())

		err = db.QueryRow(selectCurrentUserQuery).Scan(&selectCurrentUserValue)
		Expect(err).NotTo(HaveOccurred())
		Expect(selectCurrentUserValue).To(Equal(username))

		err = db.Close()
		Expect(err).NotTo(HaveOccurred())

		By("deleting the PostgresqlInstance resource")

		err = f.DeletePostgresqlInstanceByName(postgresqlInstance.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})
