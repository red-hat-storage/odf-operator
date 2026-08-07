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

// TLS versions, ciphers and groups are sourced from https://ssl-config.mozilla.org/guidelines/5.7.json
package v1

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
)

// selectorRe validates the two valid selector forms:
//
//	wildcard branch: \*
//	  \*                          - literal '*'
//
//	exact branch:   (([a-zA-Z0-9][-]?)*[a-zA-Z0-9])
//	  ([a-zA-Z0-9][-]?)*         - zero or more alnum chars each optionally followed by a hyphen
//	  [a-zA-Z0-9]                - must end with alnum (no trailing hyphen)
//
//	both branches may be followed by:
//	  (\.([a-zA-Z0-9][-]?)*[a-zA-Z0-9])*  - zero or more dot-separated labels (same alnum/hyphen rule)
//	  (\/[a-zA-Z0-9_-]+)?                 - optional server e.g. "/s3"
//
// Rejected examples: "" (empty), "/s3" (leading slash), "example.com/" (empty server),
// "a/b/c" (two slashes), "-example.com" (label starts with hyphen), "example-.com" (label ends with hyphen).
var selectorRe = regexp.MustCompile(`^(\*|(([a-zA-Z0-9][-]?)*[a-zA-Z0-9]))(\.([a-zA-Z0-9][-]?)*[a-zA-Z0-9])*(\/[a-zA-Z0-9_-]+)?$`)

// versionToGoID returns the Go crypto/tls constant for version.
func versionToGoID(version TLSProtocolVersion) (uint16, bool) {
	switch version {
	case VersionTLS1_2:
		return tls.VersionTLS12, true
	case VersionTLS1_3:
		return tls.VersionTLS13, true
	default:
		return 0, false
	}
}

// goVersionToOpenSSL returns the openssl string for version.
func goVersionToOpenSSL(version uint16) (string, bool) {
	switch version {
	case tls.VersionTLS12:
		return "TLSv1.2", true
	case tls.VersionTLS13:
		return "TLSv1.3", true
	default:
		return "", false
	}
}

// cipherToGoID returns the Go crypto/tls cipher suite ID for an IANA cipher name and tls version.
func cipherToGoID(version TLSProtocolVersion, cipher TLSCipherSuite) (uint16, bool) {
	if version == VersionTLS1_2 {
		switch cipher {
		case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
			return tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, true
		case "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":
			return tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, true
		case "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256":
			return tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, true
		case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
			return tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, true
		case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
			return tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, true
		case "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":
			return tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256, true
		}
	} else if version == VersionTLS1_3 {
		switch cipher {
		case "TLS_AES_128_GCM_SHA256":
			return tls.TLS_AES_128_GCM_SHA256, true
		case "TLS_AES_256_GCM_SHA384":
			return tls.TLS_AES_256_GCM_SHA384, true
		case "TLS_CHACHA20_POLY1305_SHA256":
			return tls.TLS_CHACHA20_POLY1305_SHA256, true
		}
	}
	return 0, false
}

func goCipherToOpenSSL(id uint16) (string, bool) {
	switch id {
	// tls 1.2
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "ECDHE-ECDSA-AES128-GCM-SHA256", true
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "ECDHE-ECDSA-AES256-GCM-SHA384", true
	case tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:
		return "ECDHE-ECDSA-CHACHA20-POLY1305", true
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "ECDHE-RSA-AES128-GCM-SHA256", true
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "ECDHE-RSA-AES256-GCM-SHA384", true
	case tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
		return "ECDHE-RSA-CHACHA20-POLY1305", true
	// tls 1.3
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256", true
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384", true
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256", true
	default:
		return "", false
	}
}

// groupToGoID returns the Go crypto/tls CurveID for an IANA/IETF group name and tls version.
func groupToGoID(version TLSProtocolVersion, group TLSGroupName) (tls.CurveID, bool) {
	// classic groups are supported in both 1.2 and 1.3
	if version == VersionTLS1_2 || version == VersionTLS1_3 {
		switch group {
		case "secp256r1":
			return tls.CurveP256, true
		case "secp384r1":
			return tls.CurveP384, true
		case "secp521r1":
			return tls.CurveP521, true
		case "X25519":
			return tls.X25519, true
		}
	}

	// hybrid groups are supported only in 1.3
	if version == VersionTLS1_3 {
		switch group {
		case "X25519MLKEM768":
			return tls.X25519MLKEM768, true
		case "SecP256r1MLKEM768":
			return tls.CurveID(4587), true // tls.SecP256r1MLKEM768; constant requires Go 1.26
		case "SecP384r1MLKEM1024":
			return tls.CurveID(4589), true // tls.SecP384r1MLKEM1024; constant requires Go 1.26
		}
	}

	return 0, false
}

func goCurveToOpenSSL(id tls.CurveID) (string, bool) {
	switch id {
	// classic
	case tls.CurveP256:
		return "prime256v1", true
	case tls.CurveP384:
		return "secp384r1", true
	case tls.CurveP521:
		return "secp521r1", true
	case tls.X25519:
		return "x25519", true
	// hybrids
	case tls.X25519MLKEM768:
		return "X25519MLKEM768", true
	case tls.CurveID(4587): // SecP256r1MLKEM768
		return "SecP256r1MLKEM768", true
	case tls.CurveID(4589): // SecP384r1MLKEM1024
		return "SecP384r1MLKEM1024", true
	default:
		return "", false
	}
}

func findSpecificity(selector, domain, server string) int {
	if !selectorRe.MatchString(selector) {
		return -1
	}
	parts := strings.SplitN(selector, "/", 2)
	specificity := -1

	// The length of the domain is always multiplied by 2 (for domain or sub-domain) to make sure that the
	// contribution of the domain part will always be even. This ensures that the contribution of the
	// sever name in the selector will result in an odd number, guaranteeing that the same specificity cannot
	// be the result of 2 different selectors.
	if suffix, ok := strings.CutPrefix(parts[0], "*"); ok {
		if strings.HasSuffix(domain, suffix) {
			specificity = 2 * len(suffix)
		}
	} else if parts[0] == domain {
		specificity = 2 * len(suffix)
	}

	// match on the server only if a domain is already matched
	if specificity > -1 && len(parts) > 1 {
		if parts[1] == server {
			specificity += 1
		} else {
			specificity = -1
		}
	}
	return specificity
}

// GetConfigForServer returns the TLS config that best matches (domain, server) across rules.
//
// Resolution order (most specific wins):
//  1. "domain.io/server" - match a server named "server" under domain "domain.io"
//  2. "domain.io"        - match all servers under "domain.io"
//  3. "*.io/server"      - match a server named "server" under any domain that ends with ".io"
//  4. "*.io"             - match all servers under any domain that ends with ".io"
//  5. "*/server"         - match a server named "server" under any domain
//  6. "*"                - matches everything
//
// Returns nil, false if no rule matches or if multiple rules match at the same specificity level (ambiguous).
func GetConfigForServer(tlsProfile *TLSProfile, domain, server string) (*TLSConfig, bool) {
	if tlsProfile == nil {
		return nil, false
	}

	rules := tlsProfile.Spec.Rules
	var matchedConfig *TLSConfig
	storedSpecificity := -1

	for i, rule := range rules {
		for _, selector := range rule.Selectors {
			incomingSpecificity := findSpecificity(string(selector), domain, server)
			if incomingSpecificity > storedSpecificity {
				storedSpecificity = incomingSpecificity
				matchedConfig = &rules[i].Config
			} else if incomingSpecificity == storedSpecificity &&
				incomingSpecificity > -1 &&
				matchedConfig != &rules[i].Config {
				return nil, false
			}
		}
	}

	return matchedConfig, storedSpecificity >= 0
}

// ValidateTLSConfig validates that ciphers and groups are compatible with the specified TLS version.
// CRD-level enum markers ensure cipher/group names and protocol version are valid;
// this validates cross-field compatibility only.
//
// TLS 1.2: only TLS 1.2 ciphers and classical groups are permitted.
// TLS 1.3: only TLS 1.3 ciphers and classical or hybrid groups are permitted.
func ValidateTLSConfig(src *TLSConfig) error {
	if src == nil {
		return fmt.Errorf("nil config supplied")
	}

	for _, cipher := range src.Ciphers {
		if _, ok := cipherToGoID(src.Version, cipher); !ok {
			return fmt.Errorf("cipher %q is not valid for %s", cipher, src.Version)
		}
	}

	for _, group := range src.Groups {
		if _, ok := groupToGoID(src.Version, group); !ok {
			return fmt.Errorf("group %q is not valid for %s", group, src.Version)
		}
	}

	return nil
}

// GetGoTLSConfig converts a validated TLSConfig to a Go crypto/tls Config.
// Call ValidateTLSConfig before this to ensure src is valid.
// Version is exact, not a range: both MinVersion and MaxVersion are set to the same value.
func GetGoTLSConfig(src *TLSConfig) *tls.Config {
	cfg := &tls.Config{}
	if v, ok := versionToGoID(src.Version); ok {
		cfg.MinVersion = v
		cfg.MaxVersion = v
	}

	for _, cipher := range src.Ciphers {
		if id, ok := cipherToGoID(src.Version, cipher); ok {
			cfg.CipherSuites = append(cfg.CipherSuites, id)
		}
	}

	for _, group := range src.Groups {
		if id, ok := groupToGoID(src.Version, group); ok {
			cfg.CurvePreferences = append(cfg.CurvePreferences, id)
		}
	}

	return cfg
}

// OpenSSLConfig holds translated OpenSSL TLS values.
// The fields are shared across OpenSSL-based servers (Nginx, HAProxy, Apache, etc.).
// +kubebuilder:object:generate=false
type OpenSSLConfig struct {
	Protocol string   // e.g. "TLSv1.2"
	Ciphers  []string // e.g. "ECDHE-RSA-AES128-GCM-SHA256 ECDHE-RSA-AES256-GCM-SHA384"
	Groups   []string // e.g. "prime256v1 x25519"
}

// OpenSSLConfigFrom translates a Go tls.Config into OpenSSL configuration values.
// Only MinVersion, CipherSuites, and CurvePreferences are read.
func OpenSSLConfigFrom(src *tls.Config) *OpenSSLConfig {
	if src == nil {
		return nil
	}

	protocol, ok := goVersionToOpenSSL(src.MinVersion)
	if !ok {
		return nil
	}
	result := &OpenSSLConfig{}
	result.Protocol = protocol

	if len(src.CipherSuites) > 0 {
		for _, id := range src.CipherSuites {
			if name, ok := goCipherToOpenSSL(id); ok {
				result.Ciphers = append(result.Ciphers, name)
			}
		}
	}

	if len(src.CurvePreferences) > 0 {
		for _, id := range src.CurvePreferences {
			if name, ok := goCurveToOpenSSL(id); ok {
				result.Groups = append(result.Groups, name)
			}
		}
	}

	return result
}
