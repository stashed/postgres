package restic

import (
	"sync"
	"time"

	"github.com/appscode/go/types"
	"k8s.io/apimachinery/pkg/util/errors"
	api_v1beta1 "stash.appscode.dev/stash/apis/stash/v1beta1"
)

// RunBackup takes backup, cleanup old snapshots, check repository integrity etc.
// It extract valuable information from respective restic command it runs and return them for further use.
func (w *ResticWrapper) RunBackup(backupOption BackupOptions) (*BackupOutput, error) {
	// Start clock to measure total session duration
	startTime := time.Now()

	// Initialize restic repository if it does not exist
	_, err := w.initRepositoryIfAbsent()
	if err != nil {
		return nil, err
	}

	backupOutput := &BackupOutput{}

	// Run backup
	hostStats, err := w.runBackup(backupOption)
	if err != nil {
		return nil, err
	}
	backupOutput.HostBackupStats = []api_v1beta1.HostBackupStats{hostStats}

	// Check repository integrity
	out, err := w.check()
	if err != nil {
		return nil, err
	}
	// Extract information from output of "check" command
	integrity := extractCheckInfo(out)
	backupOutput.RepositoryStats.Integrity = types.BoolP(integrity)

	// Cleanup old snapshots according to retention policy
	out, err = w.cleanup(backupOption.RetentionPolicy, "")
	if err != nil {
		return nil, err
	}
	// Extract information from output of cleanup command
	kept, removed, err := extractCleanupInfo(out)
	if err != nil {
		return nil, err
	}
	backupOutput.RepositoryStats.SnapshotCount = kept
	backupOutput.RepositoryStats.SnapshotsRemovedOnLastCleanup = removed

	// Read repository statics after cleanup
	out, err = w.stats("")
	if err != nil {
		return nil, err
	}

	// Extract information from output of "stats" command
	repoSize, err := extractStatsInfo(out)
	if err != nil {
		return nil, err
	}
	backupOutput.RepositoryStats.Size = repoSize

	for idx := range backupOutput.HostBackupStats {
		if backupOutput.HostBackupStats[idx].Hostname == backupOption.Host {
			backupOutput.HostBackupStats[idx].Duration = time.Since(startTime).String()
			backupOutput.HostBackupStats[idx].Phase = api_v1beta1.HostBackupSucceeded
		}
	}

	return backupOutput, nil
}

// RunParallelBackup runs multiple backup in parallel.
// Host must be different for each backup.
func (w *ResticWrapper) RunParallelBackup(backupOptions []BackupOptions, maxConcurrency int) (*BackupOutput, error) {

	// Initialize restic repository if it does not exist
	_, err := w.initRepositoryIfAbsent()
	if err != nil {
		return nil, err
	}

	// WaitGroup to wait until all go routine finish
	wg := sync.WaitGroup{}
	// concurrencyLimiter channel is used to limit maximum number simultaneous go routine
	concurrencyLimiter := make(chan bool, maxConcurrency)
	defer close(concurrencyLimiter)

	var (
		backupErrs []error
		mu         sync.Mutex // use lock to avoid racing condition
	)

	backupOutput := &BackupOutput{}

	for i := range backupOptions {
		// try to send message in concurrencyLimiter channel.
		// if maximum allowed concurrent backup is already running, program control will stuck here.
		concurrencyLimiter <- true

		// starting new go routine. add it to WaitGroup
		wg.Add(1)

		go func(opt BackupOptions, startTime time.Time) {

			// when this go routine completes it task, release a slot from the concurrencyLimiter channel
			// so that another go routine can start. Also, tell the WaitGroup that it is done with its task.
			defer func() {
				<-concurrencyLimiter
				wg.Done()
			}()

			// sh field in ResticWrapper is a pointer. we must not use same w in multiple go routine.
			// otherwise they might enter in a racing condition.
			nw := w.Copy()

			hostStats, err := nw.runBackup(opt)
			if err != nil {
				// acquire lock to make sure no other go routine is writing to backupErr
				mu.Lock()
				backupErrs = append(backupErrs, err)
				mu.Unlock()
				return
			}
			hostStats.Duration = time.Since(startTime).String()
			hostStats.Phase = api_v1beta1.HostBackupSucceeded

			// add hostStats to backupOutput. use lock to avoid racing condition.
			mu.Lock()
			backupOutput.upsertHostBackupStats(hostStats)
			mu.Unlock()
		}(backupOptions[i], time.Now())
	}

	// wait for all the go routines to complete
	wg.Wait()

	if backupErrs != nil {
		return nil, errors.NewAggregate(backupErrs)
	}

	// Check repository integrity
	out, err := w.check()
	if err != nil {
		return nil, err
	}
	// Extract information from output of "check" command
	integrity := extractCheckInfo(out)
	backupOutput.RepositoryStats.Integrity = types.BoolP(integrity)

	// Cleanup old snapshots according to retention policy
	backupOutput.RepositoryStats.SnapshotCount = 0
	backupOutput.RepositoryStats.SnapshotsRemovedOnLastCleanup = 0
	for _, opt := range backupOptions {
		out, err = w.cleanup(opt.RetentionPolicy, opt.Host)
		if err != nil {
			return nil, err
		}
		// Extract information from output of cleanup command
		kept, removed, err := extractCleanupInfo(out)
		if err != nil {
			return nil, err
		}
		backupOutput.RepositoryStats.SnapshotCount += kept
		backupOutput.RepositoryStats.SnapshotsRemovedOnLastCleanup += removed
	}

	// Read repository statics after cleanup
	out, err = w.stats("")
	if err != nil {
		return nil, err
	}

	// Extract information from output of "stats" command
	repoSize, err := extractStatsInfo(out)
	if err != nil {
		return nil, err
	}
	backupOutput.RepositoryStats.Size = repoSize

	return backupOutput, nil
}

func (w *ResticWrapper) runBackup(backupOption BackupOptions) (api_v1beta1.HostBackupStats, error) {
	hostStats := api_v1beta1.HostBackupStats{
		Hostname: backupOption.Host,
	}

	//fmt.Println("shell: ",w)
	// Backup from stdin
	if backupOption.StdinPipeCommand.Name != "" {
		out, err := w.backupFromStdin(backupOption)
		if err != nil {
			return hostStats, err
		}
		// Extract information from the output of backup command
		snapStats, err := extractBackupInfo(out, backupOption.StdinFileName, backupOption.Host)
		if err != nil {
			return hostStats, err
		}
		hostStats.Snapshots = []api_v1beta1.SnapshotStats{snapStats}
		return hostStats, nil
	}

	// Backup all target paths
	for _, path := range backupOption.BackupPaths {
		out, err := w.backup(path, backupOption.Host, nil)
		if err != nil {
			return hostStats, err
		}
		// Extract information from the output of backup command
		stats, err := extractBackupInfo(out, path, backupOption.Host)
		if err != nil {
			return hostStats, err
		}
		hostStats = upsertSnapshotStats(hostStats, stats)
	}

	return hostStats, nil
}

func upsertSnapshotStats(hostStats api_v1beta1.HostBackupStats, snapStats api_v1beta1.SnapshotStats) api_v1beta1.HostBackupStats {
	for i, s := range hostStats.Snapshots {
		// if there is already an entry for this snapshot, then update it
		if s.Name == snapStats.Name {
			hostStats.Snapshots[i] = snapStats
			return hostStats
		}
	}
	// no entry for this snapshot. add a new entry
	hostStats.Snapshots = append(hostStats.Snapshots, snapStats)
	return hostStats
}

func (backupOutput *BackupOutput) upsertHostBackupStats(hostStats api_v1beta1.HostBackupStats) {

	// check if a entry already exist for this host in backupOutput. If exist then update it.
	for i, v := range backupOutput.HostBackupStats {
		if v.Hostname == hostStats.Hostname {
			backupOutput.HostBackupStats[i] = hostStats
			return
		}
	}

	// no entry for this host. add a new entry
	backupOutput.HostBackupStats = append(backupOutput.HostBackupStats, hostStats)
	return
}
