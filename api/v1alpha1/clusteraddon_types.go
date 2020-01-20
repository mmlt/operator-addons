/*

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

// Important: Action "make" to regenerate code after modifying this file.

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterAddonSpec defines the desired state of a target k8s cluster.
type ClusterAddonSpec struct {
	// Specifies what cluster addon operations are allowed.
	// Valid values are:
	// - "AllowAll" (default): allows create, update and delete of cluster add-ons;
	// - "DenyDelete": forbids delete of cluster add-ons when ClusterAddon resource is deleted;
	// - "DenyUpdate": forbids update/delete of cluster add-ons when ClusterAddon or repo changes.
	// +optional
	Policy ClusterAddonPolicy `json:"policy,omitempty"`

	// Target is the k8s cluster that will get updated by this controller.
	Target ClusterAddonTarget `json:"target,omitempty"`

	// Sources is the map of repositories and run actions to perform on the target k8s cluster.
	Sources map[string]ClusterAddonSource `json:"sources,omitempty"`
}

// ClusterAddonPolicy describes how the cluster addons will be updated or deleted.
// Only one of the following policies may be specified.
// If none is specified, the default one AllowAll.
// +kubebuilder:validation:Enum=AllowAll;DenyDelete;DenyUpdate
type ClusterAddonPolicy string

const (
	// AllowAll allows create, update and delete of cluster add-ons.
	AllowAll ClusterAddonPolicy = "AllowAll"

	// DenyDelete forbids delete of cluster add-ons when ClusterAddon resource is deleted.
	DenyDelete ClusterAddonPolicy = "DenyDelete"

	// DenyUpdate forbids update/delete of cluster add-ons when ClusterAddon or repo changes.
	DenyUpdate ClusterAddonPolicy = "DenyUpdate"
)

type ClusterAddonTarget struct {
	// URL is the URL of the API Server.
	URL string `json:"url"`
	// CACert is the CA of the API Server base64 encoded.
	CACert []byte `json:"caCert"`
	// User is the username (used together with password) to authenticate.
	// +optional
	User string `json:"user,omitempty"`
	// Password is the user password base64 encoded.
	// +optional
	Password string `json:"password,omitempty"`
	// ClientCert is the certificate (used together with ClientKey) to authenticate.
	// +optional
	ClientCert []byte `json:"ClientCert,omitempty"`
	// ClientKey is the ClientCert key base64 encoded.
	// +optional
	ClientKey []byte `json:"ClientKey,omitempty"`
}

type ClusterAddonSource struct {
	// Type is the type of repository to use as a source.
	// Valid values are:
	// - "git" (default): GIT repository.
	// +optional
	Type ClusterAddonSourceType `json:"type,omitempty"`

	// +kubebuilder:validation:MinLength=2

	// URL is the URL of the repo that is available at $REPOROOT during the Action.
	// When Token is specified the URL is expected to start with 'https://'.
	URL string `json:"url"`

	// +kubebuilder:validation:MinLength=2

	// Branch is the repo branch to get.
	Branch string `json:"branch"`

	// Token is used to authenticate with the remote server.
	// For Type=git;
	// - Token or ~/.ssh key should be specified (azure devops requires the token to be prefixed with 'x:')
	// +optional
	Token string `json:"token,omitempty"`

	// Action specifies what to do when the content of the repository changes.
	Action ClusterAddonAction `json:"action"`
}

// ClusterAddonSourceType is the type of repository to use as a source.
// Valid values are:
// - SourceTypeGIT (default)
// +kubebuilder:validation:Enum=git
type ClusterAddonSourceType string

const (
	// SourceTypeGIT specifies a source repository of type GIT.
	SourceTypeGIT ClusterAddonSourceType = "git"
)

type ClusterAddonAction struct {
	// Type is the type of action to perform when the repository has changed.
	// Valid values are:
	// - "shell" (default): Action shell with 'cmd' and 'values'.
	// +optional
	Type ClusterAddonActionType `json:"type,omitempty"`

	// +kubebuilder:validation:MinLength=2

	// Cmd specifies what command to run in the shell.
	Cmd string `json:"cmd"`

	// Values are key-value pairs that are passed as values.yaml and environment
	// variables to the shell.
	// +optional
	Values map[string]string `json:"values,omitempty"`
}

// ClusterAddonActionType is the type of action to run when the repository has changed.
// Valid values are:
// - RunTypeShell (default)
// +kubebuilder:validation:Enum=shell
type ClusterAddonActionType string

const (
	// RunTypeShell specifies that a bash shell will be used.
	RunTypeShell ClusterAddonActionType = "shell"
)

// ClusterAddonStatus defines the observed state of a clusteraddon.
type ClusterAddonStatus struct {
	// Conditions are the latest available observations of an object's current state.
	// +optional
	Conditions []ClusterAddonCondition `json:"conditions,omitempty"`

	// Synced is true when the source/action have been applied successfully.
	// +optional
	Synced metav1.ConditionStatus `json:"synced,omitempty"`
}

type ClusterAddonConditionType string

// These are valid conditions of a clusteraddon.
const (
	// ClusterAddonTargetOk means the target cluster is OK.
	ClusterAddonTargetOk ClusterAddonConditionType = "TargetOk"
	// ClusterAddonTargetOk means the source repo is OK.
	ClusterAddonSourceOk ClusterAddonConditionType = "SourceOk"
	// ClusterAddonTargetOk means the action is performed successfully.
	ClusterAddonActionOk ClusterAddonConditionType = "ActionOk"
	// ClusterAddonSynced means the source/action have been applied successfully.
	ClusterAddonSynced ClusterAddonConditionType = "Synced"
)

// ClusterAddonCondition is one of;
// - TargetOk
// - SourceOk
// - ActionOk
// - Synced
type ClusterAddonCondition struct {
	// Type of clusteraddon condition, Complete or Failed.
	Type ClusterAddonConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status metav1.ConditionStatus `json:"status,omitempty"`

	// Last time the condition status has changed.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason for last transition in a single word.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.target.url`
// +kubebuilder:printcolumn:name="Synced",type=string,JSONPath=`.status.synced`
// +kubebuilder:subresource:status

// ClusterAddon is the Schema for the clusteraddons API
type ClusterAddon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterAddonSpec   `json:"spec,omitempty"`
	Status ClusterAddonStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterAddonList contains a list of ClusterAddon
type ClusterAddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAddon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterAddon{}, &ClusterAddonList{})
}
