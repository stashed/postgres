# Postgres-stash

[postgres-stash by AppsCode](https://github.com/stashed/postgres-stash) - PostgreSQL database backup/restore plugin for [Stash](https://github.com/stashed/).

## TL;DR;

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/postgres-stash --name=postgres-stash-11.2 --version=11.2
```

## Introduction

This chart installs necessary `Function` and `Task` definition to backup or restore PostgreSQL database 11.2 using Stash.

## Prerequisites

- Kubernetes 1.9+

## Installing the Chart

- Add AppsCode chart repository to your helm repository list,

```console
helm repo add appscode https://charts.appscode.com/stable/
```

- Update helm repositories to fetch latest charts from the remove repository,

```console
helm repo update
```

- Install the chart with the release name `postgres-stash-11.2` run the following command,

```console
helm install appscode/postgres-stash --name=postgres-stash-11.2 --version=11.2
```

The above commands installs `Functions` and `Task` crds that are necessary to backup PostgreSQL database 11.2 using Stash.

## Uninstalling the Chart

To uninstall/delete the `postgres-stash-11.2` run the following command,

```console
helm delete postgres-stash-11.2
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `postgre-stash` chart and their default values.

|        Parameter         |                                                           Description                                                            |     Default      |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `global.registry`        | Docker registry used to pull respective images                                                                                   | `appscode`       |
| `global.image`           | Docker image used to backup/restore PosegreSQL database                                                                          | `postgres-stash` |
| `global.tag`             | Tag of the image that is used to backup/restore PostgreSQL database. This is usually same as the database version it can backup. | `11.2`           |
| `global.backup.pgArgs`   | Optional arguments to pass to `pgdump` command  while bakcup                                                                     |                  |
| `global.restore.pgArgs`  | Optional arguments to pass to `psql` command while restore                                                                       |                  |
| `global.metrics.enabled` | Specifies whether to send Prometheus metrics                                                                                     | `true`           |
| `global.metrics.labels`  | Optional comma separated labels to add to the Prometheus metrics                                                                 |                  |

> We have declared all the configurable parameters as global parameter so that the parent chart can overwrite them.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

For example:

```console
helm install --name postgres-stash-11.2 --set global.metrics.enabled=false appscode/postgres-stash
```

**Tips:** Use escape character (`\`) while providing multiple comma-separated labels for `global.metrics.labels`.

```console
 helm install chart/postgres-stash --set global.metrics.labels="k1=v1\,k2=v2"
```
