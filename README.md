> **Ziggiz fork, temporary.** This is a patched build of [glalanne/provider-databricks](https://github.com/glalanne/provider-databricks) v2.5.3 with one change: `databricks_budget_policy` uses `policy_id` as its external name (upstream [issue #142](https://github.com/glalanne/provider-databricks/issues/142), [PR #143](https://github.com/glalanne/provider-databricks/pull/143)). `.github/workflows/ziggiz-publish.yml` pushes the multi-arch package to our ECR. **Exit:** when upstream releases the fix, pin ziggiz-platform's `charts/provider-databricks/templates/provider.yaml` to the upstream package by digest and archive this fork.


# Provider Databricks

<div style="text-align: center;">

![CI](https://github.com/glalanne/provider-databricks/workflows/CI/badge.svg)
[![GitHub release](https://img.shields.io/github/release/glalanne/provider-databricks/all.svg)](https://github.com/glalanne/provider-databricks/releases)
[![Contributors](https://img.shields.io/github/contributors/glalanne/provider-databricks)](https://github.com/glalanne/provider-databricks/graphs/contributors)


</div>

`provider-databricks` is a [Crossplane](https://crossplane.io/) provider that
is built using [Upjet](https://github.com/crossplane/upjet) code
generation tools and exposes XRM-conformant managed resources for the
[Databricks](https://registry.terraform.io/providers/databricks/databricks/latest/docs).


Most of the testing have been done on [Azure Databricks](https://azure.microsoft.com/en-ca/products/databricks)

## Getting Started

Install the provider by using the following command after changing the image tag
to the [latest release](https://marketplace.upbound.io/providers/lalanne/provider-databricks):
```
up ctp provider install xpkg.upbound.io/lalanne/provider-databricks:v2.5.0
```

Alternatively, you can use declarative installation:
```
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-databricks
spec:
  package: xpkg.upbound.io/lalanne/provider-databricks:v2.5.0
EOF
```

Notice that in this example Provider resource is referencing ControllerConfig with debug enabled.

You can see the API reference [here](https://doc.crds.dev/github.com/glalanne/provider-databricks).

## Exemples

A few exemples are provided [here](./examples/) to show case how to configure the provider, and how to use some resources


## Developing

Run code-generation pipeline:
```console
make generate
```

Run against a Kubernetes cluster:

```console
make run
```

Build, push, and install:

```console
make all
```

Build binary:

```console
make build
```

Local deployment:

```console
make local-deploy
```

End to end tests:

```console
# Ensure that you have a .datasource.yaml use the sample for instructions
# Ensure that you have a UPTEST_CLOUD_CREDENTIALS env define with the creds you want to use.

make e2e
```

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/glalanne/provider-databricks/issues).
