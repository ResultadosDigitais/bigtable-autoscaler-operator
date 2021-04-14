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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ServiceAccountSecretRef struct {
	// +kubebuilder:validation:MinLength=1
	Name *string `json:"name"`

	Namespace *string `json:"namespace,omitempty"`

	// +kubebuilder:validation:MinLength=1
	Key *string `json:"key"`
}

// BigtableAutoscalerSpec defines the desired state of BigtableAutoscaler
type BigtableAutoscalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Minimum=1
	// lower limit for the number of nodes that can be set by the autoscaler.
	MinNodes *int32 `json:"minNodes"`

	// +kubebuilder:validation:Minimum=1
	// upper limit for the number of nodes that can be set by the autoscaler.
	// It cannot be smaller than MinNodes.
	MaxNodes *int32 `json:"maxNodes"`

	// +kubebuilder:default:=2
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	// upper limit for the number of nodes when autoscaler scaledown.
	MaxScaleDownNodes *int32 `json:"maxScaleDownNodes"`

	// target average CPU utilization for Bigtable.
	TargetCPUUtilization *int32 `json:"targetCPUUtilization"`

	// reference to the bigtable cluster to be autoscaled
	BigtableClusterRef BigtableClusterRef `json:"bigtableClusterRef"`

	// reference to the service account to be used to get bigtable metrics
	ServiceAccountSecretRef ServiceAccountSecretRef `json:"serviceAccountSecretRef"`
}

// BigtableAutoscalerStatus defines the observed state of BigtableAutoscaler
type BigtableAutoscalerStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty"`
	LastFetchTime *metav1.Time `json:"lastFetchTime,omitempty"`

	// +kubebuilder:default:=0
	DesiredNodes *int32 `json:"desiredNodes,omitempty"`

	// +kubebuilder:default:=0
	CurrentNodes *int32 `json:"currentNodes,omitempty"`

	// +kubebuilder:default:=0
	CurrentCPUUtilization *int32 `json:"CPUUtilization,omitempty"`
}

type BigtableClusterRef struct {
	// Important: Run "make" to regenerate code after modifying this file

	ProjectID  string `json:"projectId,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
	ClusterID  string `json:"clusterId,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="nodes",type=string,JSONPath=`.status.currentNodes`
// +kubebuilder:printcolumn:name="desired_nodes",type=string,JSONPath=`.status.desiredNodes`
// +kubebuilder:printcolumn:name="cpu_usage",type=string,JSONPath=`.status.CPUUtilization`
// +kubebuilder:printcolumn:name="target_cpu",type=string,JSONPath=`.spec.targetCPUUtilization`
// +kubebuilder:subresource:status

// BigtableAutoscaler is the Schema for the bigtableautoscalers API
type BigtableAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BigtableAutoscalerSpec   `json:"spec,omitempty"`
	Status BigtableAutoscalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BigtableAutoscalerList contains a list of BigtableAutoscaler
type BigtableAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BigtableAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BigtableAutoscaler{}, &BigtableAutoscalerList{})
}
