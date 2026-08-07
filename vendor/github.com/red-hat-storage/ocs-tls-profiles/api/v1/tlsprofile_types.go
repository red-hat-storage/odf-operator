/*
Copyright 2026 Red Hat OpenShift Data Foundation.

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

// TLSProtocolVersion represents a TLS protocol version.
// Only TLS 1.2 and 1.3 are supported as TLS 1.0 and 1.1 are considered vulnerable.
// +kubebuilder:validation:Enum=TLSv1.2;TLSv1.3
type TLSProtocolVersion string

const (
	// VersionTLS1_2 is version 1.2 of the TLS security protocol.
	VersionTLS1_2 TLSProtocolVersion = "TLSv1.2"
	// VersionTLS1_3 is version 1.3 of the TLS security protocol.
	VersionTLS1_3 TLSProtocolVersion = "TLSv1.3"
)

// TLSCipherSuite is an IANA TLS cipher suite name.
// TLS 1.2 ciphers are configurable in both Go and OpenSSL servers.
// TLS 1.3 ciphers are configurable in OpenSSL only; Go selects TLS 1.3 ciphers automatically.
// Note: ChaCha20-Poly1305 ciphers are not FIPS 140-2 approved - do not use them on FIPS-enabled clusters.
// +kubebuilder:validation:Enum=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256;TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256;TLS_AES_128_GCM_SHA256;TLS_AES_256_GCM_SHA384;TLS_CHACHA20_POLY1305_SHA256
type TLSCipherSuite string

// TLSGroupName is a TLS key exchange group name.
// Classical groups (secp256r1, secp384r1, secp521r1, X25519) are valid for TLS 1.2 and TLS 1.3.
// Hybrid post-quantum groups (X25519MLKEM768, SecP256r1MLKEM768, SecP384r1MLKEM1024) are valid for TLS 1.3 only.
// Note: Hybrid post-quantum groups are not FIPS 140-2 approved - do not use them on FIPS-enabled clusters.
// +kubebuilder:validation:Enum=secp256r1;secp384r1;secp521r1;X25519;X25519MLKEM768;SecP256r1MLKEM768;SecP384r1MLKEM1024
type TLSGroupName string

// TLSConfig defines the TLS security configuration for a component.
// Ensure ciphers and groups are compatible with the specified version:
// TLS 1.2 ciphers and classical groups only for TLSv1.2; TLS 1.3 ciphers and any group for TLSv1.3.
// +kubebuilder:validation:XValidation:rule="self.version != 'TLSv1.2' || !self.ciphers.exists(c, c in ['TLS_AES_128_GCM_SHA256','TLS_AES_256_GCM_SHA384','TLS_CHACHA20_POLY1305_SHA256'])",message="TLS 1.3 ciphers are not valid for TLSv1.2"
// +kubebuilder:validation:XValidation:rule="self.version != 'TLSv1.3' || !self.ciphers.exists(c, c in ['TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256','TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384','TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256','TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256','TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384','TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256'])",message="TLS 1.2 ciphers are not valid for TLSv1.3"
// +kubebuilder:validation:XValidation:rule="self.version != 'TLSv1.2' || !self.groups.exists(g, g in ['X25519MLKEM768','SecP256r1MLKEM768','SecP384r1MLKEM1024'])",message="hybrid post-quantum groups are not valid for TLSv1.2"
type TLSConfig struct {
	// Version specifies the minimum and maximum TLS protocol version.
	// +kubebuilder:validation:Required
	Version TLSProtocolVersion `json:"version"`

	// Ciphers is the list of IANA cipher suite names to enable, in preference order.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=20
	// +listType=set
	Ciphers []TLSCipherSuite `json:"ciphers"`

	// Groups is the list of key exchange groups to enable, in preference order.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=20
	// +listType=set
	Groups []TLSGroupName `json:"groups"`
}

// Selector is a selector pattern identifying which components a TLS rule applies to.
// Resolution order (most specific wins):
//  1. "<domain>/<server>"    - match a server under exact domain
//  2. "<domain>"             - match all servers under exact domain
//  3. "*.<domain>/<server>"  - match a server under domain
//  4. "*.<domain>"           - match all servers under domain
//  5. "*/<server>"           - match a server under any domain
//  6. "*"                    - matches everything
//
// +kubebuilder:validation:Pattern=`^(\*|(([a-zA-Z0-9][-]?)*[a-zA-Z0-9]))(\.([a-zA-Z0-9][-]?)*[a-zA-Z0-9])*(\/[a-zA-Z0-9_-]+)?$`
type Selector string

// TLSProfileRules pairs a selector with a TLS configuration.
type TLSProfileRules struct {
	// Selectors identifies the components this rule applies to.
	// Examples:
	//   "example.io/s3"      - Matches the S3 server under example.io
	//   "example.io"         - Matches all servers under example.io
	//   "*.example.io/s3"    - Matches the S3 server under any subdomain of example.io
	//   "*.example.io"       - Matches all servers under any subdomain of example.io
	//   "*/s3"               - Matches the S3 server under any domain
	//   "*"                  - Matches all servers under any domain
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=20
	Selectors []Selector `json:"selectors,omitzero"`

	// Config is the TLS configuration to apply to the selected components.
	Config TLSConfig `json:"config,omitzero"`
}

// TLSProfileSpec defines the desired state of TLSProfile.
type TLSProfileSpec struct {
	// Rules is a list of TLS configuration rules.
	// When multiple rules match a component, the most specific selector wins.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=20
	Rules []TLSProfileRules `json:"rules,omitzero"`
}

type TLSProfileStatus struct{}

// TLSProfile is the Schema for the tlsprofiles API.
// It allows administrators to configure TLS settings (protocol versions, ciphers, groups)
// in a centralized way.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type TLSProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// Spec defines the desired TLS configuration rules.
	// WARNING: On FIPS-enabled clusters, restrict ciphers to AES-GCM variants and groups to
	// classical NIST curves (secp256r1, secp384r1, secp521r1). ChaCha20-Poly1305 ciphers and
	// hybrid post-quantum groups are not FIPS 140-2 approved. [Apr 2026]
	Spec TLSProfileSpec `json:"spec,omitzero"`
	// +optional
	Status TLSProfileStatus `json:"status,omitzero"`
}

//+kubebuilder:object:root=true

// TLSProfileList contains a list of TLSProfile resources.
type TLSProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []TLSProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TLSProfile{}, &TLSProfileList{})
}
