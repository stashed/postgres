
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

**Create Postgres crd:**

Below is the YAML of a sample Postgres crd that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: sample-postgres
  namespace: demo
spec:
  version: "9.6-v4"
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

Create the above Postgres crd,

```console
$ kubectl apply -f ./docs/examples/backup/postgres.yaml
postgres.kubedb.com/sample-postgres created
```

KubeDB will deploy a PostgreSQL database according to above specification. It will also create necessary secrets, services to access the database.

### Prepare Backend

## Restore PostgreSQL

## Cleanup