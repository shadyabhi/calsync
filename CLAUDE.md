# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CalSync is a Go application that synchronizes calendar events between different
calendar sources (Mac Calendar, Google Calendar, ICS feeds). It reads events
from source calendars and syncs them to target calendars, helping users maintain
a unified view across different calendar systems.

## Development Instructions

When writing code, follow these guidelines

### Packages

- Package should follow the standard `New()` function (with arguments if
  needed), and then only expose methods that are absolutely necessary.
- Filenames in packages should be meaningful with almost always entrypoint being
  the <package_name>.go

### Comments:-

- Do not use comments unless absolutely necessary. Comments are for explaining
  the idea behind a feature if we need to, not explaining code. The code should
  be self-documenting.

## Development Commands

### Build
```bash
make build                    # Build the application
```

### Run
```bash
go run .                    # Run the application directly
DEBUG=1 go run .           # Run with debug logging enabled
make run                    # Build and run the generated binary
```

### Test
```bash
make test                   # Run tests via Makefile, use this at all times
go test ./calendar/maccalendar  # Run tests for a specific package
```

### Coverage
```bash
go test -v ./... -coverprofile=./cover.out  # Generate coverage profile
```

## Architecture

### Core Components

1. **Main Entry Point** (`main.go`):
   - Initializes structured logging with `log/slog`
   - Loads configuration from `~/.config/calsync/config.toml`
   - Uses reflection to dynamically initialize calendar clients based on config
   - Orchestrates the sync process between source and target calendars

2. **Calendar Interface** (`calendar/calendar.go`):
   - Defines the core `Calendar` interface with methods: `GetEvents()`, `PutEvents()`, `DeleteAll()`, `SyncToDest()`
   - `Event` struct represents calendar events with Title, Notes, Start/Stop times, and UID
   - Events can be hashed for duplicate detection using MD5

3. **Calendar Implementations**:
   - **Google Calendar** (`calendar/gcal/`): OAuth2-based integration with Google Calendar API
   - **Mac Calendar** (`calendar/maccalendar/`): Integration with macOS native calendar via icalBuddy
   - **ICS/iCal** (`calendar/ics/`): Support for ICS feed URLs

4. **Configuration** (`config/`):
   - TOML-based configuration
   - Supports multiple source and target calendars
   - Each calendar type can be independently enabled/disabled
   - Stores OAuth credentials and tokens for Google Calendar

### Key Design Patterns

- **Reflection-based Calendar Loading**: The application uses reflection in `getCalendarsFor()` to dynamically instantiate calendar clients based on configuration, providing flexibility for multiple calendar types.

- **Interface-based Architecture**: All calendar implementations conform to the `Calendar` interface, allowing polymorphic handling of different calendar sources.

- **Structured Logging**: Uses `log/slog` for structured logging with configurable log levels (INFO/DEBUG).

## Configuration Structure

The application expects configuration at `~/.config/calsync/config.toml`:

- **Source Calendars**: Mac, ICal (ICS feeds)
- **Target Calendars**: Google Calendar
- **Sync Settings**: Number of days to sync

## OAuth and Authentication

For Google Calendar:
- Credentials file required (specified in config)
- OAuth token stored locally and refreshed automatically
- Scopes: Full calendar access (`CalendarScope`)

## Logging Configuration

- Default level: INFO (shows main application flow)
- Debug mode: Set `DEBUG=1` environment variable
- Format: Clean key-value pairs without timestamps/levels for readability
