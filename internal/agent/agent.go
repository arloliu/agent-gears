// Package agent detects which AI coding agents are installed for the
// current user and where each one keeps its user-scoped skills folder.
package agent

import (
	"fmt"
	"os"
	"path/filepath"
)

// Target describes one AI coding agent and where its user-scoped skills
// folder lives.
type Target struct {
	// Name is a human-readable label for the agent, e.g. "Claude Code".
	Name string
	// SkillsDir is the absolute path to the agent's user-scoped skills folder.
	SkillsDir string
}

// candidate is an agent we know how to detect, keyed off a marker path
// (usually the agent's config directory) that must exist for the agent to
// be considered "installed".
type candidate struct {
	name      string
	marker    string // path checked for existence to decide the agent is installed
	skillsDir string // path skills should be installed into
}

// candidates returns the list of agents agent-gears knows how to detect,
// rooted at the given home directory.
func candidates(home string) []candidate {
	return []candidate{
		{
			name:      "Claude Code",
			marker:    filepath.Join(home, ".claude"),
			skillsDir: filepath.Join(home, ".claude", "skills"),
		},
		{
			name:      "OpenCode",
			marker:    filepath.Join(home, ".config", "opencode"),
			skillsDir: filepath.Join(home, ".config", "opencode", "skill"),
		},
		{
			name:      "Codex CLI",
			marker:    filepath.Join(home, ".codex"),
			skillsDir: filepath.Join(home, ".codex", "skills"),
		},
		{
			name:      "Generic agent skills (~/.agents/skills)",
			marker:    filepath.Join(home, ".agents"),
			skillsDir: filepath.Join(home, ".agents", "skills"),
		},
	}
}

// Detect scans the current user's home directory for known AI coding agents
// and returns one Target per agent found installed.
func Detect() ([]Target, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	var found []Target
	for _, c := range candidates(home) {
		if info, err := os.Stat(c.marker); err == nil && info.IsDir() {
			found = append(found, Target{Name: c.name, SkillsDir: c.skillsDir})
		}
	}
	return found, nil
}
