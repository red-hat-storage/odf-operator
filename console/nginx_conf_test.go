/*
Copyright 2021 Red Hat OpenShift Data Foundation.

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

package console

import (
	"strings"
	"testing"

	ocstlsv1 "github.com/red-hat-storage/ocs-tls-profiles/api/v1"
)

func TestGenerateNginxConf_NilOpenSSLConfig(t *testing.T) {
	conf := GenerateNginxConf(nil)
	if strings.Contains(conf, "ssl_protocols") {
		t.Error("expected no ssl_protocols directive when OpenSSLConfig is nil")
	}
	if strings.Contains(conf, "ssl_ciphers") {
		t.Error("expected no ssl_ciphers directive when OpenSSLConfig is nil")
	}
	if strings.Contains(conf, "ssl_conf_command") {
		t.Error("expected no ssl_conf_command directive when OpenSSLConfig is nil")
	}
	if !strings.Contains(conf, "listen       9001 ssl") {
		t.Error("expected server block with ssl listener")
	}
}

func TestGenerateNginxConf_TLS13(t *testing.T) {
	ossl := &ocstlsv1.OpenSSLConfig{
		Protocol: "TLSv1.3",
		Ciphers:  []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
		Groups:   []string{"X25519MLKEM768", "prime256v1"},
	}

	conf := GenerateNginxConf(ossl)
	if !strings.Contains(conf, "ssl_protocols TLSv1.3;") {
		t.Error("expected ssl_protocols TLSv1.3 directive")
	}
	if !strings.Contains(conf, "ssl_conf_command Ciphersuites TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384;") {
		t.Error("expected ssl_conf_command Ciphersuites directive for TLS 1.3")
	}
	if strings.Contains(conf, "ssl_ciphers") {
		t.Error("TLS 1.3 should not use ssl_ciphers directive")
	}
	if !strings.Contains(conf, "ssl_conf_command Groups X25519MLKEM768:prime256v1;") {
		t.Error("expected ssl_conf_command Groups directive with colon-separated groups")
	}
}

func TestGenerateNginxConf_TLS12(t *testing.T) {
	ossl := &ocstlsv1.OpenSSLConfig{
		Protocol: "TLSv1.2",
		Ciphers:  []string{"ECDHE-RSA-AES128-GCM-SHA256"},
		Groups:   []string{"prime256v1"},
	}

	conf := GenerateNginxConf(ossl)
	if !strings.Contains(conf, "ssl_protocols TLSv1.2;") {
		t.Error("expected ssl_protocols TLSv1.2 directive")
	}
	if !strings.Contains(conf, "ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256;") {
		t.Error("expected ssl_ciphers directive")
	}
	if !strings.Contains(conf, "ssl_conf_command Groups prime256v1;") {
		t.Error("expected ssl_conf_command Groups directive")
	}
}

func TestGenerateNginxConf_EmptyCiphersAndGroups(t *testing.T) {
	ossl := &ocstlsv1.OpenSSLConfig{
		Protocol: "TLSv1.3",
	}

	conf := GenerateNginxConf(ossl)
	if !strings.Contains(conf, "ssl_protocols TLSv1.3;") {
		t.Error("expected ssl_protocols directive even with empty ciphers/groups")
	}
	if strings.Contains(conf, "Ciphersuites") {
		t.Error("expected no Ciphersuites directive when ciphers list is empty")
	}
	if strings.Contains(conf, "ssl_ciphers") {
		t.Error("expected no ssl_ciphers directive when ciphers list is empty")
	}
	if strings.Contains(conf, "Groups") {
		t.Error("expected no Groups directive when groups list is empty")
	}
}

func TestGenerateNginxConf_PreservesBaseConfig(t *testing.T) {
	ossl := &ocstlsv1.OpenSSLConfig{
		Protocol: "TLSv1.3",
		Ciphers:  []string{"TLS_AES_128_GCM_SHA256"},
		Groups:   []string{"prime256v1"},
	}

	conf := GenerateNginxConf(ossl)
	for _, expected := range []string{
		"worker_processes auto;",
		"ssl_certificate /var/serving-cert/tls.crt;",
		"ssl_certificate_key /var/serving-cert/tls.key;",
		"location /compatibility/",
		"keepalive_timeout   65;",
	} {
		if !strings.Contains(conf, expected) {
			t.Errorf("expected base config to contain %q", expected)
		}
	}
}
