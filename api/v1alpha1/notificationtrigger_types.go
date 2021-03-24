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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type NotificationTriggerResult struct {
	Triggered bool   `json:"triggered"`
	Message   string `json:"message,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// NotificationTriggerSpec defines the desired state of NotificationTrigger
type NotificationTriggerSpec struct {
	Notification string `json:"notification"`
	Monitor      string `json:"monitor"`
	FieldPath    string `json:"fieldPath"`
	Op           string `json:"op"`
	Operand      string `json:"operand"`
}

// NotificationTriggerStatus defines the observed state of NotificationTrigger
type NotificationTriggerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	History []NotificationTriggerResult `json:"history,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ntr

// NotificationTrigger is the Schema for the notificationtriggers API
type NotificationTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotificationTriggerSpec   `json:"spec,omitempty"`
	Status NotificationTriggerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotificationTriggerList contains a list of NotificationTrigger
type NotificationTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotificationTrigger `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NotificationTrigger{}, &NotificationTriggerList{})
}
