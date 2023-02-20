#!/usr/bin/env bash
set -euo pipefail
#set -x

# Usage
usage() {
  cat <<EOF
Invalidates CloudFront distribution cache.

Usage:
  $0 <distribution-id>
Requirements:
  aws, jq
EOF
}

# echo to stderr
eecho() { echo "$@" 1>&2; }

# Error handling
trap 'onerror $LINENO "$@"' ERR
onerror() {
  eecho "[ERROR] Unexpected error at line $1: status=$?"
}

# Parse arguments
# cf. https://stackoverflow.com/questions/192249/how-do-i-parse-command-line-arguments-in-bash
POSITIONAL=()
while [[ $# -gt 0 ]]
do
  key="$1"
  case $key in
    *)
      POSITIONAL+=("$1")
      shift
      ;;
  esac
done

if [[ ${#POSITIONAL[@]} -ne 1 ]]; then
  usage
  exit 1
fi

DISTRIBUTION_ID="${POSITIONAL[0]}"

eecho "[INFO] Creating new invalidation..."

invalidation_id=$(aws cloudfront create-invalidation --distribution-id "${DISTRIBUTION_ID}" --path '/*' | jq -r '.Invalidation.Id')

# Check Cloudfront invalidation status every 10 seconds
eecho "[INFO] Waiting for the invalidation to complete..."
while true; do
  sleep 10;
  status=$(aws cloudfront get-invalidation --distribution-id "${DISTRIBUTION_ID}" --id "${invalidation_id}" | jq -r '.Invalidation.Status')
  if [[ "${status}" == "Completed" ]]; then
    break
  fi
done

eecho "[INFO] Finished."
