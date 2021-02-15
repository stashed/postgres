---
title: PostgreSQL | Stash
description: Stash auto-backup for PostgreSQL database
menu:
  docs_{{ .version }}:
    identifier: standalone-postgres-{{ .subproject_version }}-auto-backup
    name: Auto-Backup
    parent: stash-postgres-guides-{{ .subproject_version }}
    weight: 20
product_name: stash
menu_name: docs_{{ .version }}
section_menu_id: stash-addons
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [Stash Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Backup Elasticsearch using Stash Auto-Backup

Stash can be configured to automatically backup any Elasticsearch database in your cluster. Stash enables cluster administrators to deploy backup blueprints ahead of time so that the database owners can easily backup their database with just a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for Elasticsearch databases in your cluster and backup them with few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.
- Install Stash in your cluster following the steps [here](/docs/setup/README.md).
- Install [KubeDB](https://kubedb.com) in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/). This step is optional. You can deploy your database using any method you want.
- If you are not familiar with how Stash backup and restore Elasticsearch databases, please check the following guide [here](/docs/addons/elasticsearch/overview.md).
- If you are not familiar with how auto-backup works in Stash, please check the following guide [here](/docs/guides/latest/auto-backup/overview.md).
- If you are not familiar with the available auto-backup options for databases in Stash, please check the following guide [here](/docs/guides/latest/auto-backup/database.md).

You should be familiar with the following `Stash` concepts:

- [BackupBlueprint](/docs/concepts/crds/backupblueprint.md)
- [BackupConfiguration](/docs/concepts/crds/backupconfiguration.md)
- [BackupSession](/docs/concepts/crds/backupsession.md)
- [Repository](/docs/concepts/crds/repository.md)
- [Function](/docs/concepts/crds/function.md)
- [Task](/docs/concepts/crds/task.md)

In this tutorial, we are going to show backup of three different Elasticsearch databases on three different namespaces named `demo`, `demo-2`, and `demo-3`. Create the namespaces as below if you haven't done it already.

```bash
❯ kubectl create ns demo
namespace/demo created

❯ kubectl create ns demo-2
namespace/demo-2 created

❯ kubectl create ns demo-3
namespace/demo-3 created
```

Make sure you have installed the Elasticsearch addon for Stash. If you haven't installed it already, please install the addon following the steps [here](/docs/addons/elasticsearch/setup/install.md).

```bash
❯ kubectl get tasks.stash.appscode.com | grep elasticsearch
elasticsearch-backup-5.6.4-v6    4d4h
elasticsearch-backup-6.2.4-v6    4d4h
elasticsearch-backup-6.3.0-v6    4d4h
elasticsearch-backup-6.4.0-v6    4d4h
elasticsearch-backup-6.5.3-v6    4d4h
elasticsearch-backup-6.8.0-v6    4d4h
elasticsearch-backup-7.2.0-v6    4d4h
elasticsearch-backup-7.3.2-v6    4d4h
elasticsearch-restore-5.6.4-v6   4d4h
elasticsearch-restore-6.2.4-v6   4d4h
elasticsearch-restore-6.3.0-v6   4d4h
elasticsearch-restore-6.4.0-v6   4d4h
elasticsearch-restore-6.5.3-v6   4d4h
elasticsearch-restore-6.8.0-v6   4d4h
elasticsearch-restore-7.2.0-v6   4d4h
elasticsearch-restore-7.3.2-v6   4d4h
```

## Prepare Backup Blueprint

To backup an Elasticsearch database using Stash, you have to create a `Secret` containing the backend credentials, a `Repository` containing the backend information, and a `BackupConfiguration` containing the schedule and target information. A `BackupBlueprint` allows you to specify a template for the `Repository` and the `BackupConfiguration`.

The `BackupBlueprint` is a non-namespaced CRD. So, once you have created a `BackupBlueprint`, you can use it to backup any Elasticsearch database of any namespace just by creating the storage `Secret` in that namespace and adding few annotations to your Elasticsearch CRO. Then, Stash will automatically create a `Repository` and a `BackupConfiguration` according to the template to backup the database.

Below is the `BackupBlueprint` object that we are going to use in this tutorial,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupBlueprint
metadata:
  name: elasticsearch-backup-template
spec:
  # ============== Blueprint for Repository ==========================
  backend:
    gcs:
      bucket: stash-testing
      prefix: stash-backup/${TARGET_NAMESPACE}/${TARGET_APP_RESOURCE}/${TARGET_NAME}
    storageSecretName: gcs-secret
  # ============== Blueprint for BackupConfiguration =================
  task:
    name: elasticsearch-backup-{{< param "info.subproject_version" >}}
  schedule: "*/5 * * * *"
  interimVolumeTemplate:
    metadata:
      name: ${TARGET_APP_RESOURCE}-${TARGET_NAME} # To ensure that the PVC names are unique for different database
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
```

Here, we are using a GCS bucket as our backend. We are providing `gcs-secret` at the `storageSecretName` field. Hence, we have to create a secret named `gcs-secret` with the access credentials of our bucket in every namespace where we want to enable backup through this blueprint.

Notice the `prefix` field of `backend` section. We have used some variables in form of `${VARIABLE_NAME}`. Stash will automatically resolve those variables from the database information to make the backend prefix unique for each database instance.

We have also used some variables in `name` field of the `interimVolumeTemplate` section. This is to ensure that the generated PVC name becomes unique for the database instances.

Let's create the `BackupBlueprint` we have shown above,

```bash
❯ kubectl apply -f https://github.com/stashed/elasticsearch/raw/{{< param "info.subproject_version" >}}/docs/auto-backup/examples/backupblueprint.yaml
backupblueprint.stash.appscode.com/elasticsearch-backup-template created
```

Now, we are ready to backup our Elasticsearch databases using few annotations. You can check available auto-backup annotations for a databases from [here](/docs/guides/latest/auto-backup/database/#available-auto-backup-annotations-for-database).

## Auto-backup with default configurations

In this section, we are going to backup an Elasticsearch database of `demo` namespace. We are going to use the default configurations specified in the `BackupBlueprint`.

### Create Storage Secret

At first, let's create the `gcs-secret` in `demo` namespace with the access credentials to our GCS bucket.

```bash
❯ echo -n 'changeit' > RESTIC_PASSWORD
❯ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
❯ cat downloaded-sa-json.key > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
❯ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

### Create Database

Now, we are going to create an Elasticsearch CRO in `demo` namespace. Below is the YAML of the Elasticsearch object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-demo
  namespace: demo
  annotations:
    stash.appscode.com/backup-blueprint: elasticsearch-backup-template
spec:
  version: 7.9.1-xpack
  replicas: 1
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  terminationPolicy: WipeOut
```

Notice the `annotations` section. We are pointing to the `BackupBlueprint` that we have created earlier though `stash.appscode.com/backup-blueprint` annotation. Stash will watch this annotation and create a `Repository` and a `BackupConfiguration` according to the `BackupBlueprint`.

Let's create the above Elasticsearch CRO,

```bash
❯ kubectl apply -f https://github.com/stashed/elasticsearch/raw/{{< param "info.subproject_version" >}}/docs/auto-backup/examples/es-demo.yaml
elasticsearch.kubedb.com/sample-elasticsearch created
```

### Verify Auto-backup configured

In this section, we are going to verify whether Stash has created the respective `Repository` and `BackupConfiguration` for our Elasticsearch database we have just deployed or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our Elasticsearch or not.

```bash
❯ kubectl get repository -n demo
NAME          INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-es-demo                                                                5s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo app-es-demo -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
...
  name: app-es-demo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: stash-backup/demo/elasticsearch/es-demo
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this database.

#### Verify BackupConfiguration

Now, let's verify whether Stash has created a `BackupConfiguration` for our Elasticsearch or not.

```bash
❯ kubectl get backupconfiguration -n demo
NAME          TASK                            SCHEDULE      PAUSED   AGE
app-es-demo   elasticsearch-backup-{{< param "info.subproject_version" >}}   */5 * * * *            12s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo app-es-demo -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: app-es-demo
  namespace: demo
  ...
spec:
  driver: Restic
  interimVolumeTemplate:
    metadata:
      name: elasticsearch-es-demo
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
    status: {}
  repository:
    name: app-es-demo
  retentionPolicy:
    keepLast: 5
    name: keep-last-5
    prune: true
  runtimeSettings: {}
  schedule: '*/5 * * * *'
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: es-demo
  task:
    name: elasticsearch-backup-{{< param "info.subproject_version" >}}
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2021-02-12T11:46:53Z"
    message: Repository demo/app-es-demo exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2021-02-12T11:46:53Z"
    message: Backend Secret demo/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2021-02-12T11:46:53Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/es-demo found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2021-02-12T11:46:53Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `interimVolumeTemplate` section. The variables of `name` field have been substituted by the equivalent information from the database.

Also, notice the `target` section. Stash has automatically added the Elasticsearch as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo -w
NAME                     INVOKER-TYPE          INVOKER-NAME   PHASE       AGE
app-es-demo-1613130605   BackupConfiguration   app-es-demo                0s
app-es-demo-1613130605   BackupConfiguration   app-es-demo    Running     10s
app-es-demo-1613130605   BackupConfiguration   app-es-demo    Succeeded   46s
```

Once the backup has been completed successfully, you should see the backed up data has been stored in the bucket at the directory pointed by the `prefix` field of the `Repository`.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/addons/elasticsearch/guides/{{< param "info.subproject_version" >}}/auto-backup/images/es-demo.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with a custom schedule

In this section, we are going to backup an Elasticsearch database of `demo-2` namespace. This time, we are going to overwrite the default schedule used in the `BackupBlueprint`.

### Create Storage Secret

At first, let's create the `gcs-secret` in `demo-2` namespace with the access credentials to our GCS bucket.

```bash
❯ kubectl create secret generic -n demo-2 gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

### Create Database

Now, we are going to create an Elasticsearch CRO in `demo-2` namespace. Below is the YAML of the Elasticsearch object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-demo-2
  namespace: demo-2
  annotations:
    stash.appscode.com/backup-blueprint: elasticsearch-backup-template
    stash.appscode.com/schedule: "*/3 * * * *"
spec:
  version: 7.9.1-xpack
  replicas: 1
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  terminationPolicy: WipeOut
```

Notice the `annotations` section. This time, we have passed a schedule via `stash.appscode.com/schedule` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above Elasticsearch CRO,

```bash
❯ kubectl apply -f https://github.com/stashed/elasticsearch/raw/{{< param "info.subproject_version" >}}/docs/auto-backup/examples/es-demo-2.yaml
elasticsearch.kubedb.com/es-demo-2 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup has been configured properly or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our Elasticsearch or not.

```bash
❯ kubectl get repository -n demo-2
NAME            INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-es-demo-2                                                                25s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-2 app-es-demo-2 -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: app-es-demo-2
  namespace: demo-2
  ...
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: stash-backup/demo-2/elasticsearch/es-demo-2
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

Now, let's verify whether Stash has created a `BackupConfiguration` for our Elasticsearch or not.

```bash
❯ kubectl get backupconfiguration -n demo-2
NAME            TASK                            SCHEDULE      PAUSED   AGE
app-es-demo-2   elasticsearch-backup-{{< param "info.subproject_version" >}}   */3 * * * *            77s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-2 app-es-demo-2 -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: app-es-demo-2
  namespace: demo-2
  ...
spec:
  driver: Restic
  interimVolumeTemplate:
    metadata:
      name: elasticsearch-es-demo-2
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
    status: {}
  repository:
    name: app-es-demo-2
  retentionPolicy:
    keepLast: 5
    name: keep-last-5
    prune: true
  runtimeSettings: {}
  schedule: '*/3 * * * *'
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: es-demo-2
  task:
    name: elasticsearch-backup-{{< param "info.subproject_version" >}}
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2021-02-12T12:24:07Z"
    message: Repository demo-2/app-es-demo-2 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2021-02-12T12:24:07Z"
    message: Backend Secret demo-2/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2021-02-12T12:24:07Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/es-demo-2 found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2021-02-12T12:24:07Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `schedule` section. This time the `BackupConfiguration` has been created with the schedule we have provided via annotation.

Also, notice the `target` section. Stash has automatically added the new Elasticsearch as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-2 -w
NAME                       INVOKER-TYPE          INVOKER-NAME    PHASE       AGE
app-es-demo-2-1613132831   BackupConfiguration   app-es-demo-2               0s
app-es-demo-2-1613132831   BackupConfiguration   app-es-demo-2   Running     17s
app-es-demo-2-1613132831   BackupConfiguration   app-es-demo-2   Succeeded   41s
```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/addons/elasticsearch/guides/{{< param "info.subproject_version" >}}/auto-backup/images/es-demo-2.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with custom parameters

In this section, we are going to backup an Elasticsearch database of `demo-3` namespace. This time, we are going to pass some parameters for the Task through the annotations.

### Create Storage Secret

At first, let's create the `gcs-secret` in `demo-3` namespace with the access credentials to our GCS bucket.

```bash
❯ kubectl create secret generic -n demo-3 gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

### Create Database

Now, we are going to create an Elasticsearch CRO in `demo-3` namespace. Below is the YAML of the Elasticsearch object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-demo-3
  namespace: demo-3
  annotations:
    stash.appscode.com/backup-blueprint: elasticsearch-backup-template
    params.stash.appscode.com/args: --ignoreType=settings,template
spec:
  version: 7.9.1-xpack
  replicas: 1
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  terminationPolicy: WipeOut
```

Notice the `annotations` section. This time, we have passed an argument via `params.stash.appscode.com/args` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above Elasticsearch CRO,

```bash
❯ kubectl apply -f https://github.com/stashed/elasticsearch/raw/{{< param "info.subproject_version" >}}/docs/auto-backup/examples/es-demo-3.yaml
elasticsearch.kubedb.com/es-demo-3 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup resources has been created or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our Elasticsearch or not.

```bash
❯ kubectl get repository -n demo-3
NAME            INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-es-demo-3                                                                23s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-3 app-es-demo-3 -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: app-es-demo-3
  namespace: demo-3
  ...
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: stash-backup/demo-3/elasticsearch/es-demo-3
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

Now, let's verify whether Stash has created a `BackupConfiguration` for our Elasticsearch or not.

```bash
❯ kubectl get backupconfiguration -n demo-3
NAME            TASK                            SCHEDULE      PAUSED   AGE
app-es-demo-3   elasticsearch-backup-{{< param "info.subproject_version" >}}   */5 * * * *            84s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-3 app-es-demo-3 -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: app-es-demo-3
  namespace: demo-3
  ...
spec:
  driver: Restic
  interimVolumeTemplate:
    metadata:
      name: elasticsearch-es-demo-3
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
    status: {}
  repository:
    name: app-es-demo-3
  retentionPolicy:
    keepLast: 5
    name: keep-last-5
    prune: true
  runtimeSettings: {}
  schedule: '*/5 * * * *'
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: es-demo-3
  task:
    name: elasticsearch-backup-{{< param "info.subproject_version" >}}
    params:
    - name: args
      value: --ignoreType=settings,template
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2021-02-12T12:39:14Z"
    message: Repository demo-3/app-es-demo-3 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2021-02-12T12:39:14Z"
    message: Backend Secret demo-3/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2021-02-12T12:39:14Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/es-demo-3 found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2021-02-12T12:39:14Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `task` section. The `args` parameter that we had passed via annotations has been added to the `params` section.

Also, notice the `target` section. Stash has automatically added the new Elasticsearch as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-3 -w
NAME                       INVOKER-TYPE          INVOKER-NAME    PHASE       AGE
app-es-demo-3-1613133604   BackupConfiguration   app-es-demo-3               0s
app-es-demo-3-1613133604   BackupConfiguration   app-es-demo-3   Running     5s
app-es-demo-3-1613133604   BackupConfiguration   app-es-demo-3   Succeeded   48s
```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/addons/elasticsearch/guides/{{< param "info.subproject_version" >}}/auto-backup/images/es-demo-3.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
❯ kubectl delete -f https://github.com/stashed/elasticsearch/raw/{{< param "info.subproject_version" >}}/docs/auto-backup/examples/
backupblueprint.stash.appscode.com "elasticsearch-backup-template" deleted
elasticsearch.kubedb.com "es-demo-2" deleted
elasticsearch.kubedb.com "es-demo-3" deleted
elasticsearch.kubedb.com "es-demo" deleted

❯ kubectl delete repository -n demo --all
repository.stash.appscode.com "app-es-demo" deleted
❯ kubectl delete repository -n demo-2 --all
repository.stash.appscode.com "app-es-demo-2" deleted
❯ kubectl delete repository -n demo-3 --all
repository.stash.appscode.com "app-es-demo-3" deleted
```

<!-- # Auto Backup for Database

This tutorial will show you how to configure automatic backup for PostgreSQL database using Stash. Here, we are going to backup two different PostgreSQL databases of two different version using a common blueprint.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install Stash in your cluster following the steps [here](/docs/setup/README.md).
- Install PostgreSQL addon for Stash following the steps [here](/docs/addons/postgres/setup/install.md).
- Install [KubeDB](https://kubedb.com) in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/). This step is optional. You can deploy your database using any method you want. We are using KubeDB because KubeDB simplifies many of the difficult or tedious management tasks of running a production grade databases on private and public clouds.
- If you are not familiar with how Stash backup and restore PostgreSQL databases, please check the following guide [here](/docs/addons/postgres/overview.md).

You should be familiar with the following `Stash` concepts:

- [BackupBlueprint](/docs/concepts/crds/backupblueprint.md)
- [BackupConfiguration](/docs/concepts/crds/backupconfiguration.md)
- [BackupSession](/docs/concepts/crds/backupsession.md)
- [Repository](/docs/concepts/crds/repository.md)
- [Function](/docs/concepts/crds/function.md)
- [Task](/docs/concepts/crds/task.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Backup Blueprint

We are going to use [GCS Backend](/docs/guides/latest/backends/gcs.md) to store the backed up data. You can use any supported backend you prefer. You just have to configure Storage Secret and `spec.backend` section of `BackupBlueprint` to match with your backend. To learn which backends are supported by Stash and how to configure them, please visit [here](/docs/guides/latest/backends/overview.md).

> For GCS backend, if the bucket does not exist, Stash needs `Storage Object Admin` role permissions to create the bucket. For more details, please check the following [guide](/docs/guides/latest/backends/gcs.md).

**Create Storage Secret:**

At first, let's create a Storage Secret for the GCS backend,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ mv downloaded-sa-json.key GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

**Create BackupBlueprint:**

Now, we have to create a `BackupBlueprint` crd with a blueprint for `Repository` and `BackupConfiguration` object.

Below is the YAML of the `BackupBlueprint` object that we are going to create,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupBlueprint
metadata:
  name: postgres-backup-blueprint
spec:
  # ============== Blueprint for Repository ==========================
  backend:
    gcs:
      bucket: appscode-qa
      prefix: stash-backup/${TARGET_NAMESPACE}/${TARGET_APP_RESOURCE}/${TARGET_NAME}
    storageSecretName: gcs-secret
  # ============== Blueprint for BackupConfiguration =================
  task:
    name: postgres-backup-${TARGET_APP_VERSION}
  schedule: "*/5 * * * *"
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
```

Here,

- `spec.task.name` specifies the `Task` crd name that will be used to backup the targeted database. We have used a variable `${TARGET_APP_VERSION}` as task name suffix. This variable will be substituted by the respective database version. This allows to backup multiple database versions with a common blueprint.

Note that we have used some variables (format: `${<variable name>}`) in `spec.backend.gcs.prefix` field. Stash will substitute these variables with values from the respective target. To learn which variables you can use in the `prefix` field, please visit [here](/docs/concepts/crds/backupblueprint.md#repository-blueprint).

Let's create the `BackupBlueprint` that we have shown above,

```bash
$ kubectl apply -f https://github.com/stashed/postgres/raw/{{< param "info.subproject_version" >}}/docs/auto-backup/examples/backupblueprint.yaml
backupblueprint.stash.appscode.com/postgres-backup-blueprint created
```

Now, automatic backup is configured for PostgreSQL database. We just have to add a annotation to the `AppBinding` of the targeted database. In order to know the available annotations for database auto-backup, please visit [here](docs/guides/latest/auto-backup/database/#available-auto-backup-annotations-for-database).

> Note: `BackupBlueprint` is a non-namespaced crd. So, you can use a `BackupBlueprint` to backup targets in multiple namespaces. However, Storage Secret is a namespaced object. So, you have to manually create the secret in each namespace where you have a target for backup. Please give us your feedback on how to improve the ux of this aspect of Stash on [GitHub](https://github.com/stashed/stash/issues/842).

## Prepare Databases

Now, we are going to deploy two sample PostgreSQL databases of two different versions using KubeDB. We are going to backup these two databases using auto-backup.

**Deploy First PostgreSQL Sample:**

Below is the YAML of the first `Postgres` crd,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: sample-postgres-1
  namespace: demo
spec:
  version: "11.2"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: Delete
```

Let's create the `Postgres` we have shown above,

```bash
$ kubectl apply -f https://github.com/stashed/docs/raw/{{< param "info.version" >}}/docs/examples/guides/latest/auto-backup/database/sample-postgres-1.yaml
postgres.kubedb.com/sample-postgres-1 created
```

KubeDB will deploy a PostgreSQL database according to the above specification and it will create the necessary secrets and services to access the database. It will also create an `AppBinding` crd that holds the necessary information to connect with the database.

Verify that an `AppBinding` has been created for this PostgreSQL sample,

```bash
$ kubectl get appbinding -n demo
NAME                AGE
sample-postgres-1   47s
```

If you view the YAML of this `AppBinding`, you will see it holds service and secret information. Stash uses this information to connect with the database.

```bash
$ kubectl get appbinding -n demo sample-postgres-1 -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: sample-postgres-1
  namespace: demo
  ...
spec:
  clientConfig:
    service:
      name: sample-postgres-1
      path: /
      port: 5432
      query: sslmode=disable
      scheme: postgresql
  secret:
    name: sample-postgres-1-auth
  secretTransforms:
  - renameKey:
      from: POSTGRES_USER
      to: username
  - renameKey:
      from: POSTGRES_PASSWORD
      to: password
  type: kubedb.com/postgres
  version: "11.2"
```

> If you have deployed the database without KubeDB, you have to create the AppBinding crd manually.

**Deploy Second PostgreSQL Sample:**

Below is the YAML of second `Postgres` object,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: sample-postgres-2
  namespace: demo
spec:
  version: "10.6-v2"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: Delete
```

Let's create the `Postgres` we have shown above,

```bash
$ kubectl apply -f https://github.com/stashed/docs/raw/{{< param "info.version" >}}/docs/examples/guides/latest/auto-backup/database/sample-postgres-2.yaml
postgres.kubedb.com/sample-postgres-2 created
```

Verify that an `AppBinding` has been created for this PostgreSQL database,

```bash
$ kubectl get appbinding -n demo
NAME                AGE
sample-postgres-1   2m49s
sample-postgres-2   10s
```

Here, we can see `AppBinding` `sample-postgres-2` has been created for our second PostgreSQL sample.

## Backup

Now, we are going to add auto-backup specific annotation to the `AppBinding` of our desired database. Stash watches for `AppBinding` crd. Once it finds an `AppBinding` with auto-backup annotation, it will create a `Repository` and a `BackupConfiguration` crd according to respective `BackupBlueprint`. Then, rest of the backup process will proceed as normal database backup as described [here](/docs/guides/latest/addons/overview.md).

### Backup First PostgreSQL Sample

Let's backup our first PostgreSQL sample using auto-backup.

**Add Annotations:**

At first, add the auto-backup specific annotation to the AppBinding `sample-postgres-1`,

```bash
$ kubectl annotate appbinding sample-postgres-1 -n demo --overwrite \
  stash.appscode.com/backup-blueprint=postgres-backup-blueprint
```

Verify that the annotation has been added successfully,

```bash
$ kubectl get appbinding -n demo sample-postgres-1 -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    stash.appscode.com/backup-blueprint: postgres-backup-blueprint
  name: sample-postgres-1
  namespace: demo
  ...
spec:
  clientConfig:
    service:
      name: sample-postgres-1
      path: /
      port: 5432
      query: sslmode=disable
      scheme: postgresql
  secret:
    name: sample-postgres-1-auth
  secretTransforms:
  - renameKey:
      from: POSTGRES_USER
      to: username
  - renameKey:
      from: POSTGRES_PASSWORD
      to: password
  type: kubedb.com/postgres
  version: "11.2"
```

Now, Stash will create a `Repository` crd and a `BackupConfiguration` crd according to the blueprint.

**Verify Repository:**

Verify that the `Repository` has been created successfully by the following command,

```bash
$ kubectl get repository -n demo
NAME                         INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
postgres-sample-postgres-1                                                                2m23s
```

If we view the YAML of this `Repository`, we are going to see that the variables `${TARGET_NAMESPACE}`, `${TARGET_APP_RESOURCE}` and `${TARGET_NAME}` has been replaced by `demo`, `postgres` and `sample-postgres-1` respectively.

```bash
$ kubectl get repository -n demo postgres-sample-postgres-1 -o yaml
```

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: Repository
metadata:
  creationTimestamp: "2019-08-01T13:54:48Z"
  finalizers:
  - stash
  generation: 1
  name: postgres-sample-postgres-1
  namespace: demo
  resourceVersion: "50171"
  selfLink: /apis/stash.appscode.com/v1beta1/namespaces/demo/repositories/postgres-sample-postgres-1
  uid: ed49dde4-b463-11e9-a6a0-080027aded7e
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: stash-backup/demo/postgres/sample-postgres-1
    storageSecretName: gcs-secret
```

**Verify BackupConfiguration:**

Verify that the `BackupConfiguration` crd has been created by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                         TASK                   SCHEDULE      PAUSED   AGE
postgres-sample-postgres-1   postgres-backup-11.2   */5 * * * *            3m39s
```

Notice the `TASK` field. It denoting that this backup will be performed using `postgres-backup-11.2` task. We had specified `postgres-backup-${TARGET_APP_VERSION}` as task name in the `BackupBlueprint`. Here, the variable `${TARGET_APP_VERSION}` has been substituted by the database version.

Let's check the YAML of this `BackupConfiguration`,

```bash
$ kubectl get backupconfiguration -n demo postgres-sample-postgres-1 -o yaml
```

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2019-08-01T13:54:48Z"
  finalizers:
  - stash.appscode.com
  generation: 1
  name: postgres-sample-postgres-1
  namespace: demo
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: false
    kind: AppBinding
    name: sample-postgres-1
    uid: a799156e-b463-11e9-a6a0-080027aded7e
  resourceVersion: "50170"
  selfLink: /apis/stash.appscode.com/v1beta1/namespaces/demo/backupconfigurations/postgres-sample-postgres-1
  uid: ed4bd257-b463-11e9-a6a0-080027aded7e
spec:
  repository:
    name: postgres-sample-postgres-1
  retentionPolicy:
    keepLast: 5
    name: keep-last-5
    prune: true
  runtimeSettings: {}
  schedule: '*/5 * * * *'
  target:
    ref:
      apiVersion: v1
      kind: AppBinding
      name: sample-postgres-1
  task:
    name: postgres-backup-11.2
  tempDir: {}
```

Notice that the `spec.target.ref` is pointing to the AppBinding `sample-postgres-1` that we have just annotated with auto-backup annotation.

**Wait for BackupSession:**

Now, wait for the next backup schedule. Run the following command to watch `BackupSession` crd:

```bash
$ watch -n 1 kubectl get backupsession -n demo -l=stash.appscode.com/backup-configuration=postgres-sample-postgres-1

Every 1.0s: kubectl get backupsession -n demo -l=stash.appscode.com/backup-configuration=postgres-sample-postgres-1  workstation: Thu Aug  1 20:35:43 2019

NAME                                    INVOKER-TYPE          INVOKER-NAME                 PHASE       AGE
postgres-sample-postgres-1-1564670101   BackupConfiguration   postgres-sample-postgres-1   Succeeded   42s
```

> Note: Backup CronJob creates `BackupSession` crd with the following label `stash.appscode.com/backup-configuration=<BackupConfiguration crd name>`. We can use this label to watch only the `BackupSession` of our desired `BackupConfiguration`.

**Verify Backup:**

When backup session is completed, Stash will update the respective `Repository` to reflect the latest state of backed up data.

Run the following command to check if a snapshot has been sent to the backend,

```bash
$ kubectl get repository -n demo postgres-sample-postgres-1
NAME                         INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
postgres-sample-postgres-1   true        1.324 KiB   1                73s                      6m7s
```

If we navigate to `stash-backup/demo/postgres/sample-postgres-1` directory of our GCS bucket, we are going to see that the snapshot has been stored there.

<figure align="center">
  <img alt="Backup data of PostgreSQL 'sample-postgres-1' in GCS backend" src="/docs/images/guides/latest/auto-backup/database/sample_postgres_1.png">
  <figcaption align="center">Fig: Backup data of PostgreSQL "sample-postgres-1" in GCS backend</figcaption>
</figure>

### Backup Second Sample PostgreSQL

Now, lets backup our second PostgreSQL sample using the same `BackupBlueprint` we have used to backup the first PostgreSQL sample.

**Add Annotations:**

Add the auto backup specific annotation to AppBinding `sample-postgres-2`,

```bash
$ kubectl annotate appbinding sample-postgres-2 -n demo --overwrite \
  stash.appscode.com/backup-blueprint=postgres-backup-blueprint
```

**Verify Repository:**

Verify that the `Repository` has been created successfully by the following command,

```bash
$ kubectl get repository -n demo
NAME                         INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
postgres-sample-postgres-1   true        1.324 KiB   1                2m3s                     6m57s
postgres-sample-postgres-2                                                                     15s
```

Here, Repository `postgres-sample-postgres-2` has been created for the second PostgreSQL sample.

If we view the YAML of this `Repository`, we are going to see that the variables `${TARGET_NAMESPACE}`, `${TARGET_APP_RESOURCE}` and `${TARGET_NAME}` has been replaced by `demo`, `postgres` and `sample-postgres-2` respectively.

```bash
$ kubectl get repository -n demo postgres-sample-postgres-2 -o yaml
```

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: Repository
metadata:
  creationTimestamp: "2019-08-01T14:37:22Z"
  finalizers:
  - stash
  generation: 1
  name: postgres-sample-postgres-2
  namespace: demo
  resourceVersion: "56103"
  selfLink: /apis/stash.appscode.com/v1beta1/namespaces/demo/repositories/postgres-sample-postgres-2
  uid: df58523c-b469-11e9-a6a0-080027aded7e
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: stash-backup/demo/postgres/sample-postgres-2
    storageSecretName: gcs-secret
```

**Verify BackupConfiguration:**

Verify that the `BackupConfiguration` crd has been created by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                         TASK                   SCHEDULE      PAUSED   AGE
postgres-sample-postgres-1   postgres-backup-11.2   */5 * * * *            7m52s
postgres-sample-postgres-2   postgres-backup-10.6   */5 * * * *            70s
```

Again, notice the `TASK` field. This time, `${TARGET_APP_VERSION}` has been replaced with `10.6` which is the database version of our second sample.

**Wait for BackupSession:**

Now, wait for the next backup schedule. Run the following command to watch `BackupSession` crd:

```bash
$ watch -n 1 kubectl get backupsession -n demo -l=stash.appscode.com/backup-configuration=postgres-sample-postgres-2
Every 1.0s: kubectl get backupsession -n demo -l=stash.appscode.com/backup-configuration=postgres-sample-postgres-2  workstation: Thu Aug  1 20:55:40 2019

NAME                                    INVOKER-TYPE          INVOKER-NAME                 PHASE       AGE
postgres-sample-postgres-2-1564671303   BackupConfiguration   postgres-sample-postgres-2   Succeeded   37s
```

**Verify Backup:**

Run the following command to check if a snapshot has been sent to the backend,

```bash
$ kubectl get repository -n demo postgres-sample-postgres-2
NAME                         INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
postgres-sample-postgres-2   true        1.324 KiB   1                52s                      19m
```

If we navigate to `stash-backup/demo/postgres/sample-postgres-2` directory of our GCS bucket, we are going to see that the snapshot has been stored there.

<figure align="center">
  <img alt="Backup data of PostgreSQL 'sample-postgres-2' in GCS backend" src="/docs/images/guides/latest/auto-backup/database/sample_postgres_2.png">
  <figcaption align="center">Fig: Backup data of PostgreSQL "sample-postgres-2" in GCS backend</figcaption>
</figure>

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pg/sample-postgres-1
kubectl delete -n demo pg/sample-postgres-2

kubectl delete -n demo repository/postgres-sample-postgres-1
kubectl delete -n demo repository/postgres-sample-postgres-2

kubectl delete -n demo backupblueprint/postgres-backup-blueprint
```

If you would like to uninstall Stash operator, please follow the steps [here](/docs/setup/README.md). -->
