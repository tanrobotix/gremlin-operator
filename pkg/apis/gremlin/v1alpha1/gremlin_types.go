package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GremlinSpec defines the desired state of Gremlin
// +k8s:openapi-gen=true
type GremlinSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	TeamID string `json:"team_id"`
	// +kubebuilder:validation:Enum=attack-container,attack
	Type string `json:"type"`
	// +kubebuilder:validation:Enum=cpu,disk,dns,io,latency,memory,packet_loss,process_killer,shutdown
	Gremlin string `json:"gremlin"`
	Length  uint   `json:"length,omitempty"`

	// CPU attack
	Cores int `json:"cores"`
	// Disk & IO attack
	BlockSize  uint   `json:"block_size,omitempty"`
	BlockCount uint   `json:"block_count,omitempty"`
	Dir        string `json:"dir,omitempty"`
	Percent    uint   `json:"percent,omitempty"`
	Workers    uint   `json:"workers,omitempty"`
	// +kubebuilder:validation:Enum=r,w,rw
	Mode string `json:"mode,omitempty"`

	// DNS, Latency  and packet_loss attack
	Device     string `json:"device,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
	IPProtocol string `json:"ip_protocol,omitempty"`
	Ms         string `json:"ms,omitempty"`
	EgressPort string `json:"egress_port,omitempty"`
	SrcPort    string `json:"src_port,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Corrupt    bool   `json:"corrupt,omitempty"`

	// Memory attack
	GigaBytes uint `json:"gigabytes,omitempty"`
	MegaBytes uint `json:"megabytes,omitempty"`

	// Process Killer
	Interval     uint   `json:"interval,omitempty"`
	Process      string `json:"process,omitempty"`
	Signal       int    `json:"signal,omitempty"`
	Group        string `json:"group,omitempty"`
	User         string `json:"user,omitempty"`
	Newest       bool   `json:"newest,omitempty"`
	Oldest       bool   `json:"oldest,omitempty"`
	Exact        bool   `json:"exact,omitempty"`
	KillChildren bool   `json:"kill_children,omitempty"`
	Full         bool   `json:"full,omitempty"`

	// Shutdown attack
	Delay  uint `json:"delay,omitempty"`
	Reboot bool `json:"reboot,omitempty"`
}

// GremlinStatus defines the observed state of Gremlin
// +k8s:openapi-gen=true
type GremlinStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Gremlin is the Schema for the gremlins API
// +k8s:openapi-gen=true
type Gremlin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GremlinSpec   `json:"spec,omitempty"`
	Status GremlinStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GremlinList contains a list of Gremlin
type GremlinList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gremlin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gremlin{}, &GremlinList{})
}
