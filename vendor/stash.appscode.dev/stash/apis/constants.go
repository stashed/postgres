package apis

const (
	Namespace      = "NAMESPACE"
	BackupSession  = "BACKUP_SESSION"
	RestoreSession = "RESTORE_SESSION"

	RepositoryName       = "REPOSITORY_NAME"
	RepositoryProvider   = "REPOSITORY_PROVIDER"
	RepositorySecretName = "REPOSITORY_SECRET_NAME"
	RepositoryBucket     = "REPOSITORY_BUCKET"
	RepositoryPrefix     = "REPOSITORY_PREFIX"
	RepositoryEndpoint   = "REPOSITORY_ENDPOINT"
	RepositoryURL        = "REPOSITORY_URL"

	Hostname = "HOSTNAME"

	TargetName        = "TARGET_NAME"
	TargetAPIVersion  = "TARGET_API_VERSION"
	TargetKind        = "TARGET_KIND"
	TargetResource    = "TARGET_RESOURCE"
	TargetNamespace   = "TARGET_NAMESPACE"
	TargetMountPath   = "TARGET_MOUNT_PATH"
	TargetDirectories = "TARGET_DIRECTORIES"

	RestoreDirectories = "RESTORE_DIRECTORIES"
	RestoreSnapshots   = "RESTORE_SNAPSHOTS"

	RetentionKeepLast    = "RETENTION_KEEP_LAST"
	RetentionKeepHourly  = "RETENTION_KEEP_HOURLY"
	RetentionKeepDaily   = "RETENTION_KEEP_DAILY"
	RetentionKeepWeekly  = "RETENTION_KEEP_WEEKLY"
	RetentionKeepMonthly = "RETENTION_KEEP_MONTHLY"
	RetentionKeepYearly  = "RETENTION_KEEP_YEARLY"
	RetentionKeepTags    = "RETENTION_KEEP_TAGS"
	RetentionPrune       = "RETENTION_PRUNE"
	RetentionDryRun      = "RETENTION_DRY_RUN"

	// default true
	// false when TmpDir.DisableCaching is true in backupConfig/restoreSession
	EnableCache    = "ENABLE_CACHE"
	MaxConnections = "MAX_CONNECTIONS"

	// from runtime settings
	NiceAdjustment  = "NICE_ADJUSTMENT"
	IONiceClass     = "IONICE_CLASS"
	IONiceClassData = "IONICE_CLASS_DATA"

	StatusSubresourceEnabled = "ENABLE_STATUS_SUBRESOURCE"

	PushgatewayURL = "PROMETHEUS_PUSHGATEWAY_URL"
)
