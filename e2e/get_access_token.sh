#!/bin/bash

set -ex

json_response=$(curl -X POST https://auth.services.bitrise.dev/auth/realms/bitrise-services/protocol/openid-connect/token -k \
    --data "client_id=artifact-pull" \
    --data "client_secret=$ARTIFACT_PULL_AUTH_CLIENT_SECRET" \
    --data "grant_type=urn:ietf:params:oauth:grant-type:uma-ticket" \
    --data "scope=build_artifact:read build:read app:read" \
    --data "claim_token=eyJidWlsZF9pZHMiOlsiNzNkMzNmYjUtMzVjNi00OTVmLWJkODAtMDE1YWU2ODFkYjMzIiwiMzk0MDRiZWUtNTJiYS00Y2EyLTg1MDgtOTE0ODllN2Y2YWZhIiwiZWQwZGEwY2YtNjZjYy00MTA5LWIyM2YtOGExNTZkNjFiMGMzIiwiMDUxMzBjZTQtODI1Yi00Y2ExLWE5YmUtNGY1NDQxM2U1ZGNkIl0sInBpcGVsaW5lX2lkIjpbIjM2ZTg1NDBkLTQxYzctNDNjZS05NzRiLTgyNTQ5OTc2OTkxZSJdfQ==" \
    --data "claim_token_format=urn:ietf:params:oauth:token-type:jwt" \
    --data "audience=bitrise-api")

auth_token=$(echo $json_response | jq -r .access_token)

envman add --key BITRISEIO_ARTIFACT_PULL_TOKEN --value $auth_token
