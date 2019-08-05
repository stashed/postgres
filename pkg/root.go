package pkg

import (
	"flag"

	"github.com/appscode/go/flags"
	v "github.com/appscode/go/version"
	"github.com/spf13/cobra"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/logs"
	"kmodules.xyz/client-go/tools/cli"
	"stash.appscode.dev/stash/apis"
	"stash.appscode.dev/stash/client/clientset/versioned/scheme"
	"stash.appscode.dev/stash/pkg/util"
)

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:               "stash-postgres",
		Short:             `PostgreSQL backup & restore plugin for Stash by AppsCode`,
		Long:              `PostgreSQL backup & restore plugin for Stash by AppsCode. For more information, visit here: https://appscode.com/products/stash`,
		DisableAutoGenTag: true,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			flags.DumpAll(c.Flags())
			cli.SendAnalytics(c, v.Version.Version)

			return scheme.AddToScheme(clientsetscheme.Scheme)
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	logs.ParseFlags()
	rootCmd.PersistentFlags().StringVar(&util.ServiceName, "service-name", "stash-operator", "Stash service name.")
	rootCmd.PersistentFlags().BoolVar(&cli.EnableAnalytics, "enable-analytics", cli.EnableAnalytics, "Send analytical events to Google Analytics")
	rootCmd.PersistentFlags().BoolVar(&apis.EnableStatusSubresource, "enable-status-subresource", apis.EnableStatusSubresource, "If true, uses sub resource for crds.")

	rootCmd.AddCommand(v.NewCmdVersion())
	rootCmd.AddCommand(NewCmdBackup())
	rootCmd.AddCommand(NewCmdRestore())

	return rootCmd
}
