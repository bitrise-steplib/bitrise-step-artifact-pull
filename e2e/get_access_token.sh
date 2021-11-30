#!/bin/bash

set -e

json_response=$(curl -X POST https://auth.services.bitrise.dev/auth/realms/bitrise-services/protocol/openid-connect/token -k \
    --data "client_id=artifact-pull" \
    --data "client_secret=$ARTIFACT_PULL_AUTH_CLIENT_SECRET" \
    --data "grant_type=urn:ietf:params:oauth:grant-type:uma-ticket" \
    --data "scope=build_artifact:read build:read app:read" \
    --data "claim_token=eyJidWlsZF9pZHMiOlsiNzI5ZDdkZjctYTdjMy00Zjk3LWE5ZDAtYWNhMjM4OGNlZDMxIiwiM2ZiYzJmMTEtNzBjMi00YTI5LWJkMWEtNWU0ZGI2MzcwZGEwIiwiNTJlODVkOGUtMGMzZC00MzZhLTllNmItNTdkMjZkOWYwMzFmIiwiNGQ4ZDIxZjItMzY0NC00ZjdhLWJlNWEtNDZmMTBmMTU1YjQxIl0sInBpcGVsaW5lX2lkIjpbIjlkYTg2N2E5LTE5M2UtNDFiMS1iZjdmLTU4YTJlZDc0NTQ0MiJdICAgICAgICB9" \
    --data "claim_token_format=urn:ietf:params:oauth:token-type:jwt" \
    --data "audience=bitrise-api")

acces_token=$(echo "$json_response" | jq .access_token)

envman add --key BITRISEIO_ARTIFACT_PULL_TOKEN --value "$access_token"
echo "$access_token"
