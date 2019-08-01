package pkg

import (
	"path/filepath"

	"github.com/appscode/go/flags"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	"stash.appscode.dev/stash/pkg/restic"
	"stash.appscode.dev/stash/pkg/util"
)

func NewCmdRestore() *cobra.Command {
	var (
		masterURL      string
		kubeconfigPath string
		namespace      string
		appBindingName string
		outputDir      string
		pgArgs         string
		setupOpt       = restic.SetupOptions{
			ScratchDir:  restic.DefaultScratchDir,
			EnableCache: false,
		}
		dumpOpt = restic.DumpOptions{
			Host:     restic.DefaultHost,
			FileName: PgDumpFile,
		}
		metrics = restic.MetricsOptions{
			JobName: JobPGBackup,
		}
	)

	cmd := &cobra.Command{
		Use:               "restore-pg",
		Short:             "Restores Postgres DB Backup",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.EnsureRequiredFlags(cmd, "app-binding", "provider", "secret-dir")

			// apply nice, ionice settings from env
			var err error
			setupOpt.Nice, err = util.NiceSettingsFromEnv()
			if err != nil {
				return util.HandleResticError(outputDir, restic.DefaultOutputFileName, err)
			}
			setupOpt.IONice, err = util.IONiceSettingsFromEnv()
			if err != nil {
				return util.HandleResticError(outputDir, restic.DefaultOutputFileName, err)
			}

			// prepare client
			config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
			if err != nil {
				return err
			}
			kubeClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}
			appCatalogClient, err := appcatalog_cs.NewForConfig(config)
			if err != nil {
				return err
			}

			// get app binding
			appBinding, err := appCatalogClient.AppcatalogV1alpha1().AppBindings(namespace).Get(appBindingName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			// get secret
			appBindingSecret, err := kubeClient.CoreV1().Secrets(namespace).Get(appBinding.Spec.Secret.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			// init restic wrapper
			resticWrapper, err := restic.NewResticWrapper(setupOpt)
			if err != nil {
				return err
			}

			// set env for psql
			resticWrapper.SetEnv(EnvPgPassword, string(appBindingSecret.Data[PostgresPassword]))
			// setup pipe command
			dumpOpt.StdoutPipeCommand = restic.Command{
				Name: PgRestoreCMD,
				Args: []interface{}{
					"-U", string(appBindingSecret.Data[PostgresUser]),
					"-h", appBinding.Spec.ClientConfig.Service.Name,
				},
			}
			if pgArgs != "" {
				dumpOpt.StdoutPipeCommand.Args = append(dumpOpt.StdoutPipeCommand.Args, pgArgs)
			}

			// wait for DB ready
			waitForDBReady(appBinding.Spec.ClientConfig.Service.Name, appBinding.Spec.ClientConfig.Service.Port)

			// Run dump
			dumpOutput, backupErr := resticWrapper.Dump(dumpOpt)
			// If metrics are enabled then generate metrics
			if metrics.Enabled {
				err := dumpOutput.HandleMetrics(&metrics, backupErr)
				if err != nil {
					return errors.NewAggregate([]error{backupErr, err})
				}
			}
			// If output directory specified, then write the output in "output.json" file in the specified directory
			if backupErr == nil && outputDir != "" {
				err := dumpOutput.WriteOutput(filepath.Join(outputDir, restic.DefaultOutputFileName))
				if err != nil {
					return err
				}
			}
			return backupErr
		},
	}

	cmd.Flags().StringVar(&pgArgs, "pg-args", pgArgs, "Additional arguments")

	cmd.Flags().StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace of Backup/Restore Session")
	cmd.Flags().StringVar(&appBindingName, "app-binding", appBindingName, "Name of the app binding")

	cmd.Flags().StringVar(&setupOpt.Provider, "provider", setupOpt.Provider, "Backend provider (i.e. gcs, s3, azure etc)")
	cmd.Flags().StringVar(&setupOpt.Bucket, "bucket", setupOpt.Bucket, "Name of the cloud bucket/container (keep empty for local backend)")
	cmd.Flags().StringVar(&setupOpt.Endpoint, "endpoint", setupOpt.Endpoint, "Endpoint for s3/s3 compatible backend")
	cmd.Flags().StringVar(&setupOpt.URL, "rest-server-url", setupOpt.URL, "URL for rest backend")
	cmd.Flags().StringVar(&setupOpt.Path, "path", setupOpt.Path, "Directory inside the bucket where backup will be stored")
	cmd.Flags().StringVar(&setupOpt.SecretDir, "secret-dir", setupOpt.SecretDir, "Directory where storage secret has been mounted")
	cmd.Flags().StringVar(&setupOpt.ScratchDir, "scratch-dir", setupOpt.ScratchDir, "Temporary directory")
	cmd.Flags().BoolVar(&setupOpt.EnableCache, "enable-cache", setupOpt.EnableCache, "Specify whether to enable caching for restic")
	cmd.Flags().IntVar(&setupOpt.MaxConnections, "max-connections", setupOpt.MaxConnections, "Specify maximum concurrent connections for GCS, Azure and B2 backend")

	cmd.Flags().StringVar(&dumpOpt.Host, "hostname", dumpOpt.Host, "Name of the host machine")
	cmd.Flags().StringVar(&dumpOpt.SourceHost, "source-hostname", dumpOpt.SourceHost, "Name of the host whose data will be restored")
	// TODO: sliceVar
	cmd.Flags().StringVar(&dumpOpt.Snapshot, "snapshot", dumpOpt.Snapshot, "Snapshot to dump")

	cmd.Flags().StringVar(&outputDir, "output-dir", outputDir, "Directory where output.json file will be written (keep empty if you don't need to write output in file)")

	cmd.Flags().BoolVar(&metrics.Enabled, "metrics-enabled", metrics.Enabled, "Specify whether to export Prometheus metrics")
	cmd.Flags().StringVar(&metrics.PushgatewayURL, "metrics-pushgateway-url", metrics.PushgatewayURL, "Pushgateway URL where the metrics will be pushed")
	cmd.Flags().StringVar(&metrics.MetricFileDir, "metrics-dir", metrics.MetricFileDir, "Directory where to write metric.prom file (keep empty if you don't want to write metric in a text file)")
	cmd.Flags().StringSliceVar(&metrics.Labels, "metrics-labels", metrics.Labels, "Labels to apply in exported metrics")

	return cmd
}
