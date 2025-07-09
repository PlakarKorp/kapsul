# Kapsul

**Kapsul** is a tool that allows you to run most `plakar` sub-commands directly on `.ptar` archives. It transparently mounts the archive as an in-memory, read-only Plakar repository, enabling inspection, restoration, and diffing operations without extracting the archive.

## Features

- Efficient access to `.ptar` archives
- Supports most `plakar` commands
- No persistent state or extraction required

## Install

```sh
go install github.com/PlakarKorp/kapsul@v0.0.0-beta.10
```

## Usage

```sh
kapsul [-c <cores>] [-f <archive>] <subcommand> [...]
```

### Options

- `-c <cores>`: Limit number of CPU cores used
- `-f <archive>`: Path to `.ptar` archive

### Subcommands

The following `plakar` sub-commands are supported:

- `archive`
- `cat`
- `check`
- `create`
- `diff`
- `digest`
- `help`
- `info`
- `locate`
- `ls`
- `mount`
- `restore`
- `server`
- `ui`
- `version`

## Environment Variables

- `KAPSUL_PASSPHRASE`: Passphrase to unlock the encrypted archive

## Examples

Inspect the list of snapshots inside an archive:

```sh
kapsul -f backup.ptar ls
```

Create a new snapshot of the current directory:

```sh
kapsul -f backup-new.ptar create .
```

Restore the file `notes.md` from snapshot `abcd` inside the archive:

```sh
kapsul -f backup.ptar restore -to . abcd:notes.md
```

Launch the ui

```sh
kapsul -f backup.ptar ui
```

## See Also

- [`plakar(1)`](./plakar.1): The underlying backup engine
- [`ptar(5)`](./ptar.5): Archive format used by kapsul

---

© 2025 Plakar Korp. All rights reserved.
