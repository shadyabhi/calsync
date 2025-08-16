package version

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type UpdateInfo struct {
	Available     bool
	LatestVersion string
}

type Checker struct {
	currentVersion string
	githubRepo     string
}

func New(currentVersion string) *Checker {
	return &Checker{
		currentVersion: currentVersion,
		githubRepo:     "https://github.com/shadyabhi/calsync/releases/latest",
	}
}

func (c *Checker) CheckForUpdate(result chan<- *UpdateInfo) {
	if c.currentVersion == "dev" || strings.Contains(c.currentVersion, "dev") {
		result <- nil
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", c.githubRepo, nil)
	if err != nil {
		result <- nil
		return
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		result <- nil
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		result <- nil
		return
	}

	location := resp.Header.Get("Location")
	if location == "" {
		result <- nil
		return
	}

	parts := strings.Split(location, "/")
	if len(parts) == 0 {
		result <- nil
		return
	}

	latestVersion := parts[len(parts)-1]
	latestVersionClean := strings.TrimPrefix(latestVersion, "v")
	currentVersionClean := strings.TrimPrefix(c.currentVersion, "v")

	if c.isDifferent(latestVersionClean, currentVersionClean) {
		result <- &UpdateInfo{
			Available:     true,
			LatestVersion: latestVersion,
		}
	} else {
		result <- nil
	}
}

func (c *Checker) isDifferent(version1, version2 string) bool {
	v1Clean := c.cleanVersion(version1)
	v2Clean := c.cleanVersion(version2)

	return v1Clean != v2Clean
}

func (c *Checker) cleanVersion(version string) string {
	parts := strings.Split(version, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return version
}
