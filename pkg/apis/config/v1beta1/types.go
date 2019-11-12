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

// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Rules",type=integer,JSONPath=`.status.recordingRules`
// +kubebuilder:printcolumn:name="Alerts",type=integer,JSONPath=`.status.alertRules`
// +kubebuilder:printcolumn:name="Errors",type=integer,JSONPath=`.status.errorCount`
// RuleGroup
type RuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuleGroupSpec   `json:"spec"`
	Status RuleGroupStatus `json:"status,omitempty"`
}

// RuleGroupSpec is the spec for a rule group resource
type RuleGroupSpec struct {
	Interval string `json:"interval,omitempty"`
	Rules    []Rule `json:"rules"`
}

// RuleGroupStatus is the status for a rule group resource
type RuleGroupStatus struct {
	RecordingRuleCount int      `json:"recordingRules"`
	AlertRuleCount     int      `json:"alertRules"`
	ErrorCount         int      `json:"errorCount"`
	Errors             []string `json:"errors,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleGroupList is a list of rule group resources
type RuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RuleGroup `json:"items"`
}

// ScrapeSpec is the spec for a scrape resource
type ScrapeSpec string

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Errors",type=integer,JSONPath=`.status.errorCount`
// Scrape
type Scrape struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScrapeSpec   `json:"spec"`
	Status ScrapeStatus `json:"status,omitempty"`
}

// ScrapeStatus is the status for a scrape resource
type ScrapeStatus struct {
	ErrorCount int      `json:"errorCount"`
	Errors     []string `json:"errors,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScrapeList is a list of Foo resources
type ScrapeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Scrape `json:"items"`
}
