/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"flag"

	"stash.appscode.dev/apimachinery/client/clientset/versioned/scheme"

	"github.com/spf13/cobra"
	"gomodules.xyz/x/flags"
	v "gomodules.xyz/x/version"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/logs"
	"kmodules.xyz/client-go/tools/cli"
)

var SupportedProducts = []string{"stash-enterprise"}

var licenseApiService string

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
	rootCmd.PersistentFlags().StringVar(&licenseApiService, "license-apiservice", "", "Name of ApiService used to expose License endpoint")
	rootCmd.PersistentFlags().BoolVar(&cli.EnableAnalytics, "enable-analytics", cli.EnableAnalytics, "Send analytical events to Google Analytics")

	rootCmd.AddCommand(v.NewCmdVersion())
	rootCmd.AddCommand(NewCmdBackup())
	rootCmd.AddCommand(NewCmdRestore())

	return rootCmd
}
