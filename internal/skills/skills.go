// Package skills embeds agent-gears' bundled skills and installs them into
// detected AI coding agents' skills folders.
package skills

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/arloliu/agent-gears/internal/agent"
)

//go:embed all:skills
var skillsFS embed.FS

// root is the subdirectory within skillsFS that holds the actual skill
// folders (embed.FS keeps the "skills/" prefix from the directive).
const root = "skills"

// Install detects installed agents, prompts the user (via reader/out) to
// choose which ones to install skills into, and copies the embedded skills
// into each chosen agent's skills folder.
func Install(reader *bufio.Reader, out io.Writer) error {
	agents, err := agent.Detect()
	if err != nil {
		return err
	}

	if len(agents) == 0 {
		fmt.Fprintln(out, "No supported agent installations were detected.")
		return nil
	}

	fmt.Fprintln(out, "Detected agent skill folders:")
	for i, a := range agents {
		fmt.Fprintf(out, "  [%d] %s -> %s\n", i+1, a.Name, a.SkillsDir)
	}

	names, err := names()
	if err != nil {
		return err
	}

	for _, a := range agents {
		fmt.Fprintf(out, "Install skills into %s (%s)? [y/N] ", a.Name, a.SkillsDir)
		line, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "y") {
			fmt.Fprintf(out, "Skipped %s.\n", a.Name)
			continue
		}

		toInstall := names
		existing := existing(a.SkillsDir, names)
		if len(existing) > 0 {
			fmt.Fprintf(out, "The following skills already exist in %s: %s\n", a.SkillsDir, strings.Join(existing, ", "))
			fmt.Fprintf(out, "Overwrite them? [y/N] ")
			line, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "y") {
				toInstall = subtract(names, existing)
			}
		}

		if len(toInstall) == 0 {
			fmt.Fprintf(out, "Nothing to install into %s.\n", a.Name)
			continue
		}

		if err := copyTo(a.SkillsDir, toInstall); err != nil {
			fmt.Fprintf(out, "Failed to install into %s: %v\n", a.Name, err)
			continue
		}
		fmt.Fprintf(out, "Installed skills into %s:\n", a.SkillsDir)
		for _, n := range toInstall {
			fmt.Fprintf(out, "  - %s\n", n)
		}
	}

	return nil
}

// names lists the top-level skill folder names bundled in skillsFS.
func names() ([]string, error) {
	entries, err := fs.ReadDir(skillsFS, root)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// existing returns the subset of names that already exist as entries under
// dst.
func existing(dst string, names []string) []string {
	var found []string
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(dst, n)); err == nil {
			found = append(found, n)
		}
	}
	return found
}

// subtract returns the elements of all that are not present in remove.
func subtract(all, remove []string) []string {
	skip := make(map[string]bool, len(remove))
	for _, n := range remove {
		skip[n] = true
	}
	var result []string
	for _, n := range all {
		if !skip[n] {
			result = append(result, n)
		}
	}
	return result
}

// copyTo recursively copies the named embedded skill folders into dst,
// creating directories as needed.
func copyTo(dst string, names []string) error {
	for _, name := range names {
		src := root + "/" + name
		err := fs.WalkDir(skillsFS, src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			target := filepath.Join(dst, rel)

			if d.IsDir() {
				return os.MkdirAll(target, 0o755)
			}
			return copyFile(path, target)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := skillsFS.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
