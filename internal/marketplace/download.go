package marketplace

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/breca/vsixdler/internal/config"
)

const downloadURLTemplate = "https://marketplace.visualstudio.com/_apis/public/gallery/publishers/%s/vsextensions/%s/%s/vspackage"

type DownloadTarget struct {
	Extension config.Extension
	Version   string
	Platform  string // empty for universal
}

func (d DownloadTarget) Filename() string {
	name := fmt.Sprintf("%s-%s", d.Extension.ID, d.Version)
	if d.Platform != "" {
		name += "@" + d.Platform
	}
	return name + ".vsix"
}

func (d DownloadTarget) URL() string {
	url := fmt.Sprintf(downloadURLTemplate, d.Extension.Publisher(), d.Extension.Name(), d.Version)
	if d.Platform != "" {
		url += "?targetPlatform=" + d.Platform
	}
	return url
}

func ResolveTargets(client *Client, extensions []config.Extension, verbose bool) ([]DownloadTarget, error) {
	var targets []DownloadTarget

	for _, ext := range extensions {
		pinned := ext.Version != ""

		result, err := client.QueryExtension(ext.ID, pinned)
		if err != nil {
			return nil, fmt.Errorf("resolving %s: %w", ext.ID, err)
		}

		if pinned {
			targets = append(targets, resolvePinned(ext, result, verbose)...)
		} else {
			targets = append(targets, resolveLatest(ext, result, verbose)...)
		}
	}

	return targets, nil
}

func resolvePinned(ext config.Extension, result *ExtensionResult, verbose bool) []DownloadTarget {
	var targets []DownloadTarget
	wantPlatforms := toSet(ext.Platforms)

	for _, ver := range result.Versions {
		if ver.Version != ext.Version {
			continue
		}

		if len(wantPlatforms) > 0 {
			if ver.TargetPlatform != "" && wantPlatforms[ver.TargetPlatform] {
				targets = append(targets, DownloadTarget{Extension: ext, Version: ver.Version, Platform: ver.TargetPlatform})
			}
		} else {
			if ver.TargetPlatform == "" {
				targets = append(targets, DownloadTarget{Extension: ext, Version: ver.Version})
			}
		}
	}

	// If user requested specific platforms but we found no platform-specific entries,
	// they may be universal — add the pinned version without platform.
	if len(wantPlatforms) > 0 && len(targets) == 0 {
		if verbose {
			log.Printf("%s@%s: no platform-specific builds found, downloading universal", ext.ID, ext.Version)
		}
		targets = append(targets, DownloadTarget{Extension: ext, Version: ext.Version})
	}

	if len(targets) == 0 {
		if verbose {
			log.Printf("warning: %s version %s not found in marketplace response", ext.ID, ext.Version)
		}
		// Still attempt the download — the direct URL may work even if the version
		// wasn't in the query response.
		targets = append(targets, DownloadTarget{Extension: ext, Version: ext.Version})
	}

	return targets
}

func resolveLatest(ext config.Extension, result *ExtensionResult, verbose bool) []DownloadTarget {
	var targets []DownloadTarget
	wantPlatforms := toSet(ext.Platforms)

	if len(wantPlatforms) == 0 {
		// Want universal / latest
		if len(result.Versions) > 0 {
			ver := result.Versions[0]
			targets = append(targets, DownloadTarget{Extension: ext, Version: ver.Version, Platform: ver.TargetPlatform})
		}
	} else {
		// Want specific platforms — find latest version for each
		seen := make(map[string]bool)
		for _, ver := range result.Versions {
			if ver.TargetPlatform != "" && wantPlatforms[ver.TargetPlatform] && !seen[ver.TargetPlatform] {
				seen[ver.TargetPlatform] = true
				targets = append(targets, DownloadTarget{Extension: ext, Version: ver.Version, Platform: ver.TargetPlatform})
			}
		}

		// Fall back to universal if no platform-specific builds found
		if len(targets) == 0 && len(result.Versions) > 0 {
			if verbose {
				log.Printf("%s: no platform-specific builds, downloading universal", ext.ID)
			}
			targets = append(targets, DownloadTarget{Extension: ext, Version: result.Versions[0].Version})
		}
	}

	return targets
}

func DownloadAll(ctx context.Context, targets []DownloadTarget, outDir string, concurrency int, verbose bool) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for _, t := range targets {
		g.Go(func() error {
			return Download(ctx, t, outDir, verbose)
		})
	}

	return g.Wait()
}

func Download(ctx context.Context, target DownloadTarget, outDir string, verbose bool) error {
	url := target.URL()
	filename := target.Filename()
	dest := filepath.Join(outDir, filename)

	if verbose {
		log.Printf("downloading %s -> %s", url, dest)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request for %s: %w", filename, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("downloading %s: HTTP %d: %s", filename, resp.StatusCode, truncate(string(body), 200))
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dest, err)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		os.Remove(dest)
		return fmt.Errorf("writing %s: %w", filename, err)
	}

	log.Printf("downloaded %s (%d bytes)", filename, written)
	return nil
}

func toSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[strings.TrimSpace(item)] = true
	}
	return m
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
