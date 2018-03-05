// Copyright (c) 2017 Northwestern Mutual.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ApiKey describes an ApiKey
type ApiKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApiKeySpec `json:"spec"`
}

// APIKeySpec is the spec for an ApiKey
type ApiKeySpec struct {
	Revisions []Revision `json:"revisions"`
}

type RevisionStatus string

const (
	RevisionStatusActive   RevisionStatus = "Active"
	RevisionStatusInactive RevisionStatus = "Inactive"
)

// ApiKeyRevision is an ApiKey revision
type Revision struct {
	Data     string         `json:"data"`
	Status   RevisionStatus `json:"status"`
	LastUsed string         `json:"lastUsed"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ApiKeyList is a list of ApiKey resources
type ApiKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiKey `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// MockTarget describes a MockTarget
type MockTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MockTargetSpec `json:"spec"`
}

// MockTargetSpec is the spec for a MockTarget
type MockTargetSpec struct {
	Routes []Route `json:"routes"`
}

// Route defines the behavior of an http request for a unique route
type Route struct {
	Path       string            `json:"path"`
	StatusCode int               `json:"status"`
	Methods    []string          `json:"methods,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// MockTargetList is a list of MockTarget resources
type MockTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MockTarget `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ApiProxy describe an ApiProxy
type ApiProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApiProxySpec `json:"spec"`
}

// ApiProxySpec is the spec for an ApiProxy
type ApiProxySpec struct {
	Source  Source   `json:"source"`
	Target  Target   `json:"target"`
	Plugins []Plugin `json:"plugins,omitempty"`
}

// Source represents an incoming request
type Source struct {
	Path        string `json:"path"`
	VirtualHost string `json:"virtualHost,omitempty"`
}

// Target describes a target proxy
type Target struct {
	Path    string  `json:"path,omitempty"`
	Backend Backend `json:"backend,omitempty"`
	SSL     *SSL    `json:"ssl,omitempty"`
}

// Mock describes a valid mock response
type Mock struct {
	MockTargetName string `json:"mockTargetName"`
}

// Backend describes an upstream server
type Backend struct {
	Endpoint *string  `json:"endpoint,omitempty"`
	Mock     *Mock    `json:"mock,omitempty"`
	Service  *Service `json:"service,omitempty"`
}

// Service describes a Kubernetes service
type Service struct {
	Name   string  `json:"name,omitempty"`
	Port   int64   `json:"port"`
	Labels []Label `json:"labels,omitempty"`
}

// Label defines a unique label to be matched against service metadata
type Label struct {
	Name   string `json:"name"`
	Header string `json:"header,omitempty"`
	Value  string `json:"value,omitempty"`
}

// SSL describes the anoatomy of a backend TLS connection
type SSL struct {
	SecretName string `json:"secretName"`
}

// Plugin describes a unique plugin to be envoked
type Plugin struct {
	Name    string            `json:"name"`
	Version string            `json:"version,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ApiProxyList is a list of ApiProxy resources
type ApiProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiProxy `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ApiKeyBinding describes an ApiKeyBinding
type ApiKeyBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApiKeyBindingSpec `json:"spec"`
}

// ApiKeyBindingSpec is the spec for an ApiKeyBinding
type ApiKeyBindingSpec struct {
	Keys []Key `json:"keys"`
}

// Key defines a unique key with permissions
type Key struct {
	Name        string `json:"name"`
	Rate        Rate   `json:"rate,omitempty"`
	DefaultRule Rule   `json:"defaultRule,omitempty"`
	Subpaths    []Path `json:"subpaths,omitempty"`
}

// Rate defines rate limit rule
type Rate struct {
	Amount int    `json:"amount,omitempty"`
	Unit   string `json:"unit,omitempty"`
}

// Rule defines a single permission
type Rule struct {
	Global   bool          `json:"global,omitempty"`
	Granular GranularProxy `json:"granular,omitempty"`
}

// Path describes a subpath with unique permissions
type Path struct {
	Path string `json:"path"`
	Rule Rule   `json:"rule,omitempty"`
}

// GranularProxy defines a list of authorized HTTP verbs
type GranularProxy struct {
	Verbs []string `json:"verbs"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ApiKeyBindingList represents a list of ApiKeyBinding resources
type ApiKeyBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiKeyBinding `json:"items"`
}
