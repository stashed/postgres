# Postgres-stash

[Postgres-stash by AppsCode](https://github.com/stashed/postgres-stash) - PostgreSQL database backup/restore plugin for [Stash](https://github.com/stashed/).

## TL;DR;

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/postgres-stash --name postgres-stash
```

## Introduction

This chart installs necessary `Function` and `Task` crd to backup/restore PostgreSQL database using Stash.

## Prerequisites

- Kubernetes 1.9+

## Installing the Chart

To install the chart with the release name `postgres-stash`:

```console
helm install appscode/postgres-stash --name postgres-stash
```

The command installs `Functions` and `Task` crds that are necessary to backup PostgreSQL database using Stash.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `postgres-stash`:

```console
helm delete postgres-stash
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Postgre-stash chart and their default values.

|        Parameter         |                                                           Description                                                            |     Default      |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `global.registry`        | Docker registry used to pull respective images                                                                                   | `appscode`       |
| `global.image`           | Docker image used to backup/restore PosegreSQL database                                                                          | `postgres-stash` |
| `global.tag`             | Tag of the image that is used to backup/restore PostgreSQL database. This is usually same as the database version it can backup. | `9.x`            |
| `global.backup.pgArgs`   | Optional arguments to pass to `pgdump` command  while bakcup                                                                     |                  |
| `global.restore.pgArgs`  | Optional arguments to pass to `psql` command while restore                                                                       |                  |
| `global.metrics.enabled` | Specifies whether to send Prometheus metrics                                                                                     | `true`           |
| `global.metrics.labels`  | Optional comma separated labels to add the Prometheus metrics                                                                    |                  |

> We have declared all the configurable parameters as global parameter so that the parent chart can overwrite them.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

For example:

```console
helm install --name postgres-stash --set global.metrics.enabled=false appscode/postgres-stash
```

**Tips:** Use escape character (`\`) while providing multiple comma-separated labels for `global.metrics.labels`.

```console
 helm install chart/postgres-stash --set global.metrics.labels="k1=v1\,k2=v2"
```
