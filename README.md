[![Go](https://github.com/shadyabhi/calsync/actions/workflows/go.yml/badge.svg)](https://github.com/shadyabhi/calsync/actions/workflows/go.yml) ![coverage](https://raw.githubusercontent.com/shadyabhi/calsync/badges/.badges/main/coverage.svg)

# Calsync

Do you like to have one calendar that gives you a unified view of your day?
Are you annoyed that work calendar only sync with Mac's Calendar or Outlook?

This project is for you, it reads `Mac's Calendar events`, then syncs it to a Google Calendar of your choice.

WIP, not ready for public release yet, but it works.

# Installation

Ensure you've `brew` installed, if not, [follow their website for instructions.](https://brew.sh/)

```bash
# Installation
$ brew install shadyabhi/tap/calsync

# Upgrade
$ brew update && brew upgrade shadyabhi/tap/calsync
```

## Upgrade

```bash
$ brew update
$ brew upgrade calsync
```

# Run

## Config file

Location: `~/.config/calsync/config.toml`, checkout sample at:-

https://github.com/shadyabhi/calsync/blob/main/config/testdata/config.toml

## Run CLI

```
calsync
```

## Periodically as a cron

As Mac has permissions when reading Calendar data, it is not easy to run a cronjob or launchd daemon.
For now, the workaround is [documented here](https://github.com/shadyabhi/calsync/wiki/MacOS-Cronjob).

It will take a few clicks to get it working, but it works! 🎉

# Developer Notes

## Release

To release a new version:-

```bash
git tag -a v0.1.0 -m "My awesome release"
git push origin v0.1.0
```
