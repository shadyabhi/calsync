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

To get `credentials.json`, [follow this guide](https://github.com/shadyabhi/calsync/wiki/Google-Calendar-authorization). 

```toml
➤ cat ~/.config/calsync/config.toml
[Secrets]
# credentails is downloaded from Google Cloud console,
# Written Instructions: https://github.com/shadyabhi/calsync/wiki/Google-Calendar-authorization
# Video: https://youtu.be/c2b2yUNWFzI?t=227
Credentials = "credentials.json"
# token.json contains the token that's refreshed frequently, auto-generated and managed by code.
Token = "token.json"

[Mac]
ICalBuddyBinary = "/usr/local/bin/icalBuddy"
# Name of source Calendar in Calendar app
Name = "Calendar"
# Days to sync in future
Days = 7

[Google]
# Calendar name to sync on personal account
# if this Id is invalid, program will list all possible options available on your Google account.
Id = "abcd@group.calendar.google.com"
```

## Run CLI

```
calsync
```
Run from a folder where `credentials.json` file exists.

## Periodically as a cron

As Mac has permissions when reading Calendar data, it is not easy to run a cronjob or launchd daemon. 
For now, the workaround is [documented here](https://github.com/shadyabhi/calsync/wiki/MacOS-Cronjob). 

It will take a few clicks to get it working, but it works! 🎉

