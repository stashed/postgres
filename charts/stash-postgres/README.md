# stash-postgres

[stash-postgres](https://github.com/stashed/stash-postgres) - PostgreSQL database backup/restore plugin for [Stash by AppsCode](https://appscode.com/products/stash/).

## TL;DR;

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install stash-postgres-10.6 appscode/stash-postgres -n kube-system --version=10.6
```

## Introduction

This chart installs necessary `Function` and `Task` definition to backup or restore PostgreSQL database 10.6 using Stash.

## Prerequisites

- Kubernetes 1.11+

## Installing the Chart

- Add AppsCode chart repository to your helm repository list,

```console
helm repo add appscode https://charts.appscode.com/stable/
```

- Update helm repositories to fetch latest charts from the remove repository,

```console
helm repo update
```

- Install the chart with the release name `stash-postgres-10.6` run the following command,

```console
helm install stash-postgres-10.6 appscode/stash-postgres -n kube-system --version=10.6
```

The above commands installs `Functions` and `Task` crds that are necessary to backup PostgreSQL database 10.6 using Stash.

## Uninstalling the Chart

To uninstall/delete the `stash-postgres-10.6` run the following command,

```console
helm uninstall stash-postgres-10.6 -n kube-system
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `stash-postgres` chart and their default values.

|     Parameter     |                                                           Description                                                            |     Default      |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `docker.registry` | Docker registry used to pull respective images                                                                                   | `stashed`        |
| `docker.image`    | Docker image used to backup/restore PosegreSQL database                                                                          | `stash-postgres` |
| `docker.tag`      | Tag of the image that is used to backup/restore PostgreSQL database. This is usually same as the database version it can backup. | `10.6`           |
| `backup.args`   | Optional arguments to pass to `pgdump` command  during bakcup process                                                            |                  |
| `restore.args`  | Optional arguments to pass to `psql` command during restore process                                                              |                  |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

For example:

```console
helm install stash-postgres-10.6 appscode/stash-postgres -n kube-system --set docker.registry=my-registry
```
