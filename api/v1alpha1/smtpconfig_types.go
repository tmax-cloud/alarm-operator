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

// SMTPConfigSpec defines the desired state of SMTPConfig
type SMTPConfigSpec struct {
	Host             string `json:"host"`
	Port             int    `json:"port"`
	CredentialSecret string `json:"secret"`
}

// SMTPConfigStatus defines the observed state of SMTPConfig
type SMTPConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.host`
// +kubebuilder:printcolumn:name="Port",type=integer,JSONPath=`.spec.port`
// +kubebuilder:printcolumn:name="Secret",type=string,JSONPath=`.spec.secret`

// SMTPConfig is the Schema for the smtpconfigs API
type SMTPConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SMTPConfigSpec   `json:"spec,omitempty"`
	Status SMTPConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SMTPConfigList contains a list of SMTPConfig
type SMTPConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SMTPConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SMTPConfig{}, &SMTPConfigList{})
}
