#!/usr/bin/env bash
# Downloads the Loghub Apache 2k sample, pinned to a source commit, into data/raw/ (gitignored).
# Dataset terms: see docs/dataset-provenance.md and https://github.com/logpai/loghub/blob/master/LICENSE
set -euo pipefail

COMMIT="dd61d0952749ee7963bde24220d1be5ede023033"
BASE="https://raw.githubusercontent.com/logpai/loghub/${COMMIT}/Apache"
DEST="$(cd "$(dirname "$0")/.." && pwd)/data/raw"
mkdir -p "$DEST"

for f in Apache_2k.log Apache_2k.log_structured.csv; do
  echo "Fetching ${BASE}/${f}"
  curl -fsSL --retry 3 -o "${DEST}/${f}" "${BASE}/${f}"
done
curl -fsSL --retry 3 -o "${DEST}/LICENSE.loghub" "https://raw.githubusercontent.com/logpai/loghub/${COMMIT}/LICENSE"

echo "Source commit: ${COMMIT}"
( cd "$DEST" && sha256sum Apache_2k.log Apache_2k.log_structured.csv LICENSE.loghub | tee SHA256SUMS.txt )
