# relink

[![Release](https://github.com/USA-RedDragon/relink/actions/workflows/release.yaml/badge.svg)](https://github.com/USA-RedDragon/relink/actions/workflows/release.yaml) [![go.mod version](https://img.shields.io/github/go-mod/go-version/USA-RedDragon/relink.svg)](https://github.com/USA-RedDragon/relink) [![License](https://badgen.net/github/license/USA-RedDragon/relink)](https://github.com/USA-RedDragon/relink/blob/main/LICENSE) [![Release](https://img.shields.io/github/release/USA-RedDragon/relink.svg)](https://github.com/USA-RedDragon/relink/releases/) [![coverage](https://raw.githubusercontent.com/USA-RedDragon/relink/main/.github/badges/coverage.svg)](https://github.com/USA-RedDragon/relink/actions)

A simple utility to find duplicate files and replace them with hardlinks to save disk space. It recursively scans directories to identify files with identical content and creates hardlinks while preserving the original file attributes.

## Configuration

<!-- configulator:begin -->

| Key           | Type    | Default    | Environment   | Flag            | Description                                                                     |
|---------------|---------|------------|---------------|-----------------|---------------------------------------------------------------------------------|
| `log-level`   | string  | `info`     | `LOG_LEVEL`   | `--log-level`   | Logging level for the application. One of debug, info, warn, or error           |
| `source`      | string  |            | `SOURCE`      | `--source`      | Source directory to read the files from                                         |
| `target`      | string  |            | `TARGET`      | `--target`      | Target directory to write the relinked files to                                 |
| `hash-jobs`   | integer | `4`        | `HASH_JOBS`   | `--hash-jobs`   | Number of jobs to use for hashing files                                         |
| `buffer-size` | integer | `4096`     | `BUFFER_SIZE` | `--buffer-size` | Buffer size for file checksum operations in bytes                               |
| `cache-type`  | string  | `memory`   | `CACHE_TYPE`  | `--cache-type`  | Cache type to use for storing file hashes. One of memory or sqlite              |
| `cache-path`  | string  | `:memory:` | `CACHE_PATH`  | `--cache-path`  | Path to the SQLite database file for caching. Only used if cache-type is sqlite |

<!-- configulator:end -->
