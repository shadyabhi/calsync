# Copilot Instructions for CalSync

## Overview
CalSync is a Go application that synchronizes calendar events between different calendar sources (Mac Calendar, Google Calendar, ICS feeds).

It uses structured logging with `log/slog` for better observability and debugging.

## Running the Application
To run the application, use:
```bash
go run .
```

For debug mode with verbose logging:
```bash
DEBUG=1 go run .
```

## Project Structure

- `main.go` - Main application entry point and orchestration
- `config/` - Configuration handling (TOML-based)
- `calendar/` - Calendar implementations
  - `gcal/` - Google Calendar integration
  - `ics/` - ICS feed support
  - `maccalendar/` - macOS Calendar integration

## Key Technologies

- **Go 1.24.1** - Main language
- **log/slog** - Structured logging (migrated from standard log package)
- **Google Calendar API** - For Google Calendar sync
- **TOML** - Configuration format
- **OAuth2** - Authentication for Google Calendar

## Logging

The application uses structured logging with `log/slog`:
- Default level: INFO - Shows main application flow
- Debug level: Set `DEBUG=1` environment variable for verbose output
- Format: Clean key-value pairs without timestamps/levels for readability
- Example: `msg="Running calsync" version=dev-none-unknown`

## Configuration

Configuration is stored in `~/.config/calsync/config.toml`. The app looks for:

- Source calendars (what to sync from)
- Target calendars (where to sync to)
- Sync settings (number of days, etc.)

## Development Notes

- Uses reflection for dynamic calendar type handling
- Implements duplicate detection for events
- Supports multiple calendar sources and targets
- OAuth token management for Google Calendar
- Event parsing and transformation between different calendar formats

## Testing

Run tests with:
```bash
go test ./...
```

## Build

Build the application:
```bash
go build .
```

## Common Operations

- Event synchronization between calendars
- Duplicate event detection and handling
- OAuth token refresh for Google Calendar
- Event filtering and transformation
- Time range queries for calendar events
