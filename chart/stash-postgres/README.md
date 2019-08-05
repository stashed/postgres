# stash-postgres

[stash-postgres](https://github.com/stashed/stash-postgres) - PostgreSQL database backup/restore plugin for [Stash by AppsCode](https://appscode.com/products/stash/).

## TL;DR;

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/stash-postgres --name=stash-postgres-11.2 --version=11.2
```

## Introduction

This chart installs necessary `Function` and `Task` definition to backup or restore PostgreSQL database 11.2 using Stash.

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

- Install the chart with the release name `stash-postgres-11.2` run the following command,

```console
helm install appscode/stash-postgres --name=stash-postgres-11.2 --version=11.2
```

The above commands installs `Functions` and `Task` crds that are necessary to backup PostgreSQL database 11.2 using Stash.

## Uninstalling the Chart

To uninstall/delete the `stash-postgres-11.2` run the following command,

```console
helm delete stash-postgres-11.2
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `stash-postgres` chart and their default values.

|     Parameter     |                                                           Description                                                            |     Default      |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `docker.registry` | Docker registry used to pull respective images                                                                                   | `stashed`        |
| `docker.image`    | Docker image used to backup/restore PosegreSQL database                                                                          | `stash-postgres` |
| `docker.tag`      | Tag of the image that is used to backup/restore PostgreSQL database. This is usually same as the database version it can backup. | `11.2`           |
| `backup.pgArgs`   | Optional arguments to pass to `pgdump` command  during bakcup process                                                            |                  |
| `restore.pgArgs`  | Optional arguments to pass to `psql` command during restore process                                                              |                  |
| `metrics.enabled` | Specifies whether to send Prometheus metrics                                                                                     | `true`           |
| `metrics.labels`  | Optional comma separated labels to add to the Prometheus metrics                                                                 |                  |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

For example:

```console
helm install --name stash-postgres-11.2 --set metrics.enabled=false appscode/stash-postgres
```

**Tips:** Use escape character (`\`) while providing multiple comma-separated labels for `metrics.labels`.

```console
 helm install chart/stash-postgres --set metrics.labels="k1=v1\,k2=v2"
```
