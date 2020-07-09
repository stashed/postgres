# stash-postgres

[stash-postgres](https://github.com/stashed/postgres) - PostgreSQL database backup/restore plugin for [Stash by AppsCode](https://stash.run)

## TL;DR;

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install stash-postgres-11.1-beta.20200709 appscode/stash-postgres -n kube-system --version=11.1-beta.20200709
```

## Introduction

This chart deploys necessary `Function` and `Task` definition to backup or restore PostgreSQL 11.1 using Stash on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.11+

## Installing the Chart

To install the chart with the release name `stash-postgres-11.1-beta.20200709`:

```console
$ helm install stash-postgres-11.1-beta.20200709 appscode/stash-postgres -n kube-system --version=11.1-beta.20200709
```

The command deploys necessary `Function` and `Task` definition to backup or restore PostgreSQL 11.1 using Stash on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `stash-postgres-11.1-beta.20200709`:

```console
$ helm delete stash-postgres-11.1-beta.20200709 -n kube-system
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `stash-postgres` chart and their default values.

|    Parameter     |                                                           Description                                                            |     Default      |
|------------------|----------------------------------------------------------------------------------------------------------------------------------|------------------|
| nameOverride     | Overrides name template                                                                                                          | `""`             |
| fullnameOverride | Overrides fullname template                                                                                                      | `""`             |
| image.registry   | Docker registry used to pull Postgres addon image                                                                                | `stashed`        |
| image.repository | Docker image used to backup/restore PosegreSQL database                                                                          | `stash-postgres` |
| image.tag        | Tag of the image that is used to backup/restore PostgreSQL database. This is usually same as the database version it can backup. | `"11.1"`         |
| backup.cmd       | Postgres dump command, can either be: pg_dumpall  or pg_dump                                                                     | `"pg_dumpall"`   |
| backup.args      | Arguments to pass to `backup.cmd` command during backup process                                                                  | `""`             |
| restore.args     | Arguments to pass to `psql` command during restore process                                                                       | `""`             |
| waitTimeout      | Number of seconds to wait for the database to be ready before backup/restore process.                                            | `300`            |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```console
$ helm install stash-postgres-11.1-beta.20200709 appscode/stash-postgres -n kube-system --version=11.1-beta.20200709 --set image.registry=stashed
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while
installing the chart. For example:

```console
$ helm install stash-postgres-11.1-beta.20200709 appscode/stash-postgres -n kube-system --version=11.1-beta.20200709 --values values.yaml
```
