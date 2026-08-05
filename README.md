# agent-gears

A small Go CLI that detects which AI coding agents are installed for the
current user (Claude Code, OpenCode, Codex CLI, or a generic
`~/.agents/skills` setup) and installs bundled **skills** and **gears**
(prebuilt binaries) for them.

## Install

```sh
go install github.com/arloliu/agent-gears/cmd/agent-gears@latest
```

This installs a single self-contained `agent-gears` binary — the bundled
skills and gears are embedded in it at build time, so no other files need to
be downloaded separately.

## Usage

```sh
agent-gears [command]
```

| Command   | Description                                                          |
|-----------|-----------------------------------------------------------------------|
| `install` | Install both skills and gears (default when no command is given)      |
| `skills`  | Install skills into detected agent skill folders                      |
| `gears`   | Install gears (prebuilt binaries) into `~/.local/bin`                 |
| `version` | Print the agent-gears version                                         |
| `help`    | Show the help message                                                 |

Run `agent-gears <command> -h` for command-specific help. Every install step
prompts before writing anything, and prompts again before overwriting a
skill or gear that's already present at the destination.

## Project layout

```
cmd/agent-gears/     CLI entry point (flag parsing, subcommands)
internal/agent/      detects installed AI coding agents
internal/skills/      embeds and installs bundled skills
internal/gears/       embeds and installs bundled gears (prebuilt binaries)
```

Adding a new skill means dropping a folder under `internal/skills/skills/`.
Adding a new gear means adding a folder under `internal/gears/gears/<name>/`
containing an `index.json` manifest (`name`, `version`, `description`) and a
`<name>-v<version>.gz` gzip-compressed binary.

## Development

```sh
make build   # build ./bin/agent-gears
make check   # gofmt check + go vet + go test
make run     # go run the CLI (pass args via ARGS="...")
make help    # list all Makefile targets
```

## License

[MIT](LICENSE)
