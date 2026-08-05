// Command agent-gears detects installed AI coding agents on the current
// machine and offers to install this module's bundled skills and gears
// (prebuilt binaries) for them.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/arloliu/agent-gears/internal/gears"
	"github.com/arloliu/agent-gears/internal/skills"
)

// version is the agent-gears tool version.
const version = "0.1.0"

const rootUsage = `agent-gears detects installed AI coding agents and lets you install
this module's bundled skills and gears (prebuilt binaries) for them.

Usage:
  agent-gears [command]

Commands:
  install    Install both skills and gears (default when no command is given)
  skills     Install skills into detected agent skill folders
  gears      Install gears (prebuilt binaries) into ~/.local/bin
  version    Print the agent-gears version
  help       Show this help message

Flags:
  -h, --help  Show help for agent-gears or a specific command

Run 'agent-gears <command> -h' for command-specific help.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		return runInstall(nil)
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "install":
		return runInstall(rest)
	case "skills":
		return runSkills(rest)
	case "gears":
		return runGears(rest)
	case "version":
		fmt.Println("agent-gears", version)
		return 0
	case "help", "-h", "--help":
		fmt.Print(rootUsage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "agent-gears: unknown command %q\n\n", cmd)
		fmt.Fprint(os.Stderr, rootUsage)
		return 1
	}
}

func runInstall(args []string) int {
	fs := newFlagSet("install", `Usage: agent-gears install

Detect installed AI coding agents and install this module's bundled
skills into each one, then detect bundled gears (prebuilt binaries)
and install them into ~/.local/bin. Prompts before each install and
before overwriting anything already present.
`)
	if code, done := fs.parse(args); done {
		return code
	}

	reader := bufio.NewReader(os.Stdin)
	if err := skills.Install(reader, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "agent-gears:", err)
		return 1
	}
	if err := gears.Install(reader, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "agent-gears:", err)
		return 1
	}
	return 0
}

func runSkills(args []string) int {
	fs := newFlagSet("skills", `Usage: agent-gears skills

Detect installed AI coding agents and install this module's bundled
skills into each detected agent's skills folder. Prompts before each
install and before overwriting any skill that's already present.
`)
	if code, done := fs.parse(args); done {
		return code
	}

	reader := bufio.NewReader(os.Stdin)
	if err := skills.Install(reader, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "agent-gears:", err)
		return 1
	}
	return 0
}

func runGears(args []string) int {
	fs := newFlagSet("gears", `Usage: agent-gears gears

Detect gears (prebuilt binaries) bundled with agent-gears and install
them into ~/.local/bin. Prompts before installing and before
overwriting any gear that's already present.
`)
	if code, done := fs.parse(args); done {
		return code
	}

	reader := bufio.NewReader(os.Stdin)
	if err := gears.Install(reader, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "agent-gears:", err)
		return 1
	}
	return 0
}

// subFlagSet wraps flag.FlagSet with a fixed usage message and consistent
// -h/--help handling for agent-gears subcommands.
type subFlagSet struct {
	*flag.FlagSet
}

func newFlagSet(name, usage string) *subFlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	return &subFlagSet{fs}
}

// parse parses args and reports whether the caller should stop, along with
// the exit code to use in that case (0 for -h/--help, 2 for a parse error).
func (fs *subFlagSet) parse(args []string) (code int, done bool) {
	err := fs.Parse(args)
	if err == nil {
		return 0, false
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0, true
	}
	return 2, true
}
