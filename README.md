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

##### Basic step config

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

##### Wildcard based artifact pull

During a pipeline, workflows receive the finished stages and workflows object. Developers can find it on a build VM's environment variable: `BITRISEIO_FINISHED_STAGES`.

Let's suppose that we get the following JSON object about the previously finished stages and workflows.

```json
[
  {
    "id": "083aa861-55b1-4132-ba70-0dfcd48fe929",
    "name": "stage-1",
    "workflows": [
      {
        "external_id": "73d33fb5-35c6-495f-bd80-015ae681db33",
        "finished_at": "2021-12-07T14:04:45Z",
        "id": "b1c6f0a1-06e7-4f63-a172-ac541a467d71",
        "name": "placeholder",
        "started_at": "2021-12-07T14:04:27Z",
        "status": "succeeded"
      },
      {
        "external_id": "39404bee-52ba-4ca2-8508-91489e7f6afa",
        "finished_at": "2021-12-07T14:05:07Z",
        "id": "f3bda7bb-37be-409f-9291-b377717cba60",
        "name": "textfile_generator",
        "started_at": "2021-12-07T14:04:48Z",
        "status": "succeeded"
      }
    ]
  },
  {
    "id": "4919fe0e-877a-45ca-ab25-7da2ddf54bce",
    "name": "stage-2",
    "workflows": [
      {
        "external_id": "ed0da0cf-66cc-4109-b23f-8a156d61b0c3",
        "finished_at": "2021-12-07T14:06:41Z",
        "id": "f572ca4e-2f06-40f1-a4cf-c208af15ff28",
        "name": "deployer",
        "started_at": "2021-12-07T14:06:13Z",
        "status": "succeeded"
      },
      {
        "external_id": "05130ce4-825b-4ca1-a9be-4f54413e5dcd",
        "finished_at": "2021-12-07T14:07:04Z",
        "id": "861fd1be-48b1-4a6b-ae4c-ee5449eaa6b6",
        "name": "textfile_generator",
        "started_at": "2021-12-07T14:06:45Z",
        "status": "succeeded"
      }
    ]
  }
]
```

As the key names in the object are self-describing, we will not cover those names except the `external_id`. The `external_id` is the build's slug in the PipelineService context.

Let's see the following use-cases, the use cases first part is the demand, the second is the `artifact_sources` config:

- As a developer, I would like to get the build artifact(s) of the _stage-1_'s _placeholder_'s workflow: `stage-1.placeholder`.

- As a developer, I would like to get the build artifact(s) of the _stage-2_'s _deployer_'s workflow and the _stage-1_'s _placeholder_'s workflow: `stage-1.placeholder,stage-2.deployer`. The two expressions are separated by a comma.

- As a developer, I would like to retrieve already generated artifacts: `*` or `"" (empty string)`. As the example shows, developers can use wildcard expressions.

- As a developer, I would like to retrieve the generated artifacts from the _stage-2_ stage: `stage-2.*`.

- As a developer, I would like to get the _textfile_generator_ workflow artifacts: `*.textfile_generator`

And so on. The syntax is: `{stage-name}.{workflow-name}`.

The results will be in the `$BITRISE_ARTIFACT_PATHS` env. var. The list is delimited with a `|` pipe character.

```bash
$BITRISE_ARTIFACT_PATHS = /var/folders/sd/lvn5cp9x5dn_xh1vhfgjjjw40000gp/T/_artifact_pull3010595419/generated_text_file.txt|/var/folders/sd/lvn5cp9x5dn_xh1vhfgjjjw40000gp/T/_artifact_pull3010595419/app-release-unsigned.apk
```


## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `verbose` | Enable logging additional information for debugging | required | `false` |
| `artifact_sources` | A comma separated list of workflows and stage paths, which can generate artifacts. You need to use the `{stage}.{workflow}` syntax. The "dot" character is the delimiter between the stage and the workflow. You can use wildcards in the expression. If you leave it empty, the default value will be the "*" (star), which means, it will get every artifact from every workflow. |  | `*` |
| `finished_stage` | This is a JSON representation of the finished stages for which the step can download build artifacts. | required | `$BITRISEIO_FINISHED_STAGES` |
| `bitrise_api_base_url` | The base URL of the Bitrise API used to process the download requests. | required | `https://api.bitrise.io` |
| `bitrise_api_access_token` | The OAuth access token that authorizes to call the Bitrise API. | sensitive | `$BITRISEIO_ARTIFACT_PULL_TOKEN` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `BITRISE_ARTIFACT_PATHS` | An absolute path list of the downloaded artifacts. The list is separated with pipe (\|) characters. |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/steps-artifact-pull/pulls) and [issues](https://github.com/bitrise-steplib/steps-artifact-pull/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/bitrise-cli/run-your-first-build/).

Note: this step's end-to-end tests (defined in e2e/bitrise.yml) are working with secrets which are intentionally not stored in this repo. External contributors won't be able to run those tests. Don't worry, if you open a PR with your contribution, we will help with running tests and make sure that they pass.

Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/contributors/create-your-own-step/)
- [Testing your Step](https://devcenter.bitrise.io/contributors/testing-and-versioning-your-steps/)
