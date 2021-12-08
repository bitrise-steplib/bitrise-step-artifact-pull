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
