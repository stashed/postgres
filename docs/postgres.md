---
title: Backup PostgreSQL | Stash
description: Backup PostgreSQL database using Stash
menu:
  product_stash_0.8.3:
    identifier: database-postgres
    name: PostgreSQL
    parent: database
    weight: 20
product_name: stash
menu_name: product_stash_0.8.3
section_menu_id: guides
---

# Backup and Restore PostgreSQL database using Stash

Stash supports backup and restores PostgreSQL database. This guide will show you how you can backup and restore your PostgreSQL database with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.

- Install Stash in your cluster following the steps [here](https://appscode.com/products/stash/0.8.3/setup/install/).

- Install [KubeDB](https://kubedb.com) in your cluster following the steps [here](https://kubedb.com/docs/0.12.0/setup/install/).

- If you are not familiar with how Stash backup and restore databases, please check the following guide:
  - [How Stash backup and restore databases](https://appscode.com/products/stash/0.8.3/guides/databases/overview/).

You have to be familiar with following custom resources:

- [AppBinding](https://appscode.com/products/stash/0.8.3/concepts/crds/appbinding/)
- [Function](https://appscode.com/products/stash/0.8.3/concepts/crds/function/)
- [Task](https://appscode.com/products/stash/0.8.3/concepts/crds/task/)
- [BackupConfiguration](https://appscode.com/products/stash/0.8.3/concepts/crds/backupconfiguration/)
- [RestoreSession](https://appscode.com/products/stash/0.8.3/concepts/crds/restoresession/)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```console
$ kubectl create ns demo
namespace/demo created
```

>Note: YAML files used in this tutorial are stored [here](https://github.com/stashed/postgres/examples/).

## Install Postgres Catalog for Stash

At first, we have to install Postgres plugin (`postgres-catalog`) for Stash. This plugin creates necessary `Function` and `Task` definition which is used by Stash to backup or restore a PostgreSQL database. We are going to use [Helm](https://helm.sh/) to install `postgres-catalog` chart.

If you have already installed `stash-catalog` which contains necessary `Function` and `Task` definition to backup or restore all the databases supported by Stash, you can skip installing `postgres-catalog`.

Let's install `postgres-catalog` chart,

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/postgres-catalog --name postgres-catalog
```

Once installed, this will create `pg-backup-*` and `pg-recovery-*` Functions for all supported PostgreSQL versions. Verify that the Functions has been created successfully by,

```console
$ kubectl get function
NAME             AGE
pg-backup-9.6    6s
pg-backup-10.2   6s
pg-backup-10.6   6s
pg-backup-11.1   6s
pg-backup-11.2   6s
pg-restore-9.6   6s
pg-restore-10.2  6s
pg-restore-10.6  6s
pg-restore-11.1  6s
pg-restore-11.2  6s
update-status    6d19h
```

This will also create `pg-backup-*` and `pg-restore-*` Tasks for all supported PostgreSQL versions. Verify that they have been created successfully by,

```console
$ kubectl get task
NAME             AGE
NAME             AGE
pg-backup-9.6    10s
pg-backup-10.2   10s
pg-backup-10.6   10s
pg-backup-11.1   10s
pg-backup-11.2   10s
pg-restore-9.6   10s
pg-restore-10.2  10s
pg-restore-10.6  10s
pg-restore-11.1  10s
pg-restore-11.2  10s
```

Now, Stash is ready to backup PostgreSQL database.

## Backup PostgreSQL

This section will demonstrate how to backup PostgreSQL databse. We are going to use [KubeDB](https://kubedb.com) to deploy a sample database. You can deploy your database using any method you want. We are using `KubeDB` because it automates some tasks that you have to do manually otherwise.

### Deploy Sample PosgreSQL Database

Let's deploy a sample PostgreSQL database and insert some data into it.

**Create Postgres CRD:**

Below is the YAML of a sample Postgres crd that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: sample-postgres
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

Create the above `Postgres` crd,

```console
$ kubectl apply -f ./docs/examples/backup/postgres.yaml
postgres.kubedb.com/sample-postgres created
```

KubeDB will deploy a PostgreSQL database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get pg -n demo sample-postgres
NAME              VERSION   STATUS    AGE
sample-postgres   11.2      Running   3m11s
```

The database is `Running`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```console
$ kubectl get secret -n demo -l=kubedb.com/name=sample-postgres
NAME                   TYPE     DATA   AGE
sample-postgres-auth   Opaque   2      27h

$ kubectl get service -n demo -l=kubedb.com/name=sample-postgres
NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
sample-postgres            ClusterIP   10.106.147.155   <none>        5432/TCP   22h
sample-postgres-replicas   ClusterIP   10.96.231.122    <none>        5432/TCP   22h
```

Here, we have to use service `sample-postgres` and secret `sample-postgres-auth` to connect with the database. KubeDB creates an [AppBinding](https://appscode.com/products/stash/0.8.3/concepts/crds/appbinding/) crd that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```console
$ kubectl get appbindings -n demo
NAME              AGE
sample-postgres   20m
```

Let's check the YAML of the above `AppBinding`,

```console
$ kubectl get appbindings -n demo sample-postgres -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
...
  name: sample-postgres
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-postgres
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgres
    app.kubernetes.io/version: "11.2"
    kubedb.com/kind: Postgres
    kubedb.com/name: sample-postgres
...
spec:
  clientConfig:
    service:
      name: sample-postgres
      path: /
      port: 5432
      query: sslmode=disable
      scheme: postgresql
  secret:
    name: sample-postgres-auth
  secretTransforms:
  - renameKey:
      from: POSTGRES_USER
      to: username
  - renameKey:
      from: POSTGRES_PASSWORD
      to: password
  type: kubedb.com/postgres
```

Stash uses the `AppBinding` crd to connect with the target database. It requires the following two fields to set in AppBinding's `Spec` section.

- `spec.clientConfig.service.name` specifies the name of the service that connects to the database.
- `spec.secret` specifies the name of the secret that holds necessary credentials to access the database.

**Creating AppBinding Manually:**

If you deploy PostgreSQL database without KubeDB, you have to create the AppBinding crd manually in the same namespace as the service and secret of the database.

The following YAML shows a minimal AppBinding specification that you have to create if you deploy PostgreSQL database without KubeDB.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: my-custom-appbinding
  namespace: my-database-namespace
spec:
  clientConfig:
    service:
      name: my-database-service
      port: 5432
  secret:
    name: my-database-credentials-secret
```

**Insert Sample Data:**

Now, we will exec into the database pod and create some sample data. At first, find out the database pod using the following command,

```console
$ kubectl get pods -n demo --selector="kubedb.com/name=sample-postgres"
NAME                READY   STATUS    RESTARTS   AGE
sample-postgres-0   1/1     Running   0          8m58s
```

Now, let's exec into the pod and create a table,

```console
$ kubectl exec -it -n demo sample-postgres-0 sh
# login as "postgres" superuser.
/ # psql -U postgres
psql (11.2)
Type "help" for help.

# list available databases
postgres=# \l
                                 List of databases
   Name    |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
-----------+----------+----------+------------+------------+-----------------------
 postgres  | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0 | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
           |          |          |            |            | postgres=CTc/postgres
 template1 | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
           |          |          |            |            | postgres=CTc/postgres
(3 rows)

# connect to "postgres" database
postgres=# \c postgres
You are now connected to database "postgres" as user "postgres".

# create a table
postgres=# CREATE TABLE COMPANY( NAME TEXT NOT NULL, EMPLOYEE INT NOT NULL);
CREATE TABLE

# list tables
postgres=# \d
          List of relations
 Schema |  Name   | Type  |  Owner   
--------+---------+-------+----------
 public | company | table | postgres
(1 row)

# quit from the database
postgres=# \q

# exit from the pod
/ # exit
```

Now, we are ready to backup this sample database.

### Prepare Backend

We are going to store our backed up data into a GCS bucket. At first, we need to create a secret with GCS credentials then we need to create a `Repository` crd. If you want to use a different backend, please read the respective backend configuration doc from [here](https://appscode.com/products/stash/0.8.3/guides/backends/overview/).

**Create Storage Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```console
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat downloaded-sa-json.key > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

**Create Repository:**

Now, crete a `Respository` using this secret. Below is the YAML of Repository crd we are going to create,

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcs-repo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: /demo/postgres/sample-postgres
    storageSecretName: gcs-secret
```

Let's create the `Repository` we have shown above,

```console
$ kubectl apply -f ./docs/examples/backup/repository.yaml
repository.stash.appscode.com/gcs-repo created
```

Now, we are ready to backup our database to our desired backend.

### Backup

We have to create a `BackupConfiguration` targeting respective AppBinding crd of our desired database. Then Stash will create a CronJob to periodically backup the database.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` crd to backup the `sample-postgres` database we have deployed earlier.,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  task:
    name: pg-backup-11.2
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-postgres
  retentionPolicy:
    keepLast: 5
    prune: true
```

Here,

- `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
- `spec.task.name` specifies the name of the task crd that specifies the necessary Function and their execution order to backup a PostgreSQL databse.
- `spec.target.ref` refers to the `AppBinding` crd that was created for `sample-postgres` database.

Let's create the `BackupConfiguration` crd we have shown above,

```console
$ kubectl apply -f ./docs/examples/backup/backupconfiguration.yaml
backupconfiguration.stash.appscode.com/sample-postgres-backup created
```

**Verify CronJob:**

If everything goes well, Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` crd.

Verify that the CronJob has been created using the following command,

```console
$ kubectl get cronjob -n demo
NAME                     SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-postgres-backup   */5 * * * *   False     0        <none>          61s
```

**Wait for BackupSession:**

The `sample-postgres-backup` CronJob will trigger a backup on each schedule by creating a `BackpSession` crd.

Wait for a schedule to appear. Run the following command to watch `BackupSession` crd,

```console
$ kubectl get backupsession -n demo -w
NAME                                BACKUPCONFIGURATION      PHASE       AGE
sample-postgres-backup-1560350521   sample-postgres-backup   Running     5m19s
sample-postgres-backup-1560350521   sample-postgres-backup   Succeeded   5m45s
```

We can see above that the backup session has succeeded. Now, we will verify that the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, Stash will update the respective `Repository` crd to reflect the backup. Check that the repository `gcs-repo` has been updated by the following command,

```console
$ kubectl get repository -n demo gcs-repo
NAME       INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo   true        3.441 KiB   1                31s                      17m
```

Now, if we navigate to the GCS bucket, we will see backed up data has been stored in `demo/postgres/sample-postgres` directory as specified by `spec.backend.gcs.prefix` field of Repository crd.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/images/sample-postgres-backup.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

>Note: Stash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

## Restore PostgreSQL

We will restore the database from the backup we have taken in the previous section. We will deploy a new database and initialize it from the backup.

**Deploy Restored Database:**

Now, we have to deploy the restored database similarly as we have deployed the original `sample-psotgres` database. However, this time there will be the following differences:

- We have to use the same secret that was used in the original database. We will specify it using `spec.databaseSecret` field.
- We have to specify `spec.init` section to tell KubeDB that we are going to use Stash to initialize this database from backup. KubeDB will keep the database phase to `Initializing` until Stash finishes its initialization.

Below is the YAML for `Postgres` crd we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: restored-postgres
  namespace: demo
spec:
  version: "11.2"
  storageType: Durable
  databaseSecret:
    secretName: sample-postgres-auth # use same secret as original the database
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    stashRestoreSession:
      name: sample-postgres-restore
  terminationPolicy: Delete
```

Here,

- `spec.init.stashRestoreSession.name` specifies the `RestoreSession` crd name that we are going to use to restore this database.

Let's create the above database,

```console
$ kubectl apply -f ./docs/examples/restore/restored-postgres.yaml
postgres.kubedb.com/restored-postgres created
```

If you check the database status, you will see it is stuck in `Initializing` state.

```console
$ kubectl get pg -n demo restored-postgres
NAME                VERSION   STATUS         AGE
restored-postgres   11.2      Initializing   3m21s
```

**Create RestoreSession:**

Now, we need to create a `RestoreSession` crd pointing to the AppBinding for this restored database.

Check AppBinding has been created for the `restored-postgres` database using the following command,

```console
$ kubectl get appbindings -n demo restored-postgres
NAME                AGE
restored-postgres   9m59s
```

>If you are not using KubeDB to deploy database, create the AppBinding manually.

Below is the YAML for the `RestoreSession` crd that we are going to create to restore backed up data into `restored-postgres` database.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-postgres-restore
  namespace: demo
  labels:
    kubedb.com/kind: Postgres # this label is mandatory if you are using KubeDB to deploy the database.
spec:
  task:
    name: pg-restore-11.2
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-postgres
  rules:
  - snapshots: [latest]
```

Here,

- `metadata.labels` specifies a `kubedb.com/kind: Postgres` label that is used by KubeDB to watch this `RestoreSession`.
- `spec.task.name` specifies the name of the `Task` crd that specifies the Functions and their execution order to restore a PostgreSQL database.
- `spec.repository.name` specifies the `Repository` crd that holds the backend information where our backed up data has been stored.
- `spec.target.ref` refers to the AppBinding crd for the `restored-postgres` databse.
- `spec.rules` specifies that we are restoring from the latest backup snapshot of the database.

> **Warning:** Label `kubedb.com/kind: Postgres` is mandatory if you are uisng KubeDB to deploy the databse. Otherwise, the database will be stuck in `Initializing` state.

Let's create the `RestoreSession` crd we have shown above,

```console
$ kubectl apply -f ./docs/examples/restore/restoresession.yaml
restoresession.stash.appscode.com/sample-postgres-restore created
```

Once, you have created the `RestoreSession` crd, Stash will create a job to restore. We can watch the `RestoreSession` phase to check if the restore process is succeeded or not.

Run the following command to watch `RestoreSession` phase,

```console
$ kubectl get restoresession -n demo sample-postgres-restore -w
NAME                      REPOSITORY-NAME   PHASE       AGE
sample-postgres-restore   gcs-repo          Running     5s
sample-postgres-restore   gcs-repo          Succeeded   43s
```

So, we can see from the output of the above command that the restore process succeeded.

**Verify Restored Data:**

In this section, we will verify that the desired data has been restored successfully. We will connect to the database and check whether the table we had created in the original database is restored or not.

At first, check if the database has gone into `Running` state by the following command,

```console
$ kubectl get pg -n demo restored-postgres
NAME                VERSION   STATUS    AGE
restored-postgres   11.2      Running   2m16s
```

Now, find out the database pod by the following command,

```console
$ kubectl get pods -n demo --selector="kubedb.com/name=restored-postgres"
NAME                  READY   STATUS    RESTARTS   AGE
restored-postgres-0   1/1     Running   0          3m15s
```

Now, exec into the database pod and list available tables,

```console
$ kubectl exec -it -n demo restored-postgres-0 sh
# login as "postgres" superuser.
/ # psql -U postgres
psql (11.2)
Type "help" for help.

# list available databases
postgres=# \l
                                 List of databases
   Name    |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
-----------+----------+----------+------------+------------+-----------------------
 postgres  | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0 | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
           |          |          |            |            | postgres=CTc/postgres
 template1 | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
           |          |          |            |            | postgres=CTc/postgres
(3 rows)

# connect to "postgres" database
postgres=# \c postgres
You are now connected to database "postgres" as user "postgres".

# check the table we had created in the original database has been restored here
postgres=# \d
          List of relations
 Schema |  Name   | Type  |  Owner   
--------+---------+-------+----------
 public | company | table | postgres
(1 row)

# quit from the database
postgres=# \q

# exit from the pod
/ # exit
```

So, from the above output, we can see the table `company` that we had created in the original database `sample-postgres` is restored in the restored database `restored-postgres`.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete restoresession -n demo sample-postgres-restore
kubectl delete backupconfiguration -n demo sample-postgres-backup
kubectl delete pg -n demo restored-postgres
kubectl delete pg -n demo sample-postgres
```

To uninstall `postgres-catalog` chart, run the following command,

```console
helm delete postgres-catalog
```
