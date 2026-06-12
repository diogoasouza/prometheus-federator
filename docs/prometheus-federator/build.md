# Build Process For Project Operators

As a type of [Project Operator](https://github.com/rancher/helm-project-operator), Prometheus Federator is primarily composed of three components:
- The Underlying Helm Chart (Build Dependency)
- The Project Operator Image (Build Dependency)
- The Project Operator Helm Chart (What A User Actually Deploys)

## Underlying Helm Chart

Prometheus Federator no longer compiles an underlying project-monitoring chart into the binary. Instead, the binary is configured at runtime with an approved chart reference using `--managed-chart-name`, `--managed-chart-repo`, and `--managed-chart-version` (or the equivalent Helm chart values).

## The Project Operator Image

To implement a Project Operator, Helm Project Operator expects a user to run the `operator.Init` command, which appears in Prometheus Federator's [`main.go`](../../main.go) as follows:

```go
operator.Init(ctx, f.Namespace, cfg, common.Options{
    OperatorOptions: common.OperatorOptions{
        HelmAPIVersion:   HelmAPIVersion,
        ReleaseName:      ReleaseName,
        SystemNamespaces: SystemNamespaces,
        Singleton:        true, // indicates only one HelmChart can be registered per project defined
    },
    RuntimeOptions: f.RuntimeOptions,
})
```

Once your [`main.go`](../../main.go) is ready to be built, you can run `./scripts/build`, which will run the underlying `go build` command and place the created binary in `bin/prometheus-federator`.

Once the binary has been created, it is then packaged into a container image in the [`scripts/package`](../../scripts/package) step, where we build the Dockerfile found in `packages/Dockerfile` to produce the final image.

## The Project Operator Helm Chart

This is the component that the average user is actually expected to directly deploy; it is also maintained in the `packages/` directory, like the Underlying Helm Chart.

As explained above, packages are a construct of any [rancher/charts-build-scripts](https://github.com/rancher/charts-build-scripts) repository (see [the docs on Packages](https://github.com/rancher/charts-build-scripts/blob/master/templates/template/docs/packages.md) for more information), so just like with the Underlying Helm Chart, it is expected that a developer who files a PR with changes will run the `make charts` command to ensure that the package is read by the `rancher/charts-build-scripts` binary to **produce / auto-generate** the Helm Charts and manage the `assets/`+`charts/` directories as well as the `index.yaml` entries to introduce this package in a standard Helm repository fashion.

Once `make charts` has been run and the chart is built from `packages/prometheus-federator` -> `charts/prometheus-federator/${VERSION}` (part of the `make charts` command), the chart is now visible on the Helm repository maintained within your fork!

## TLDR; Putting It All Together

Therefore, the build process now looks as follows:
- Build the operator binary with `./scripts/build`
- Build the operator image with `./scripts/package`
- Configure the deployment chart with an approved runtime chart reference for project monitoring
