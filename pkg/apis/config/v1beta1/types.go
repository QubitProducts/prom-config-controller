package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Rule describes an alerting or recording rule.
type Rule struct {
	Record      string            `json:"record,omitempty"`
	Alert       string            `json:"alert,omitempty"`
	Expr        string            `json:"expr"`
	For         string            `json:"for,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleGroup
type RuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuleGroupSpec   `json:"spec"`
	Status RuleGroupStatus `json:"status,omitempty"`
}

// RuleGroupSpec is the spec for a Foo resource
type RuleGroupSpec struct {
	Interval string `json:"interval,omitempty"`
	Rules    []Rule `json:"rules"`
}

// RuleGroupStatus is the status for a Foo resource
type RuleGroupStatus struct {
	Errors []string `json:"errors,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleGroupList is a list of Foo resources
type RuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RuleGroup `json:"items"`
}

// ScrapeSpec is the spec for a Foo resource
type ScrapeSpec string

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Scrape
type Scrape struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScrapeSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScrapeList is a list of Foo resources
type ScrapeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Scrape `json:"items"`
}
