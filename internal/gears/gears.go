// Package gears embeds agent-gears' bundled gears (prebuilt binaries) and
// installs them into the user's ~/.local/bin.
package gears

import (
	"bufio"
	"compress/gzip"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:gears
var gearsFS embed.FS

// root is the subdirectory within gearsFS that holds one folder per prebuilt
// gear, each containing an index.json manifest and a single gzip-compressed
// binary.
const root = "gears"

// indexFile is the manifest filename expected in each gear's folder.
const indexFile = "index.json"

// Gear describes one prebuilt binary bundled with agent-gears.
type Gear struct {
	// Name is the installed binary's filename, e.g. "rtk".
	Name string `json:"name"`
	// Version is the gear's version, e.g. "0.44.2".
	Version string `json:"version"`
	// Description is a short human-readable summary of the gear.
	Description string `json:"description"`
	// gzPath is the path within gearsFS to the gzip-compressed binary.
	gzPath string
}

// list finds every gear bundled under root: one subdirectory per gear, each
// expected to contain an index.json manifest and a matching
// "<name>-v<version>.gz" compressed binary.
func list() ([]Gear, error) {
	dirs, err := fs.ReadDir(gearsFS, root)
	if err != nil {
		return nil, err
	}

	var gears []Gear
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		dirPath := root + "/" + d.Name()

		raw, err := fs.ReadFile(gearsFS, dirPath+"/"+indexFile)
		if err != nil {
			return nil, fmt.Errorf("read %s/%s: %w", dirPath, indexFile, err)
		}
		var g Gear
		if err := json.Unmarshal(raw, &g); err != nil {
			return nil, fmt.Errorf("parse %s/%s: %w", dirPath, indexFile, err)
		}

		g.gzPath = fmt.Sprintf("%s/%s-v%s.gz", dirPath, g.Name, g.Version)
		if _, err := fs.Stat(gearsFS, g.gzPath); err != nil {
			return nil, fmt.Errorf("gear %q: expected binary at %s: %w", g.Name, g.gzPath, err)
		}

		gears = append(gears, g)
	}
	return gears, nil
}

// Install detects prebuilt gears bundled with agent-gears, prompts the user
// (via reader/out) to install them into ~/.local/bin, and, if any of them
// already exist there, prompts separately before overwriting.
func Install(reader *bufio.Reader, out io.Writer) error {
	gears, err := list()
	if err != nil {
		return err
	}
	if len(gears) == 0 {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}
	binDir := filepath.Join(home, ".local", "bin")

	fmt.Fprintln(out, "Detected prebuilt gears:")
	for i, g := range gears {
		fmt.Fprintf(out, "  [%d] %s v%s - %s\n", i+1, g.Name, g.Version, g.Description)
	}

	fmt.Fprintf(out, "Install gears into %s? [y/N] ", binDir)
	line, _ := reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "y") {
		fmt.Fprintln(out, "Skipped gear installation.")
		return nil
	}

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", binDir, err)
	}

	toInstall := gears
	var existing []string
	for _, g := range gears {
		if _, err := os.Stat(filepath.Join(binDir, g.Name)); err == nil {
			existing = append(existing, g.Name)
		}
	}
	if len(existing) > 0 {
		fmt.Fprintf(out, "The following gears already exist in %s: %s\n", binDir, strings.Join(existing, ", "))
		fmt.Fprintf(out, "Overwrite them? [y/N] ")
		line, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "y") {
			skip := make(map[string]bool, len(existing))
			for _, n := range existing {
				skip[n] = true
			}
			var filtered []Gear
			for _, g := range gears {
				if !skip[g.Name] {
					filtered = append(filtered, g)
				}
			}
			toInstall = filtered
		}
	}

	if len(toInstall) == 0 {
		fmt.Fprintln(out, "Nothing to install.")
		return nil
	}

	var installed []Gear
	for _, g := range toInstall {
		target := filepath.Join(binDir, g.Name)
		if err := install(g, target); err != nil {
			fmt.Fprintf(out, "Failed to install %s: %v\n", g.Name, err)
			continue
		}
		installed = append(installed, g)
	}

	fmt.Fprintf(out, "Installed gears into %s:\n", binDir)
	for _, g := range installed {
		fmt.Fprintf(out, "  - %s v%s\n", g.Name, g.Version)
	}

	return nil
}

// install decompresses the gear's embedded gzip binary to target and marks
// it executable.
func install(g Gear, target string) error {
	src, err := gearsFS.Open(g.gzPath)
	if err != nil {
		return err
	}
	defer src.Close()

	gzr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gzr.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, gzr); err != nil {
		return err
	}
	return out.Chmod(0o755)
}
