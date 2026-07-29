package console

import (
	"testing"

	consolev1 "github.com/openshift/api/console/v1"
)

func TestMergeConsolePluginProxy_PreservesUnknownEntries(t *testing.T) {
	ns := "openshift-storage"
	extra := consolev1.ConsolePluginProxy{
		Alias: "internalRgwS3",
		Endpoint: consolev1.ConsolePluginProxyEndpoint{
			Type: consolev1.ProxyTypeService,
			Service: &consolev1.ConsolePluginProxyServiceConfig{
				Name:      "rook-ceph-rgw-ocs-storagecluster-cephobjectstore",
				Namespace: ns,
				Port:      443,
			},
		},
		Authorization: consolev1.None,
	}

	existing := append(GetConsolePluginProxy(ns), extra)
	merged := MergeConsolePluginProxy(existing, ns, nil)

	found := false
	for _, p := range merged {
		if p.Alias == "internalRgwS3" {
			found = true
			break
		}
	}
	if !found {
		t.Error("MergeConsolePluginProxy dropped the internalRgwS3 entry added by another controller")
	}

	desired := GetConsolePluginProxy(ns)
	if len(merged) != len(desired)+1 {
		t.Errorf("expected %d proxy entries, got %d", len(desired)+1, len(merged))
	}
}

func TestMergeConsolePluginProxy_UpdatesKnownEntries(t *testing.T) {
	ns := "openshift-storage"
	stale := []consolev1.ConsolePluginProxy{
		{
			Alias: "provider-proxy",
			Endpoint: consolev1.ConsolePluginProxyEndpoint{
				Type: consolev1.ProxyTypeService,
				Service: &consolev1.ConsolePluginProxyServiceConfig{
					Name:      "old-name",
					Namespace: ns,
					Port:      1234,
				},
			},
			Authorization: consolev1.None,
		},
	}

	merged := MergeConsolePluginProxy(stale, ns, nil)

	for _, p := range merged {
		if p.Alias == "provider-proxy" {
			if p.Endpoint.Service.Name != "ux-backend-proxy" || p.Endpoint.Service.Port != 8888 {
				t.Errorf("expected provider-proxy to be updated, got service=%s port=%d",
					p.Endpoint.Service.Name, p.Endpoint.Service.Port)
			}
			return
		}
	}
	t.Error("provider-proxy not found in merged result")
}

func TestMergeConsolePluginProxy_EmptyExisting(t *testing.T) {
	ns := "openshift-storage"
	merged := MergeConsolePluginProxy(nil, ns, nil)
	desired := GetConsolePluginProxy(ns)

	if len(merged) != len(desired) {
		t.Errorf("expected %d proxy entries from empty existing, got %d", len(desired), len(merged))
	}
}

func TestMergeConsolePluginProxy_RemovesAliases(t *testing.T) {
	ns := "openshift-storage"
	extra := consolev1.ConsolePluginProxy{
		Alias: "internalRgwS3",
		Endpoint: consolev1.ConsolePluginProxyEndpoint{
			Type: consolev1.ProxyTypeService,
			Service: &consolev1.ConsolePluginProxyServiceConfig{
				Name:      "rook-ceph-rgw-ocs-storagecluster-cephobjectstore",
				Namespace: ns,
				Port:      443,
			},
		},
		Authorization: consolev1.None,
	}

	existing := append(GetConsolePluginProxy(ns), extra)
	merged := MergeConsolePluginProxy(existing, ns, []string{"internalRgwS3"})

	for _, p := range merged {
		if p.Alias == "internalRgwS3" {
			t.Error("MergeConsolePluginProxy should remove entries matching removeAliases")
		}
	}

	desired := GetConsolePluginProxy(ns)
	if len(merged) != len(desired) {
		t.Errorf("expected %d proxy entries after removal, got %d", len(desired), len(merged))
	}
}
