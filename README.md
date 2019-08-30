# Kluster

[![CircleCI](https://circleci.com/gh/jakub-gawlas/kluster.svg?style=svg)](https://circleci.com/gh/jakub-gawlas/kluster)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=jakub-gawlas_kluster&metric=alert_status)](https://sonarcloud.io/dashboard?id=jakub-gawlas_kluster)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=jakub-gawlas_kluster&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=jakub-gawlas_kluster)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=jakub-gawlas_kluster&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=jakub-gawlas_kluster)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=jakub-gawlas_kluster&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=jakub-gawlas_kluster)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=jakub-gawlas_kluster&metric=bugs)](https://sonarcloud.io/dashboard?id=jakub-gawlas_kluster)

Provision local Kubernetes cluster with application stack defined as a code.

## Prerequisites

- Docker
- Helm
- kubectl

## Installation

```sh
go get github.com/jakub-gawlas/kluster
```

## Config

Example config file (by default is loaded from `cluster.yaml`):

```yaml
kind: Cluster // constant value
apiVersion: kind.sigs.k8s.io/v1alpha3 // constant value
name: test // cluster name
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30001
        hostPort: 3001
charts:
  - name: test
    path: helm/test
    apps:
      - name: worker
        dockerfile: worker.Dockerfile
```

Application expects

## CLI

### deploy

Provisions new local cluster with using KinD (Kubernetes in Docker) if not already exists.
Builds defined docker images and install helm charts.

Flags:
- **config** *(optional)* - path to config file (default `cluster.yaml`)

```sh
kluster deploy
```

### destroy

Deletes existing local cluster.

Flags:
- **config** *(optional)* - path to config file (default `cluster.yaml`)

```sh
kluster deploy
```

### kubectl

Executes kubectl command on cluster.

```sh
kluster kubectl get pod
```