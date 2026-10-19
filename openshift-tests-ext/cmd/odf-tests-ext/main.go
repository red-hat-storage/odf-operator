package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/red-hat-storage/odf-operator/openshift-tests-ext/pkg/adapter"
	"github.com/red-hat-storage/odf-operator/openshift-tests-ext/scripts"
	"github.com/red-hat-storage/odf-operator/openshift-tests-ext/test/odf"
)

func main() {
	defaultTimeout := 5 * time.Minute

	registry := e.NewRegistry()
	ext := e.NewExtension("odf", "external", "odf-tests-ext")

	ext.AddSuite(e.Suite{
		Name:             "odf/all",
		Description:      "All ODF tests",
		Qualifiers:       []string{`tags["component"] == "odf"`},
		ClusterStability: e.ClusterStabilityStable,
		Parallelism:      1,
		TestTimeout:      &defaultTimeout,
	})

	ext.AddSuite(e.Suite{
		Name:             "odf/smoke",
		Description:      "ODF smoke tests",
		Qualifiers:       []string{`labels.exists(l, l=="SMOKE")`},
		ClusterStability: e.ClusterStabilityStable,
		Parallelism:      1,
		TestTimeout:      &defaultTimeout,
	})

	specs, err := adapter.BuildExtensionTestSpecsFromScenarios(odf.ODFScenarios, scripts.FS)
	if err != nil {
		panic(fmt.Sprintf("couldn't build extension test specs from ODF scenarios: %+v", err.Error()))
	}

	ext.AddSpecs(specs)
	registry.Register(ext)

	root := &cobra.Command{
		Use:   "odf-tests-ext",
		Short: "ODF Tests Extension for OpenShift",
		Long:  "OpenShift Tests Extension (OTE) binary for ODF testing",
	}

	root.AddCommand(cmd.DefaultExtensionCommands(registry)...)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
