# Artifact pull

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/steps-artifact-pull?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/steps-artifact-pull/releases)

Step to pull artifacts of a pipeline

<details>
<summary>Description</summary>

The step downloads build artifacts of a pipeline to a local folder.

By default, all artifacts generated by any workflow of the pipeline are downloaded. This can be limited
by setting the `artifact_sources` input variable.

Please note that this step is designed to be executed on the CI only.
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://devcenter.bitrise.io/steps-and-workflows/steps-and-workflows-index/).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

#### Examples

```yaml
steps:
  - git::https://github.com/bitrise-steplib/bitrise-step-artifact-pull.git@main::
      title: Pull artifacts
      inputs:
      - verbose: true
      - artifact_sources: stage-1.*
```

Use the `artifact_sources` input variable to limit the downloads to a set of stages or workflows:
  - `stage1.workflow1` - Gets the artifacts from the stage1's workflow1.
  - `stage1.*` - Gets all artifacts from the stage1's workflows.
  - `*workflow1` - Gets the workflows' artifacts from all stages.
  - `*` - Gets every generated artifacts in the pipeline.


## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `verbose` | Enable logging additional information for debugging | required | `false` |
| `artifact_sources` | A comma separated list of workflows and stage paths, which can generate artifacts. You need to use the `{stage}.{workflow}` syntax. The "dot" character is the delimiter between the stage and the workflow. You can use wildcards in the expression. If you leave it empty, the default value will be the "*" (star), which means, it will get every artifact from every workflow. |  |  |
| `finished_stage` | This is a JSON representation of the finished stages for which the step can download build artifacts. | required | `$BITRISEIO_FINISHED_STAGES` |
| `bitrise_api_base_url` | The base URL of the Bitrise API used to process the download requests. | required | `https://api.bitrise.io` |
| `bitrise_api_access_token` | The OAuth access token that authorizes to call the Bitrise API. | sensitive | `$BITRISEIO_ARTIFACT_PULL_TOKEN` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `BITRISE_ARTIFACT_PATHS` | An absolute path list of the downloaded artifacts. The list is separated with newlines (\n) |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/steps-artifact-pull/pulls) and [issues](https://github.com/bitrise-steplib/steps-artifact-pull/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/bitrise-cli/run-your-first-build/).

Note: this step's end-to-end tests (defined in e2e/bitrise.yml) are working with secrets which are intentionally not stored in this repo. External contributors won't be able to run those tests. Don't worry, if you open a PR with your contribution, we will help with running tests and make sure that they pass.

Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/contributors/create-your-own-step/)
- [Testing your Step](https://devcenter.bitrise.io/contributors/testing-and-versioning-your-steps/)
