Please note that this step is configured to be run in Bitrise's internal environment. Running the step, including its E2E tests, is not supported for external developers.

### E2E Tests

To be able to run tests locally, create `.bitrise.secrets.yml` with the contents of the Bitrise internal LastPass secret `bitrise-step-artifact-pull-secrets`.

Tests use the following staging Bitrise app https://app-staging.bitrise.io/app/11abc8954aa46c5a
