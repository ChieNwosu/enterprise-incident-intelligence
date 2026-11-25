# Dataset provenance: Loghub Apache 2k

| Item | Value |
| --- | --- |
| Project | Loghub, https://github.com/logpai/loghub |
| Pinned source commit | `dd61d0952749ee7963bde24220d1be5ede023033` |
| Raw log URL | https://raw.githubusercontent.com/logpai/loghub/dd61d0952749ee7963bde24220d1be5ede023033/Apache/Apache_2k.log |
| Reference CSV URL | https://raw.githubusercontent.com/logpai/loghub/dd61d0952749ee7963bde24220d1be5ede023033/Apache/Apache_2k.log_structured.csv |
| License URL | https://github.com/logpai/loghub/blob/master/LICENSE |
| Download script | `scripts/download_loghub_apache.sh` |
| Local location | `data/raw/` (gitignored, never committed) |

## SHA-256 (as downloaded, recorded by the script into `data/raw/SHA256SUMS.txt`)

```
c7efa3eb686e3a96bd2f8f4457b2a7887e9cf2f3649327f1b4e87af841363ce8  Apache_2k.log
54331d12eedf513f2127f4d89f0284c8b15fddfa5471103abf9db2c73d737778  Apache_2k.log_structured.csv
9ffa6ae259833cdc6e7ed8a0219fec72ef6455fb457b3c44096cb1ac1f0696d4  LICENSE.loghub
```

Because the URLs are pinned to a commit, a different checksum means the file was altered or corrupted.

## License terms (summary of the upstream LICENSE text at the pinned commit)

- The datasets are freely available for research or academic work.
- For any usage or distribution, refer to the loghub repository URL (https://github.com/logpai/loghub) and cite the Loghub paper where applicable.
- The license notice must be included in all copies of the datasets. The script saves a copy as `data/raw/LICENSE.loghub`.

This repository does not redistribute the dataset. This project is a learning exercise and should be treated as research or academic use. Anyone reusing the data must check the terms themselves.

## Citation

Jieming Zhu, Shilin He, Pinjia He, Jinyang Liu, Michael R. Lyu. Loghub: A Large Collection of System Log Datasets for AI-driven Log Analytics. IEEE ISSRE 2023.

## Observed raw format

Lines look like `[Sun Dec 04 04:47:44 2005] [notice] message`. The file uses CRLF line endings and the last line has no terminator (2000 lines, `wc -l` reports 1999). The parser strips a trailing CR.

## Reference fields

The CSV header is `LineId,Time,Level,Content,EventId,EventTemplate`. `EventId` and `EventTemplate` are log-pattern reference fields from a log parsing benchmark. They are not ground-truth incident escalation or anomaly labels, and this project does not use them as such. Day 1 compares only Time, Level and Content.
