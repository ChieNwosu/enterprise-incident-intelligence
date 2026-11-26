# CASE-001: parser stops with "input record exceeds maximum line size"

## Symptom

`eii` exits with code 1 and prints:

```
fatal: read failed at line 2: input record exceeds maximum line size (read 1 lines before failure; no summary written)
```

## Environment

Go 1.25.1 linux/amd64, `eii` built from this repo, default `-max-line-bytes` (1 MiB).

## Reproduce (synthetic input)

```bash
mkdir -p data/tmp
printf '[Mon Jan 02 10:00:00 2023] [notice] ok\n[Mon Jan 02 10:00:01 2023] [error] %s\n' \
  "$(head -c 2000000 /dev/zero | tr '\0' 'A')" > data/tmp/big.log
go build -o data/tmp/eii ./cmd/eii
data/tmp/eii -input data/tmp/big.log; echo "exit=$?"
```

Observed: the message above and `exit=1`. No JSON is written.

## Diagnosis

Line 2 is about 2 MB, over the 1 MiB limit. `bufio.Scanner` reports `bufio.ErrTooLong`; the parser maps it to `ErrLineTooLong` and fails the run. Failing is deliberate: continuing would drop or truncate the record and make counts misleading. This is distinct from a malformed record, which is counted as unparsed and exits 0.

## Resolution

If the long record is legitimate, raise the limit:

```bash
data/tmp/eii -input data/tmp/big.log -max-line-bytes 4000000
```

Observed: `input=2 parsed=2 unparsed=0`, exit 0. If it is not legitimate (for example a binary blob or a missing newline), fix the producer. Clean up with `rm -rf data/tmp`.

## Related check: malformed lines

`internal/apachelog/testdata/malformed.log` (synthetic) gives `input=6 parsed=2 unparsed=4` with unparsed line numbers 2, 3, 5, 6 and exit 0. With `-strict` the exit code is 3.

## Follow-up

Evergreen note: [bufio-scanner-token-limit](../../../knowledge/evergreen/bufio-scanner-token-limit.md).
