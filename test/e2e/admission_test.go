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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/travelaudience/cloudsql-postgres-operator/pkg/admission"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/apis/cloudsql/v1alpha1"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/constants"
	"github.com/travelaudience/cloudsql-postgres-operator/pkg/util/pointers"
	"github.com/travelaudience/cloudsql-postgres-operator/test/e2e/framework"
)

var _ = Describe("PostgresqlInstance", func() {
	framework.AdmissionIt("is mutated with default values upon creation", func() {
		var (
			err error
			obj *v1alpha1.PostgresqlInstance
		)

		// Create a minimal PostgresqlInstance resource.
		obj, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Create(&v1alpha1.PostgresqlInstance{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: framework.PostgresqlInstanceMetadataNamePrefix,
			},
			Spec: v1alpha1.PostgresqlInstanceSpec{
				Name: f.NewRandomPostgresqlInstanceSpecName(),
				Networking: &v1alpha1.PostgresqlInstanceSpecNetworking{
					PublicIP: &v1alpha1.PostgresqlInstanceSpecNetworkingPublicIP{
						Enabled: pointers.NewBool(true),
					},
				},
				Paused: true,
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// Make sure that the "cloudsql.travelaudience.com/allow-deletion" annotation has the expected value.
		Expect(obj.Annotations).To(HaveKeyWithValue(constants.AllowDeletionAnnotationKey, v1alpha1.False))

		// Make sure that all fields have the expected values.
		Expect(*obj.Spec.Availability.Type).To(Equal(admission.PostgresqlInstanceSpecAvailabilityTypeDefault))
		Expect(*obj.Spec.Backups.Daily.Enabled).To(Equal(admission.PostgresqlInstanceSpecBackupsDailyEnabledDefault))
		Expect(*obj.Spec.Backups.Daily.StartTime).To(Equal(admission.PostgresqlInstanceSpecBackupsDailyStartTimeDefault))
		Expect(obj.Spec.Flags).To(HaveLen(0))
		Expect(obj.Spec.Labels).To(HaveKeyWithValue(admission.PostgresqlInstanceSpecLabelsOwnerName, constants.ApplicationName))
		Expect(*obj.Spec.Location.Region).To(Equal(admission.PostgresqlInstanceSpecLocationRegionDefault))
		Expect(*obj.Spec.Location.Zone).To(Equal(admission.PostgresqlInstanceSpecLocationZoneDefault))
		Expect(*obj.Spec.Maintenance.Day).To(Equal(admission.PostgresqlInstanceSpecMaintenanceDayDefault))
		Expect(*obj.Spec.Maintenance.Hour).To(Equal(admission.PostgresqlInstanceSpecMaintenanceHourDefault))
		Expect(*obj.Spec.Networking.PrivateIP.Enabled).To(Equal(admission.PostgresqlInstanceSpecNetworkingPrivateIPEnabledDefault))
		Expect(*obj.Spec.Networking.PrivateIP.Network).To(Equal(admission.PostgresqlInstanceSpecNetworkingPrivateIPNetworkDefault))
		Expect(*obj.Spec.Networking.PublicIP.Enabled).To(BeTrue()) // NOTE: We have explicitly set ".spec.networking.publicIp.enabled" to "true" above.
		Expect(obj.Spec.Networking.PublicIP.AuthorizedNetworks).To(HaveLen(0))
		Expect(*obj.Spec.Resources.Disk.SizeMaximumGb).To(Equal(admission.PostgresqlInstanceSpecResourcesDiskSizeMaximumGbDefault))
		Expect(*obj.Spec.Resources.Disk.SizeMinimumGb).To(Equal(admission.PostgresqlInstanceSpecResourcesDiskSizeMinimumGbDefault))
		Expect(*obj.Spec.Resources.Disk.Type).To(Equal(admission.PostgresqlInstanceSpecResourcesDiskTypeDefault))
		Expect(*obj.Spec.Resources.InstanceType).To(Equal(admission.PostgresqlInstanceSpecResourcesInstanceTypeDefault))
		Expect(*obj.Spec.Version).To(Equal(admission.PostgresqlInstanceSpecVersionDefault))

		// Delete the PostgresqlInstance resource.
		err = f.DeletePostgresqlInstanceByName(obj.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	framework.AdmissionIt("cannot be deleted unless \"cloudsql.travelaudience.com/allow-deletion\" is \"true\"", func() {
		var (
			err error
			obj *v1alpha1.PostgresqlInstance
		)

		// Create a minimal PostgresqlInstance resource.
		obj, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Create(&v1alpha1.PostgresqlInstance{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: framework.PostgresqlInstanceMetadataNamePrefix,
			},
			Spec: v1alpha1.PostgresqlInstanceSpec{
				Name: f.NewRandomPostgresqlInstanceSpecName(),
				Networking: &v1alpha1.PostgresqlInstanceSpecNetworking{
					PublicIP: &v1alpha1.PostgresqlInstanceSpecNetworkingPublicIP{
						Enabled: pointers.NewBool(true),
					},
				},
				Paused: true,
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// Make sure that the PostgresqlInstance resource cannot be deleted, and that an adequate error message is returned.
		err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Delete(obj.Name, &metav1.DeleteOptions{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp(`the resource cannot be deleted unless the "cloudsql.travelaudience.com/allow-deletion" annotation is set to "true"`))

		// Update the value of the "cloudsql.travelaudience.com/allow-deletion" annotation, setting it to "true".
		obj.Annotations[constants.AllowDeletionAnnotationKey] = v1alpha1.True
		obj, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Update(obj)
		Expect(err).NotTo(HaveOccurred())
		Expect(obj.Annotations).To(HaveKeyWithValue(constants.AllowDeletionAnnotationKey, v1alpha1.True))

		// Make sure that the PostgresqlInstance can be deleted.
		err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Delete(obj.Name, metav1.NewDeleteOptions(0))
		Expect(err).NotTo(HaveOccurred())
	})

	framework.AdmissionIt("cannot be created with invalid values", func() {
		tests := []struct {
			errorMessageRegex string
			fn                func(*v1alpha1.PostgresqlInstance)
		}{
			{
				errorMessageRegex: `the availability type of the instance must be one of "Regional" or "Zonal" \(got "foo"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					t := v1alpha1.PostgresqlInstanceSpecAvailabilityType("foo")
					instance.Spec.Availability = &v1alpha1.PostgresqlInstanceSpecAvailability{
						Type: &t,
					}
				},
			},
			{
				errorMessageRegex: `the start time for daily backups of the instance must be a valid hour of the day in 24-hour format \(got "foo"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Backups = &v1alpha1.PostgresqlInstanceSpecBackups{
						Daily: &v1alpha1.PostgresqlInstancSpecBackupsDaily{
							Enabled:   pointers.NewBool(true),
							StartTime: pointers.NewString("foo"),
						},
					}
				},
			},
			{
				errorMessageRegex: `flags must be specified in the "<name>=<value>" format \(got "foo-bar"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Flags = []string{
						"foo-bar",
					}
				},
			},
			{
				errorMessageRegex: `the day of the week for periodic maintenance must be "Any" or a valid weekday \(got "foo"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					d := v1alpha1.PostgresqlInstanceSpecMaintenanceDay("foo")
					instance.Spec.Maintenance = &v1alpha1.PostgresqlInstanceSpecMaintenance{
						Day: &d,
					}
				},
			},
			{
				errorMessageRegex: `the hour of the day for periodic maintenance must be "Any" or a valid hour of the day in 24-hour format \(got "foo"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Maintenance = &v1alpha1.PostgresqlInstanceSpecMaintenance{
						Hour: pointers.NewString("foo"),
					}
				},
			},
			{
				errorMessageRegex: `the name of the instance cannot be empty`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Name = ""
				},
			},
			{
				errorMessageRegex: `the name of the instance must match the ".*" regular expression \(got "cloud\$ql-inst@nce"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Name = "cloud$ql-inst@nce"
				},
			},
			{
				errorMessageRegex: `the name of the instance must not exceed \d+ characters \(got "very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Name = "very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name-very-long-name"
				},
			},
			{
				errorMessageRegex: `at least one of private or public ip access to the instance must be enabled`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Networking = nil
				},
			},
			{
				errorMessageRegex: `the resource link of the vpc network for the instance cannot be empty`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Networking.PrivateIP = &v1alpha1.PostgresqlInstanceSpecNetworkingPrivateIP{
						Enabled: pointers.NewBool(true),
						Network: pointers.NewString(""),
					}
				},
			},
			{
				errorMessageRegex: `the minimum disk size in gb for the instance is 10 \(got "1"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Resources = &v1alpha1.PostgresqlInstanceSpecResources{
						Disk: &v1alpha1.PostgresqlInstanceSpecResourcesDisk{
							SizeMinimumGb: pointers.NewInt32(1),
						},
					}
				},
			},
			{
				errorMessageRegex: `the maximum disk size in gb for the instance must be 0 or at least 12 \(got "10"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Resources = &v1alpha1.PostgresqlInstanceSpecResources{
						Disk: &v1alpha1.PostgresqlInstanceSpecResourcesDisk{
							SizeMaximumGb: pointers.NewInt32(10),
							SizeMinimumGb: pointers.NewInt32(12),
						},
					}
				},
			},
			{
				errorMessageRegex: `the disk type for the instance must be one of "HDD" or "SSD" \(got "foo"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					t := v1alpha1.PostgresqlInstanceSpecResourcesDiskType("foo")
					instance.Spec.Resources = &v1alpha1.PostgresqlInstanceSpecResources{
						Disk: &v1alpha1.PostgresqlInstanceSpecResourcesDisk{
							Type: &t,
						},
					}
				},
			},
			{
				errorMessageRegex: `the version of the instance must be "9\.6" \(got "foo"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					v := v1alpha1.PostgresqlInstanceSpecVersion("foo")
					instance.Spec.Version = &v
				},
			},
		}
		for _, test := range tests {
			var (
				err error
				obj *v1alpha1.PostgresqlInstance
			)
			// Create a PostgresqlInstance resource and make sure that the expected error message has been returned.
			obj = &v1alpha1.PostgresqlInstance{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: framework.PostgresqlInstanceMetadataNamePrefix,
				},
				Spec: v1alpha1.PostgresqlInstanceSpec{
					Name: f.NewRandomPostgresqlInstanceSpecName(),
					Networking: &v1alpha1.PostgresqlInstanceSpecNetworking{
						PublicIP: &v1alpha1.PostgresqlInstanceSpecNetworkingPublicIP{
							Enabled: pointers.NewBool(true),
						},
					},
					Paused: true,
				},
			}
			test.fn(obj)
			_, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Create(obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(test.errorMessageRegex))
		}
	})

	framework.AdmissionIt("cannot be updated with invalid values", func() {
		var (
			obj *v1alpha1.PostgresqlInstance
			err error
		)

		// Create a PostgresqlInstance resource with random values for ".metadata.name" and ".spec.name", and having private IP and public IP networking enabled.
		obj, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Create(&v1alpha1.PostgresqlInstance{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: framework.PostgresqlInstanceMetadataNamePrefix,
			},
			Spec: v1alpha1.PostgresqlInstanceSpec{
				Name: f.NewRandomPostgresqlInstanceSpecName(),
				Location: &v1alpha1.PostgresqlInstanceSpecLocation{
					Region: pointers.NewString("europe-west4"),
				},
				Networking: &v1alpha1.PostgresqlInstanceSpecNetworking{
					PrivateIP: &v1alpha1.PostgresqlInstanceSpecNetworkingPrivateIP{
						Enabled: pointers.NewBool(true),
						Network: pointers.NewString("projects/" + f.ProjectId + "/global/networks/default"),
					},
					PublicIP: &v1alpha1.PostgresqlInstanceSpecNetworkingPublicIP{
						Enabled: pointers.NewBool(true),
					},
				},
				Resources: &v1alpha1.PostgresqlInstanceSpecResources{
					Disk: &v1alpha1.PostgresqlInstanceSpecResourcesDisk{
						SizeMaximumGb: pointers.NewInt32(0),
						SizeMinimumGb: pointers.NewInt32(20),
					},
				},
				Paused: true,
			},
		})
		Expect(err).NotTo(HaveOccurred())

		tests := []struct {
			errorMessageRegex string
			fn                func(instance *v1alpha1.PostgresqlInstance)
		}{
			{
				errorMessageRegex: `the region where the instance is located cannot be changed \(had "europe-west4", got "us-central-1"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Location.Region = pointers.NewString("us-central-1")
				},
			},
			{
				errorMessageRegex: `the name of the instance cannot be changed \(had ".*", got "new-name"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Name = "new-name"
				},
			},
			{
				errorMessageRegex: `private ip access to the instance cannot be disabled after having been enabled`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Networking.PrivateIP.Enabled = pointers.NewBool(false)
				},
			},
			{
				errorMessageRegex: `the resource link of the vpc network for the instance cannot be removed`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Networking.PrivateIP.Network = pointers.NewString("")
				},
			},
			{
				errorMessageRegex: `the minimum disk size for the instance cannot be decreased \(had "20", got "19"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					instance.Spec.Resources.Disk.SizeMinimumGb = pointers.NewInt32(*obj.Spec.Resources.Disk.SizeMinimumGb - 1)
				},
			},
			{
				errorMessageRegex: `the disk type for the instance cannot be changed \(had "SSD", got "HDD"\)`,
				fn: func(instance *v1alpha1.PostgresqlInstance) {
					newSpecResourcesDiskType := v1alpha1.PostgresqlInstanceSpecResourceDiskTypeHDD
					instance.Spec.Resources.Disk.Type = &newSpecResourcesDiskType
				},
			},
		}

		// Create a clone of the original PostgresqlInstance resource so we can perform the required changes on a fresh, valid source.
		// Then, do apply the required changes and make sure that the expected error message is returned.
		for _, test := range tests {
			updatedObj := obj.DeepCopy()
			test.fn(updatedObj)
			_, err = f.SelfClient.CloudsqlV1alpha1().PostgresqlInstances().Update(updatedObj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(test.errorMessageRegex))
		}

		// Delete the PostgresqlInstance resource.
		err = f.DeletePostgresqlInstanceByName(obj.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})
