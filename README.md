<p align="center">
    <img width="200px" height=auto src="https://avatars.githubusercontent.com/u/90051847?s=280&v=4" />
</p>

<p align="center">
    <a href="https://github.com/opentdf/charts"><img src="https://badgen.net/github/stars/opentdf/charts?icon=github" /></a>
    <a href="https://github.com/opentdf/charts"><img src="https://badgen.net/github/forks/opentdf/charts?icon=github" /></a>
    <!-- <a href="https://artifacthub.io/packages/search?repo=opentdf"><img src="https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/opentdf" /></a> -->
    <a href="https://github.com/opentdf/charts/actions/workflows/chart-releaser.yaml"><img src="https://github.com/opentdf/charts/actions/workflows/chart-releaser.yaml/badge.svg" /></a>
</p>

# OpenTDF Helm Charts

## Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to
Helm's [documentation](https://helm.sh/docs) to get started

Once Helm has been set up correctly, add the repo as follows:

    helm repo add opentdf https://opentdf.github.io/charts

If you had already added this repo earlier, run `helm repo update` to retrieve
the latest versions of the packages.  You can then run `helm search repo
opentdf` to see the charts.

For chart specific documentation, please refer to the README.md files in the respective chart directories.

### Charts

- [Platform](charts/platform/README.md)

#### Contributing

When updating the charts, run `helm-docs` after to update the
README.md with the proper changes.
