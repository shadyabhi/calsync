package cmd

import (
	"calsync/config"
	versioncheck "calsync/version"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func run(cmdArgs cmdArgs) {
	setupLogging()

	slog.Info("Running calsync", "version", fmt.Sprintf("%s-%s-%s", Version, Commit, Date))

	ctx := context.Background()

	cfg, err := config.GetConfig(os.Getenv("HOME") + "/.config/calsync/config.toml")
	if err != nil {
		slog.Error("Failed to get config", "error", err)
		os.Exit(1)
	}

	// Handle delete-dst flag if provided
	if cmdArgs.deleteDst != "" {
		if err := handleDeleteDestination(ctx, cfg, cmdArgs.deleteDst); err != nil {
			slog.Error("Failed to delete events from destination", "error", err, "calendar", cmdArgs.deleteDst)
			os.Exit(1)
		}
		slog.Info("Successfully deleted all calsync-managed events", "calendar", cmdArgs.deleteDst)
		return
	}

	versionChan := make(chan *versioncheck.UpdateInfo, 1)
	versionChecker := versioncheck.New(Version)
	go versionChecker.CheckForUpdate(versionChan)

	syncCalendars(ctx, cfg)

	// Check for update info at the very end
	select {
	case update := <-versionChan:
		if update != nil && update.Available {
			slog.Warn("A newer version is available",
				"current", Version,
				"latest", update.LatestVersion,
				"action", "Run 'brew update && brew upgrade shadyabhi/tap/calsync' to update")
		}
	default:
		slog.Debug("No update info received, this version is up-to-date")
	}
}

func setupLogging() {
	level := slog.LevelInfo
	if os.Getenv("DEBUG") != "" {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	slog.SetDefault(logger)
}

func handleDeleteDestination(ctx context.Context, cfg *config.Config, calendarName string) error {
	cal, err := getCalendarByName(ctx, cfg, strings.ToLower(calendarName))
	if err != nil {
		return fmt.Errorf("failed to get calendar %s: %w", calendarName, err)
	}

	slog.Info("Deleting all calsync-managed events",
		"calendar", calendarName,
		"nDays", cfg.Sync.Days)

	if err := cal.DeleteAll(cfg.Sync.Days); err != nil {
		return fmt.Errorf("failed to delete events: %w", err)
	}

	return nil
}
