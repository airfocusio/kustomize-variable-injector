package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/airfocusio/kustomize-variable-injector/internal"
	"github.com/spf13/cobra"
)

type rootCmd struct {
	cmd *cobra.Command
}

func newRootCmd(version FullVersion) *rootCmd {
	result := &rootCmd{}
	cmd := &cobra.Command{
		Version:      version.Version,
		Use:          "kustomize-variable-injector",
		Short:        "An converter from helm charts to kustomizations",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := internal.Run()
			if err != nil {
				os.Exit(1)
				return
			}
		},
	}

	result.cmd = cmd
	return result
}

func Execute(version FullVersion) error {
	rootCmd := newRootCmd(version)
	return rootCmd.cmd.Execute()
}

type FullVersion struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

func (v FullVersion) ToString() string {
	result := v.Version
	if v.Commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, v.Commit)
	}
	if v.Date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, v.Date)
	}
	if v.BuiltBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, v.BuiltBy)
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}
