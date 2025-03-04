/*
Copyright 2021 Red Hat, Inc.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// DisplayName refers to the name that an application will be deployed with in App Studio.
	DisplayName string `json:"displayName,omitempty"`

	// AppModelRepository refers to the git repository that will store the application model (a devfile)
	// Can be the same as GitOps repository.
	// A repository will be generated if this field is left blank.
	AppModelRepository ApplicationGitRepository `json:"appModelRepository,omitempty"`

	// GitOpsRepository refers to the git repository that will store the gitops resources.
	// Can be the same as App Model Repository.
	// A repository will be generated if this field is left blank.
	GitOpsRepository ApplicationGitRepository `json:"gitOpsRepository,omitempty"`

	// Description refers to a brief description of the application.
	Description string `json:"description,omitempty"`
}

// ApplicationGitRepository defines a git repository for a given Application resource (either appmodel or gitops)
type ApplicationGitRepository struct {
	// URL refers to the repository URL that should be used.
	// +required
	URL string `json:"url"`

	// Branch corresponds to the branch in the repository that should be used
	Branch string `json:"branch,omitempty"`

	// Context corresponds to the context within the repository that should be used
	Context string `json:"context,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions"`

	// Devfile corresponds to the devfile representation of the Application resource
	Devfile string `json:"devfile,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Application is the Schema for the applications API
// +kubebuilder:resource:path=applications,shortName=hasapp;ha;app
// +kubebuilder:subresource:status
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
