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

const (
	// ReasonConflict is the reason used in conditions and events that indicate that a conflict was found while updating a CSQLP instance.
	ReasonConflict = "Conflict"
	// ReasonInstanceCreated is the reason used in conditions and events that indicate that a CSQLP instance has been created.
	ReasonInstanceCreated = "InstanceCreated"
	// ReasonInstanceNotReady is the reason used in conditions and events that indicate that a CSQLP instance is not ready.
	ReasonInstanceNotReady = "InstanceNotReady"
	// v is the reason used in conditions and events that indicate that a CSQLP instance is ready.
	ReasonInstanceReady = "InstanceReady"
	// ReasonInstanceUpdated is the reason used in conditions and events that indicate that a CSQLP instance has been updated.
	ReasonInstanceUpdated = "InstanceUpdated"
	// ReasonInstanceUpToDate is the reason used in conditions and events that indicate that a CSQLP instance is up-to-date.
	ReasonInstanceUpToDate = "InstanceUpToDate"
	// ReasonInvalidSpec is the reason used in conditions and events that indicate that an invalid specification was provided for a CSQLP instance.
	ReasonInvalidSpec = "InvalidSpec"
	// ReasonNameUnavailable is the reason used in conditions and events that indicate that the chosen name for a CSQLP instance is unavailable.
	ReasonNameUnavailable = "NameUnavailable"
	// ReasonOperationInProgress is the reason used in conditions and events that indicate that an operation is still in progress for a CSQLP instance.
	ReasonOperationInProgress = "OperationInProgress"
	// ReasonUnexpectedError is the reason used in conditions and events that indicate that an unexpected error occurred while managing a CSQLP instance.
	ReasonUnexpectedError = "UnexpectedError"
)
