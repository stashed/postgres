/*
Copyright AppsCode Inc. and Contributors

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

package restic

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"stash.appscode.dev/apimachinery/apis"
	api_v1beta1 "stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	cs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/pkg/invoker"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

const (
	MetricsLabelDriver     = "driver"
	MetricsLabelKind       = "kind"
	MetricsLabelAppGroup   = "group"
	MetricsLabelName       = "name"
	MetricsLabelNamespace  = "namespace"
	MetricsLabelRepository = "repository"
	MetricsLabelBackend    = "backend"
	MetricsLabelBucket     = "bucket"
	MetricsLabelPrefix     = "prefix"
	MetricLabelInvokerKind = "invoker_kind"
	MetricLabelInvokerName = "invoker_name"
	MetricLabelHostname    = "hostname"
)

// BackupMetrics defines prometheus metrics for backup process
type BackupMetrics struct {
	// BackupSessionMetrics shows metrics related to entire backup session
	BackupSessionMetrics *BackupSessionMetrics
	// BackupTargetMetrics shows metrics related to a target
	BackupTargetMetrics *BackupTargetMetrics
	// BackupHostMetrics shows backup metrics for individual hosts
	BackupHostMetrics *BackupHostMetrics
}

// BackupSessionMetrics defines metrics for entire backup session
type BackupSessionMetrics struct {
	// SessionSuccess indicates whether the entire backup session was succeeded or not
	SessionSuccess prometheus.Gauge
	// SessionDuration indicates total time taken to complete the entire backup session
	SessionDuration prometheus.Gauge
	// TargetCount indicates the total number of targets that was backed up in this backup session
	TargetCount prometheus.Gauge
	// LastSuccessTime indicates the time(in unix epoch) when the last BackupSession was succeeded
	LastSuccessTime prometheus.Gauge
}

// BackupTargetMetrics defines metrics related to a target
type BackupTargetMetrics struct {
	// TargetBackupSucceeded indicates whether the backup for a target has succeeded or not
	TargetBackupSucceeded prometheus.Gauge
	// HostCount indicates the total number of hosts that was backed up for this target
	HostCount prometheus.Gauge
	// LastSuccessTime indicates the time (in unix epoch) when the last backup was successful for this target
	LastSuccessTime prometheus.Gauge
}

// BackupHostMetrics defines Prometheus metrics for individual hosts backup
type BackupHostMetrics struct {
	// BackupSuccess indicates whether the backup for a host succeeded or not
	BackupSuccess prometheus.Gauge
	// BackupDuration indicates total time taken to complete the backup process for a host
	BackupDuration prometheus.Gauge
	// DataSize indicates total size of the target data to backup for a host (in bytes)
	DataSize prometheus.Gauge
	// DataUploaded indicates the amount of data uploaded to the repository for a host (in bytes)
	DataUploaded prometheus.Gauge
	// DataProcessingTime indicates total time taken to backup the target data for a host
	DataProcessingTime prometheus.Gauge
	// FileMetrics shows information of backup files
	FileMetrics *FileMetrics
}

// FileMetrics defines Prometheus metrics for target files of a backup process for a host
type FileMetrics struct {
	// TotalFiles shows total number of files that has been backed up for a host
	TotalFiles prometheus.Gauge
	// NewFiles shows total number of new files that has been created since last backup for a host
	NewFiles prometheus.Gauge
	// ModifiedFiles shows total number of files that has been modified since last backup for a host
	ModifiedFiles prometheus.Gauge
	// UnmodifiedFiles shows total number of files that has not been changed since last backup for a host
	UnmodifiedFiles prometheus.Gauge
}

// RestoreMetrics defines metrics for the restore process
type RestoreMetrics struct {
	// RestoreSessionMetrics shows metrics related to entire restore session
	RestoreSessionMetrics *RestoreSessionMetrics
	// RestoreTargetMetrics shows metrics related to a restore target
	RestoreTargetMetrics *RestoreTargetMetrics
	// RestoreHostMetrics shows metrics related to the individual host of a restore target
	RestoreHostMetrics *RestoreHostMetrics
}

// RestoreSessionMetrics defines metrics related to entire restore session
type RestoreSessionMetrics struct {
	// SessionSuccess indicates whether the restore session succeeded or not
	SessionSuccess prometheus.Gauge
	// SessionDuration indicates the total time taken to complete the entire restore session
	SessionDuration prometheus.Gauge
	// TargetCount indicates the number of targets that was restored in this restore session
	TargetCount prometheus.Gauge
}

// RestoreTargetMetrics defines metrics related to a restore target
type RestoreTargetMetrics struct {
	// TargetRestoreSucceeded indicates whether the restore for a target has succeeded or not
	TargetRestoreSucceeded prometheus.Gauge
	// HostCount indicates the total number of hosts that was restored up for a restore target
	HostCount prometheus.Gauge
}

// RestoreHostMetrics defines restore metrics for the individual hosts
type RestoreHostMetrics struct {
	// RestoreSuccess indicates whether restore was succeeded or not for a host
	RestoreSuccess prometheus.Gauge
	// RestoreDuration indicates the time taken to complete the restore process for a host
	RestoreDuration prometheus.Gauge
}

// RepositoryMetrics defines Prometheus metrics for Repository state after each backup
type RepositoryMetrics struct {
	// RepoIntegrity shows result of repository integrity check after last backup
	RepoIntegrity prometheus.Gauge
	// RepoSize show size of repository after last backup
	RepoSize prometheus.Gauge
	// SnapshotCount shows number of snapshots stored in the repository
	SnapshotCount prometheus.Gauge
	// SnapshotsRemovedOnLastCleanup shows number of old snapshots cleaned up according to retention policy on last backup session
	SnapshotsRemovedOnLastCleanup prometheus.Gauge
}

func newBackupSessionMetrics(labels prometheus.Labels) *BackupMetrics {
	return &BackupMetrics{
		BackupSessionMetrics: &BackupSessionMetrics{
			SessionSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "session_success",
					Help:        "Indicates whether the entire backup session was succeeded or not",
					ConstLabels: labels,
				},
			),
			SessionDuration: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "session_duration_seconds",
					Help:        "Indicates total time taken to complete the entire backup session",
					ConstLabels: labels,
				},
			),
			TargetCount: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "target_count_total",
					Help:        "Indicates the total number of target that was backed up in this backup session",
					ConstLabels: labels,
				},
			),
			LastSuccessTime: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "last_success_time_seconds",
					Help:        "Indicates total time taken to complete the entire backup session",
					ConstLabels: labels,
				},
			),
		},
	}
}

func newBackupTargetMetrics(labels prometheus.Labels) *BackupMetrics {
	return &BackupMetrics{
		BackupTargetMetrics: &BackupTargetMetrics{
			TargetBackupSucceeded: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "target_success",
					Help:        "Indicates whether the backup for a target has succeeded or not",
					ConstLabels: labels,
				},
			),
			HostCount: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "target_host_count_total",
					Help:        "Indicates the total number of hosts that was backed up for this target",
					ConstLabels: labels,
				},
			),
			LastSuccessTime: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "target_last_success_time_seconds",
					Help:        "Indicates total time taken to complete the entire backup session",
					ConstLabels: labels,
				},
			),
		},
	}
}

func newBackupHostMetrics(labels prometheus.Labels) *BackupMetrics {
	return &BackupMetrics{
		BackupHostMetrics: &BackupHostMetrics{
			BackupSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "host_backup_success",
					Help:        "Indicates whether the backup for a host succeeded or not",
					ConstLabels: labels,
				},
			),
			BackupDuration: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "host_backup_duration_seconds",
					Help:        "Indicates total time taken to complete the backup process for a host",
					ConstLabels: labels,
				},
			),
			DataSize: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "host_data_size_bytes",
					Help:        "Total size of the target data to backup for a host (in bytes)",
					ConstLabels: labels,
				},
			),
			DataUploaded: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "host_data_uploaded_bytes",
					Help:        "Amount of data uploaded to the repository for a host (in bytes)",
					ConstLabels: labels,
				},
			),
			DataProcessingTime: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "backup",
					Name:        "host_data_processing_time_seconds",
					Help:        "Total time taken to process the target data for a host",
					ConstLabels: labels,
				},
			),
			FileMetrics: &FileMetrics{
				TotalFiles: prometheus.NewGauge(
					prometheus.GaugeOpts{
						Namespace:   "stash",
						Subsystem:   "backup",
						Name:        "host_files_total",
						Help:        "Total number of files that has been backed up for a host",
						ConstLabels: labels,
					},
				),
				NewFiles: prometheus.NewGauge(
					prometheus.GaugeOpts{
						Namespace:   "stash",
						Subsystem:   "backup",
						Name:        "host_files_new",
						Help:        "Total number of new files that has been created since last backup for a host",
						ConstLabels: labels,
					},
				),
				ModifiedFiles: prometheus.NewGauge(
					prometheus.GaugeOpts{
						Namespace:   "stash",
						Subsystem:   "backup",
						Name:        "host_files_modified",
						Help:        "Total number of files that has been modified since last backup for a host",
						ConstLabels: labels,
					},
				),
				UnmodifiedFiles: prometheus.NewGauge(
					prometheus.GaugeOpts{
						Namespace:   "stash",
						Subsystem:   "backup",
						Name:        "host_files_unmodified",
						Help:        "Total number of files that has not been changed since last backup for a host",
						ConstLabels: labels,
					},
				),
			},
		},
	}
}

func newRepositoryMetrics(labels prometheus.Labels) *RepositoryMetrics {
	return &RepositoryMetrics{
		RepoIntegrity: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "stash",
				Subsystem:   "repository",
				Name:        "integrity",
				Help:        "Result of repository integrity check after last backup",
				ConstLabels: labels,
			},
		),
		RepoSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "stash",
				Subsystem:   "repository",
				Name:        "size_bytes",
				Help:        "Indicates size of repository after last backup (in bytes)",
				ConstLabels: labels,
			},
		),
		SnapshotCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "stash",
				Subsystem:   "repository",
				Name:        "snapshot_count",
				Help:        "Indicates number of snapshots stored in the repository",
				ConstLabels: labels,
			},
		),
		SnapshotsRemovedOnLastCleanup: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "stash",
				Subsystem:   "repository",
				Name:        "snapshot_cleaned",
				Help:        "Indicates number of old snapshots cleaned up according to retention policy on last backup session",
				ConstLabels: labels,
			},
		),
	}
}

func newRestoreSessionMetrics(labels prometheus.Labels) *RestoreMetrics {
	return &RestoreMetrics{
		RestoreSessionMetrics: &RestoreSessionMetrics{
			SessionSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "restore",
					Name:        "session_success",
					Help:        "Indicates whether the entire restore session was succeeded or not",
					ConstLabels: labels,
				},
			),
			SessionDuration: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "restore",
					Name:        "session_duration_seconds",
					Help:        "Indicates the total time taken to complete the entire restore session",
					ConstLabels: labels,
				},
			),
			TargetCount: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "restore",
					Name:        "target_count_total",
					Help:        "Indicates the total number of targets that was restored in this restore session",
					ConstLabels: labels,
				},
			),
		},
	}
}

func newRestoreTargetMetrics(labels prometheus.Labels) *RestoreMetrics {
	return &RestoreMetrics{
		RestoreTargetMetrics: &RestoreTargetMetrics{
			TargetRestoreSucceeded: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "restore",
					Name:        "target_success",
					Help:        "Indicates whether the restore for a target has succeeded or not",
					ConstLabels: labels,
				},
			),
			HostCount: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "restore",
					Name:        "target_host_count_total",
					Help:        "Indicates the total number of hosts that was restored for this restore target",
					ConstLabels: labels,
				},
			),
		},
	}
}

func newRestoreHostMetrics(labels prometheus.Labels) *RestoreMetrics {
	return &RestoreMetrics{
		RestoreHostMetrics: &RestoreHostMetrics{
			RestoreSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "restore",
					Name:        "host_restore_success",
					Help:        "Indicates whether the restore process was succeeded for a host",
					ConstLabels: labels,
				},
			),
			RestoreDuration: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "stash",
					Subsystem:   "restore",
					Name:        "host_restore_duration_seconds",
					Help:        "Indicates the total time taken to complete the restore process for a host",
					ConstLabels: labels,
				},
			),
		},
	}
}

// SendBackupSessionMetrics send backup session related metrics to the Pushgateway
func (metricOpt *MetricsOptions) SendBackupSessionMetrics(inv invoker.BackupInvoker, status api_v1beta1.BackupSessionStatus) error {
	// create metric registry
	registry := prometheus.NewRegistry()

	// generate metrics labels
	labels, err := backupInvokerLabels(inv, metricOpt.Labels)
	if err != nil {
		return err
	}
	// create metrics
	metrics := newBackupSessionMetrics(labels)

	if status.Phase == api_v1beta1.BackupSessionSucceeded {
		// mark entire backup session as succeeded
		metrics.BackupSessionMetrics.SessionSuccess.Set(1)

		// set total time taken to complete the entire backup session
		duration, err := time.ParseDuration(status.SessionDuration)
		if err != nil {
			return err
		}
		metrics.BackupSessionMetrics.SessionDuration.Set(duration.Seconds())

		// set total number of target that was backed up in this backup session
		metrics.BackupSessionMetrics.TargetCount.Set(float64(len(status.Targets)))

		// set last successful session time to current time
		metrics.BackupSessionMetrics.LastSuccessTime.SetToCurrentTime()

		// register metrics to the registry
		registry.MustRegister(
			metrics.BackupSessionMetrics.SessionSuccess,
			metrics.BackupSessionMetrics.SessionDuration,
			metrics.BackupSessionMetrics.TargetCount,
			metrics.BackupSessionMetrics.LastSuccessTime,
		)
	} else {
		// mark entire backup session as failed
		metrics.BackupSessionMetrics.SessionSuccess.Set(0)
		registry.MustRegister(metrics.BackupSessionMetrics.SessionSuccess)
	}

	// send metrics to the pushgateway
	return metricOpt.sendMetrics(registry, metricOpt.JobName)
}

// SendBackupSessionMetrics send backup session metrics to the Pushgateway
func (metricOpt *MetricsOptions) SendBackupTargetMetrics(config *rest.Config, i invoker.BackupInvoker, targetRef api_v1beta1.TargetRef, status api_v1beta1.BackupSessionStatus) error {
	// create metric registry
	registry := prometheus.NewRegistry()

	// generate backup session related labels
	labels, err := backupInvokerLabels(i, metricOpt.Labels)
	if err != nil {
		return err
	}
	// generate target related labels
	targetLabels, err := targetLabels(config, targetRef, i.ObjectMeta.Namespace)
	if err != nil {
		return err
	}
	labels = upsertLabel(labels, targetLabels)

	// create metrics
	metrics := newBackupTargetMetrics(labels)

	// only send the metric for the target specified by targetRef
	for _, targetStatus := range status.Targets {
		if invoker.TargetMatched(targetStatus.Ref, targetRef) {
			if targetStatus.Phase == api_v1beta1.TargetBackupSucceeded {
				// mark target backup as succeeded
				metrics.BackupTargetMetrics.TargetBackupSucceeded.Set(1)

				// set last successful backup time for this target to current time
				metrics.BackupTargetMetrics.LastSuccessTime.SetToCurrentTime()

				// set total number of target that was backed up in this backup session
				if targetStatus.TotalHosts != nil {
					metrics.BackupTargetMetrics.HostCount.Set(float64(*targetStatus.TotalHosts))
				}

				// register metrics to the registry
				registry.MustRegister(
					metrics.BackupTargetMetrics.TargetBackupSucceeded,
					metrics.BackupTargetMetrics.LastSuccessTime,
					metrics.BackupTargetMetrics.HostCount,
				)
			} else {
				// mark target backup as failed
				metrics.BackupTargetMetrics.TargetBackupSucceeded.Set(0)
				registry.MustRegister(metrics.BackupTargetMetrics.TargetBackupSucceeded)
			}

			// send metrics to the pushgateway
			return metricOpt.sendMetrics(registry, metricOpt.JobName)
		}
	}
	return nil
}

// SendBackupSessionMetrics send backup metrics for individual hosts to the Pushgateway
func (metricOpt *MetricsOptions) SendBackupHostMetrics(config *rest.Config, i invoker.BackupInvoker, targetRef api_v1beta1.TargetRef, backupOutput *BackupOutput) error {
	if backupOutput == nil {
		return fmt.Errorf("invalid backup output. Backup output shouldn't be nil")
	}

	// create metric registry
	registry := prometheus.NewRegistry()

	// generate backup session related labels
	labels, err := backupInvokerLabels(i, metricOpt.Labels)
	if err != nil {
		return err
	}
	// generate target related labels
	targetLabels, err := targetLabels(config, targetRef, i.ObjectMeta.Namespace)
	if err != nil {
		return err
	}
	labels = upsertLabel(labels, targetLabels)

	// create metrics for the individual host
	for _, hostStats := range backupOutput.BackupTargetStatus.Stats {
		// add host name as label
		hostLabel := map[string]string{
			MetricLabelHostname: hostStats.Hostname,
		}
		metrics := newBackupHostMetrics(upsertLabel(labels, hostLabel))

		if hostStats.Error == "" {
			// set metrics values from backupOutput
			err := metrics.setValues(hostStats)
			if err != nil {
				return err
			}
			metrics.BackupHostMetrics.BackupSuccess.Set(1)

			registry.MustRegister(
				// register backup session metrics
				metrics.BackupHostMetrics.BackupSuccess,
				metrics.BackupHostMetrics.BackupDuration,
				metrics.BackupHostMetrics.FileMetrics.TotalFiles,
				metrics.BackupHostMetrics.FileMetrics.NewFiles,
				metrics.BackupHostMetrics.FileMetrics.ModifiedFiles,
				metrics.BackupHostMetrics.FileMetrics.UnmodifiedFiles,
				metrics.BackupHostMetrics.DataSize,
				metrics.BackupHostMetrics.DataUploaded,
				metrics.BackupHostMetrics.DataProcessingTime,
			)
		} else {
			metrics.BackupHostMetrics.BackupSuccess.Set(0)

			registry.MustRegister(
				// register backup session metrics
				metrics.BackupHostMetrics.BackupSuccess,
			)
		}
	}
	return metricOpt.sendMetrics(registry, metricOpt.JobName)
}

// SendRepositoryMetrics send backup session related metrics to the Pushgateway
func (metricOpt *MetricsOptions) SendRepositoryMetrics(config *rest.Config, i invoker.BackupInvoker, repoStats RepositoryStats) error {
	// create metric registry
	registry := prometheus.NewRegistry()

	// generate backup invoker labels
	labels, err := backupInvokerLabels(i, metricOpt.Labels)
	if err != nil {
		return err
	}

	repoMetricLabels, err := repoMetricLabels(config, i, metricOpt.Labels)
	if err != nil {
		return err
	}

	// create repository metrics
	repoMetrics := newRepositoryMetrics(upsertLabel(labels, repoMetricLabels))
	err = repoMetrics.setValues(repoStats)
	if err != nil {
		return err
	}
	// register repository metrics
	registry.MustRegister(
		repoMetrics.RepoIntegrity,
		repoMetrics.RepoSize,
		repoMetrics.SnapshotCount,
		repoMetrics.SnapshotsRemovedOnLastCleanup,
	)
	// send metrics to the pushgateway
	return metricOpt.sendMetrics(registry, metricOpt.JobName)
}

// SendRestoreSessionMetrics send restore session related metrics to the Pushgateway
func (metricOpt *MetricsOptions) SendRestoreSessionMetrics(inv invoker.RestoreInvoker) error {
	// create metric registry
	registry := prometheus.NewRegistry()

	// generate metrics labels
	labels, err := restoreInvokerLabels(inv, metricOpt.Labels)
	if err != nil {
		return err
	}
	// create metrics
	metrics := newRestoreSessionMetrics(labels)

	if inv.Status.Phase == api_v1beta1.RestoreSucceeded {
		// mark the entire restore session as succeeded
		metrics.RestoreSessionMetrics.SessionSuccess.Set(1)

		// set total time taken to complete the restore session
		duration, err := time.ParseDuration(inv.Status.SessionDuration)
		if err != nil {
			return err
		}
		metrics.RestoreSessionMetrics.SessionDuration.Set(duration.Seconds())

		// set total number of target that was restored in this restore session
		metrics.RestoreSessionMetrics.TargetCount.Set(float64(len(inv.Status.TargetStatus)))

		// register metrics to the registry
		registry.MustRegister(
			metrics.RestoreSessionMetrics.SessionSuccess,
			metrics.RestoreSessionMetrics.SessionDuration,
			metrics.RestoreSessionMetrics.TargetCount,
		)
	} else {
		// mark entire restore session as failed
		metrics.RestoreSessionMetrics.SessionSuccess.Set(0)
		registry.MustRegister(metrics.RestoreSessionMetrics.SessionSuccess)
	}

	// send metrics to the pushgateway
	return metricOpt.sendMetrics(registry, metricOpt.JobName)
}

// SendRestoreTargetMetrics send restore target related metrics to the Pushgateway
func (metricOpt *MetricsOptions) SendRestoreTargetMetrics(config *rest.Config, i invoker.RestoreInvoker, targetRef api_v1beta1.TargetRef) error {
	// create metric registry
	registry := prometheus.NewRegistry()

	// generate metrics labels
	labels, err := restoreInvokerLabels(i, metricOpt.Labels)
	if err != nil {
		return err
	}
	// generate target related labels
	targetLabels, err := targetLabels(config, targetRef, i.ObjectMeta.Namespace)
	if err != nil {
		return err
	}
	labels = upsertLabel(labels, targetLabels)

	// create metrics
	metrics := newRestoreTargetMetrics(labels)

	// only send the metric of the target specified by targetRef
	for _, targetStatus := range i.Status.TargetStatus {
		if invoker.TargetMatched(targetStatus.Ref, targetRef) {
			if targetStatus.Phase == api_v1beta1.TargetRestoreSucceeded {
				// mark entire restore target as succeeded
				metrics.RestoreTargetMetrics.TargetRestoreSucceeded.Set(1)

				// set total number of host that was restored in this restore session
				if targetStatus.TotalHosts != nil {
					metrics.RestoreTargetMetrics.HostCount.Set(float64(*targetStatus.TotalHosts))
				}

				// register metrics to the registry
				registry.MustRegister(
					metrics.RestoreTargetMetrics.TargetRestoreSucceeded,
					metrics.RestoreTargetMetrics.HostCount,
				)
			} else {
				// mark entire restore target as failed
				metrics.RestoreTargetMetrics.TargetRestoreSucceeded.Set(0)
				registry.MustRegister(metrics.RestoreTargetMetrics.TargetRestoreSucceeded)
			}

			// send metrics to the pushgateway
			return metricOpt.sendMetrics(registry, metricOpt.JobName)
		}
	}
	return nil
}

// SendRestoreHostMetrics send restore metrics for individual hosts to the Pushgateway
func (metricOpt *MetricsOptions) SendRestoreHostMetrics(config *rest.Config, i invoker.RestoreInvoker, targetRef api_v1beta1.TargetRef, restoreOutput *RestoreOutput) error {
	if restoreOutput == nil {
		return fmt.Errorf("invalid restore output. Restore output shouldn't be nil")
	}

	// create metric registry
	registry := prometheus.NewRegistry()

	// generate restore session related labels
	labels, err := restoreInvokerLabels(i, metricOpt.Labels)
	if err != nil {
		return err
	}
	// generate target related labels
	targetLabels, err := targetLabels(config, targetRef, i.ObjectMeta.Namespace)
	if err != nil {
		return err
	}
	labels = upsertLabel(labels, targetLabels)

	// create metrics for the individual host
	for _, hostStats := range restoreOutput.RestoreTargetStatus.Stats {
		// add host name as label
		hostLabel := map[string]string{
			MetricLabelHostname: hostStats.Hostname,
		}
		metrics := newRestoreHostMetrics(upsertLabel(labels, hostLabel))

		if hostStats.Error == "" {
			// mark the host restore as success
			metrics.RestoreHostMetrics.RestoreSuccess.Set(1)

			// set the time that has been taken to restore the host
			duration, err := time.ParseDuration(hostStats.Duration)
			if err != nil {
				return err
			}
			metrics.RestoreHostMetrics.RestoreDuration.Set(duration.Seconds())

			registry.MustRegister(
				metrics.RestoreHostMetrics.RestoreSuccess,
				metrics.RestoreHostMetrics.RestoreDuration,
			)
		} else {
			// mark the host restore as failure
			metrics.RestoreHostMetrics.RestoreSuccess.Set(0)
			registry.MustRegister(
				metrics.RestoreHostMetrics.RestoreSuccess,
			)
		}
	}

	return metricOpt.sendMetrics(registry, metricOpt.JobName)
}

func (backupMetrics *BackupMetrics) setValues(hostOutput api_v1beta1.HostBackupStats) error {
	var (
		totalDataSize        float64
		totalUploadSize      float64
		totalProcessingTime  uint64
		totalFiles           int64
		totalNewFiles        int64
		totalModifiedFiles   int64
		totalUnmodifiedFiles int64
	)

	for _, v := range hostOutput.Snapshots {
		dataSizeBytes, err := convertSizeToBytes(v.TotalSize)
		if err != nil {
			return err
		}
		totalDataSize = totalDataSize + dataSizeBytes

		uploadSizeBytes, err := convertSizeToBytes(v.Uploaded)
		if err != nil {
			return err
		}
		totalUploadSize = totalUploadSize + uploadSizeBytes

		processingTimeSeconds, err := convertTimeToSeconds(v.ProcessingTime)
		if err != nil {
			return err
		}
		totalProcessingTime = totalProcessingTime + processingTimeSeconds

		totalFiles = totalFiles + *v.FileStats.TotalFiles
		totalNewFiles = totalNewFiles + *v.FileStats.NewFiles
		totalModifiedFiles = totalModifiedFiles + *v.FileStats.ModifiedFiles
		totalUnmodifiedFiles = totalUnmodifiedFiles + *v.FileStats.UnmodifiedFiles
	}

	backupMetrics.BackupHostMetrics.DataSize.Set(totalDataSize)
	backupMetrics.BackupHostMetrics.DataUploaded.Set(totalUploadSize)
	backupMetrics.BackupHostMetrics.DataProcessingTime.Set(float64(totalProcessingTime))
	backupMetrics.BackupHostMetrics.FileMetrics.TotalFiles.Set(float64(totalFiles))
	backupMetrics.BackupHostMetrics.FileMetrics.NewFiles.Set(float64(totalNewFiles))
	backupMetrics.BackupHostMetrics.FileMetrics.ModifiedFiles.Set(float64(totalModifiedFiles))
	backupMetrics.BackupHostMetrics.FileMetrics.UnmodifiedFiles.Set(float64(totalUnmodifiedFiles))

	duration, err := time.ParseDuration(hostOutput.Duration)
	if err != nil {
		return err
	}
	backupMetrics.BackupHostMetrics.BackupDuration.Set(duration.Seconds())

	return nil
}

func (repoMetrics *RepositoryMetrics) setValues(repoStats RepositoryStats) error {
	// set repository metrics values
	if repoStats.Integrity != nil && *repoStats.Integrity {
		repoMetrics.RepoIntegrity.Set(1)
	} else {
		repoMetrics.RepoIntegrity.Set(0)
	}
	repoSize, err := convertSizeToBytes(repoStats.Size)
	if err != nil {
		return err
	}
	repoMetrics.RepoSize.Set(repoSize)
	repoMetrics.SnapshotCount.Set(float64(repoStats.SnapshotCount))
	repoMetrics.SnapshotsRemovedOnLastCleanup.Set(float64(repoStats.SnapshotsRemovedOnLastCleanup))

	return nil
}

func (metricOpt *MetricsOptions) sendMetrics(registry *prometheus.Registry, jobName string) error {
	// if Pushgateway URL is provided, then push the metrics to Pushgateway
	if metricOpt.PushgatewayURL != "" {
		pusher := push.New(metricOpt.PushgatewayURL, jobName)
		err := pusher.Gatherer(registry).Add()
		if err != nil {
			return err
		}
	}

	// if metric file directory is specified, then write the metrics in "metric.prom" text file in the specified directory
	if metricOpt.MetricFileDir != "" {
		err := prometheus.WriteToTextfile(filepath.Join(metricOpt.MetricFileDir, "metric.prom"), registry)
		if err != nil {
			return err
		}
	}
	return nil
}

// nolint:unparam
func backupInvokerLabels(inv invoker.BackupInvoker, userProvidedLabels []string) (prometheus.Labels, error) {
	// add user provided labels
	promLabels := parseUserProvidedLabels(userProvidedLabels)

	// add invoker information
	promLabels[MetricLabelInvokerKind] = inv.TypeMeta.Kind
	promLabels[MetricLabelInvokerName] = inv.ObjectMeta.Name
	promLabels[MetricsLabelNamespace] = inv.ObjectMeta.Namespace

	// insert target information as metrics label
	if inv.Driver == api_v1beta1.VolumeSnapshotter {
		promLabels = upsertLabel(promLabels, volumeSnapshotterLabels())
	} else {
		promLabels[MetricsLabelDriver] = string(api_v1beta1.ResticSnapshotter)
		promLabels[MetricsLabelRepository] = inv.Repository
	}

	return promLabels, nil
}

// nolint:unparam
func restoreInvokerLabels(inv invoker.RestoreInvoker, userProvidedLabels []string) (prometheus.Labels, error) {
	// add user provided labels
	promLabels := parseUserProvidedLabels(userProvidedLabels)

	// add invoker information
	promLabels[MetricLabelInvokerKind] = inv.TypeMeta.Kind
	promLabels[MetricLabelInvokerName] = inv.ObjectMeta.Name
	promLabels[MetricsLabelNamespace] = inv.ObjectMeta.Namespace

	// insert target information as metrics label
	if inv.Driver == api_v1beta1.VolumeSnapshotter {
		promLabels = upsertLabel(promLabels, volumeSnapshotterLabels())
	} else {
		promLabels[MetricsLabelDriver] = string(api_v1beta1.ResticSnapshotter)
		promLabels[MetricsLabelRepository] = inv.Repository
	}

	return promLabels, nil
}

func repoMetricLabels(clientConfig *rest.Config, i invoker.BackupInvoker, userProvidedLabels []string) (prometheus.Labels, error) {
	// add user provided labels
	promLabels := parseUserProvidedLabels(userProvidedLabels)

	// insert repository information as label
	stashClient, err := cs.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	repository, err := stashClient.StashV1alpha1().Repositories(i.ObjectMeta.Namespace).Get(context.TODO(), i.Repository, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	provider, err := repository.Spec.Backend.Provider()
	if err != nil {
		return nil, err
	}
	bucket, err := repository.Spec.Backend.Container()
	if err != nil {
		return nil, err
	}
	prefix, err := repository.Spec.Backend.Prefix()
	if err != nil {
		return nil, err
	}

	promLabels[MetricsLabelName] = repository.Name
	promLabels[MetricsLabelNamespace] = repository.Namespace
	promLabels[MetricsLabelBackend] = provider
	if bucket != "" {
		promLabels[MetricsLabelBucket] = bucket
	}
	if prefix != "" {
		promLabels[MetricsLabelPrefix] = prefix
	}
	return promLabels, nil
}

func upsertLabel(original, new map[string]string) map[string]string {
	labels := make(map[string]string)
	// copy old original labels
	for k, v := range original {
		labels[k] = v
	}
	// insert new labels
	for k, v := range new {
		labels[k] = v
	}
	return labels
}

// targetLabels returns backup/restore target specific labels
func targetLabels(config *rest.Config, target api_v1beta1.TargetRef, namespace string) (map[string]string, error) {

	labels := make(map[string]string)
	switch target.Kind {
	case apis.KindAppBinding:
		appGroup, appKind, err := getAppGroupKind(config, target.Name, namespace)
		// For PerconaXtradDB cluster restore, AppBinding will not exist during restore.
		// In this case, we can not add AppBinding specific labels.
		if err == nil {
			labels[MetricsLabelKind] = appKind
			labels[MetricsLabelAppGroup] = appGroup
		} else if !kerr.IsNotFound(err) {
			return nil, err
		}
	default:
		labels[MetricsLabelKind] = target.Kind
		gv, err := schema.ParseGroupVersion(target.APIVersion)
		if err != nil {
			return nil, err
		}
		labels[MetricsLabelAppGroup] = gv.Group
	}
	labels[MetricsLabelName] = target.Name
	return labels, nil
}

// volumeSnapshotterLabels returns volume snapshot specific labels
func volumeSnapshotterLabels() map[string]string {
	return map[string]string{
		MetricsLabelDriver:   string(api_v1beta1.VolumeSnapshotter),
		MetricsLabelKind:     apis.KindPersistentVolumeClaim,
		MetricsLabelAppGroup: core.GroupName,
	}
}

func getAppGroupKind(clientConfig *rest.Config, name, namespace string) (string, string, error) {
	appClient, err := appcatalog_cs.NewForConfig(clientConfig)
	if err != nil {
		return "", "", err
	}
	appbinding, err := appClient.AppcatalogV1alpha1().AppBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	// if app type is provided then use app group and app resource name.
	// otherwise, default to AppBinding's group,resources name
	targetAppGroup, targetAppResource := appbinding.AppGroupResource()
	if targetAppGroup == "" && targetAppResource == "" {
		targetAppGroup = appbinding.GroupVersionKind().Group
		targetAppResource = appcatalog.ResourceApps
	}
	return targetAppGroup, targetAppResource, nil
}

// parseUserProvidedLabels parses the labels provided by user as an array of key-value pair
// and returns labels in Prometheus labels format
func parseUserProvidedLabels(userLabels []string) prometheus.Labels {
	labels := prometheus.Labels{}
	for _, v := range userLabels {
		parts := strings.Split(v, "=")
		if len(parts) == 2 {
			labels[parts[0]] = parts[1]
		}
	}
	return labels
}
