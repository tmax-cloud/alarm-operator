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

type NotificationType string

const (
	NotificationTypeMail    NotificationType = "Email"
	NotificationTypeWebhook NotificationType = "Webhook"
	NotificationTypeSlack   NotificationType = "Slack"
	NotificationTypeUnknown NotificationType = "Unknown"
)

type EmailNotification struct {
	SMTPConfig string `json:"smtpcfg"`
	From       string `json:"from"`
	To         string `json:"to"`
	Cc         string `json:"cc,omitempty"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
}

type WebhookNotification struct {
	Url     string `json:"url"`
	Message string `json:"message"`
}

type SlackNotification struct {
	SenderAccountSecret string `json:"account"`
	Workspace           string `json:"workspace"`
	Channel             string `json:"channel"`
	Message             string `json:"message"`
}

// NotificationSpec defines the desired state of Notification
type NotificationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:OneOf
	Email EmailNotification `json:"email,omitempty"`
	// +kubebuilder:validation:OneOf
	Webhook WebhookNotification `json:"webhook,omitempty"`
	// +kubebuilder:validation:OneOf
	Slack SlackNotification `json:"slack,omitempty"`
}

// NotificationStatus defines the observed state of Notification
type NotificationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Type NotificationType `json:"type,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=not
// +kubebuilder:printcolumn:name="Action",type=string,JSONPath=`.status.type`

// Notification is the Schema for the notifications API
type Notification struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotificationSpec   `json:"spec,omitempty"`
	Status NotificationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotificationList contains a list of Notification
type NotificationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Notification `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Notification{}, &NotificationList{})
}
