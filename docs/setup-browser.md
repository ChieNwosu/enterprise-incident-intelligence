# Browser-based setup (GitHub Codespaces)

No local install is needed.

1. Open the repository on GitHub, choose **Code**, then the **Codespaces** tab, then **Create codespace on main** (or on your branch).
2. Wait for the container to build. It uses `.devcontainer/devcontainer.json` (Go 1.25 image) and runs `go version` afterwards.
3. In the terminal, verify the toolchain:
   ```bash
   go vet ./... && go test -race ./... && go build ./...
   ```
4. Fetch the sample dataset (about 2000 log lines, pinned commit):
   ```bash
   bash scripts/download_loghub_apache.sh
   ```
5. Run the parser:
   ```bash
   go run ./cmd/eii -input data/raw/Apache_2k.log \
     -reference data/raw/Apache_2k.log_structured.csv
   ```
   The line `input=N parsed=N unparsed=N` goes to stderr, JSON to stdout.

Notes: `data/` is gitignored, so downloaded datasets cannot be committed by accident. Codespaces needs network access to `raw.githubusercontent.com` for step 4. Codespaces usage may count against your GitHub plan quota.
