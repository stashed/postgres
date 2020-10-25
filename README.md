[![Go Report Card](https://goreportcard.com/badge/stash.appscode.dev/postgres)](https://goreportcard.com/report/stash.appscode.dev/postgres)
![CI](https://github.com/stashed/postgres/workflows/CI/badge.svg)
[![Docker Pulls](https://img.shields.io/docker/pulls/stashed/stash-postgres.svg)](https://hub.docker.com/r/stashed/stash-postgres/)
[![Slack](https://slack.appscode.com/badge.svg)](https://slack.appscode.com)
[![Twitter](https://img.shields.io/twitter/follow/kubestash.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=KubeStash)

# Postgres

Postgres backup and restore plugin for [Stash by AppsCode](https://stash.run).

## Install

Install PostgreSQL 11.2 backup or restore plugin for Stash as below.

```console
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/stash-postgres --name=stash-postgres-11.2 --version=11.2
```

To install catalog for all supported PostgreSQL versions, please visit [here](https://github.com/stashed/catalog).

## Uninstall

Uninstall PostgreSQL 11.2 backup or restore plugin for Stash as below.

```console
helm delete stash-postgres-11.2
```

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product annoucements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
