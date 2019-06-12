
# Backup and Restore PostgreSQL database using Stash

Stash supports backup and restore PostgreSQL database. This guide will show you how you can backup/restore your PostgreSQL database with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.

- Install Stash in your cluster following the steps [here](https://appscode.com/products/stash/0.8.3/setup/install/).

- Install [KubeDB](https://kubedb.com)(`Optional`) in your cluster following the steps [here](https://kubedb.com/docs/0.12.0/setup/install/).

- If you are not familiar with how Stash backup and restore databases, please check following guides:
  - [How Stash backup databases](https://appscode.com/products/stash/0.8.3/guides/databases/backup/).
  - [How Stash restore databases from a backup](https://appscode.com/products/stash/0.8.3/guides/databases/restore/).

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

## Install Postgres plugin for Stash

At first, we have to install Postgres plugin `postgres-stash` for Stash. This plugin creates necessary `Function` and `Task` definition which is used by Stash to backup/restore PostgreSQL database. We are going to use [Helm](https://helm.sh/) to install `postgres-stash` chart.

Let's install `postgres-stash` chart,

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install appscode/postgres-stash --name postgres-stash
```

Once installed, this will create `pg-backup` and `pg-recovery` Function. Verify that the Functions has been created successfully by,

```console
$ kubectl get function
NAME            AGE
pg-backup       3h7m
pg-restore      3h7m
update-status   3h7m
```

This will also create `pg-backup` and `pg-restore` Task. Verify that they have been created successfully by,

```console
$ kubectl get task
NAME         AGE
pg-backup    3h9m
pg-restore   3h9m
```

Now, Stash is ready to backup PostgreSQL database.

## Backup PostgreSQL

This section will demonstrate how to backup PostgreSQL databse. We are going to use [KubeDB](https://kubedb.com) to deploy a sample database. You can deploy your database using any methond you want. We are using `KubeDB` because it automates some tasks that you have to do manually otherwise.

### Deploy Sample PosgreSQL Database

Let's deploy a sample PostgreSQL database and inset some data into it.

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

KubeDB will deploy a PostgreSQL database according to the above specification. It will also create necessary secrets, services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get pg -n demo sample-postgres
NAME              VERSION   STATUS    AGE
sample-postgres   11.2      Running   3m11s
```

The database is `Running`. Verify that KubeDB has created a Secret and a Service for this database using follwoing commands,

```console
$ kubectl get secret -n demo -l=kubedb.com/name=sample-postgres
NAME                   TYPE     DATA   AGE
sample-postgres-auth   Opaque   2      27h

$ kubectl get service -n demo -l=kubedb.com/name=sample-postgres
NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
sample-postgres            ClusterIP   10.106.147.155   <none>        5432/TCP   22h
sample-postgres-replicas   ClusterIP   10.96.231.122    <none>        5432/TCP   22h
```

Here, we have to use service `sample-postgres` and secret `sample-postgres-auth` to connect with the database. KubeDB creates an [AppBinding](https://appscode.com/products/stash/0.8.3/concepts/crds/appbinding/) crd that holds these necessary information to connect with the database.

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

Stash uses the `AppBinding` crd to connect with the target database. It requires following two field to set in AppBinding's `Spec` section.

- `spec.clientConfig.service.name` specifies the name of the service that connect to the database.
- `spec.secret` specifies name of the secret that holds necessary credentials to access the database.

**Creating AppBinding Manually:**

If you deploy PostgreSQL database without KubeDB, you have to create the AppBinding crd manually in the same namespace as the service and secret.

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
  secret:
    name: my-database-credentials-secret
```

**Insert Sample Data:**

Now, we will exec into the database pod and create some sample data. At first find out the database pod using the following command,

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

**Create Storage Secret:**

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

```console
$ kubectl apply -f ./docs/examples/backup/repository.yaml
repository.stash.appscode.com/gcs-repo created
```

### Backup

**Create BackupConfiguration:**

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  task:
    name: pg-backup
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

```console
$ kubectl apply -f ./docs/examples/backup/backupconfiguration.yaml
backupconfiguration.stash.appscode.com/sample-postgres-backup created
```

**CronJob:**

```console
$ kubectl get cronjob -n demo
NAME                     SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-postgres-backup   */5 * * * *   False     0        <none>          61s
```

**BackupSession:**

```console
```

**Verify Backup:**

## Restore PostgreSQL

## Cleanup
