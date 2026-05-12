#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:9101}"

printf 'Checking health endpoint...\n'
curl --fail --silent --show-error "${BASE_URL}/healthz"
printf '\n\nChecking model endpoint...\n'
curl --fail --silent --show-error "${BASE_URL}/api/v1/models"
printf '\n\nChecking text-to-image validation...\n'
text_status="$(curl --silent --show-error --output /tmp/image_generation_text_validation.json --write-out '%{http_code}' \
  -X POST "${BASE_URL}/api/v1/images/generations" \
  -H 'Content-Type: application/json' \
  -d '{}')"
if [[ "${text_status}" != "400" ]]; then
  printf 'Unexpected status for text-to-image validation: %s\n' "${text_status}"
  cat /tmp/image_generation_text_validation.json
  exit 1
fi
cat /tmp/image_generation_text_validation.json
printf '\n\nChecking image-to-image validation...\n'
i2i_status="$(curl --silent --show-error --output /tmp/image_generation_i2i_validation.json --write-out '%{http_code}' \
  -X POST "${BASE_URL}/api/v1/images/edits" \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"test"}')"
if [[ "${i2i_status}" != "400" ]]; then
  printf 'Unexpected status for image-to-image validation: %s\n' "${i2i_status}"
  cat /tmp/image_generation_i2i_validation.json
  exit 1
fi
cat /tmp/image_generation_i2i_validation.json
printf '\n\nSmoke test completed.\n'
