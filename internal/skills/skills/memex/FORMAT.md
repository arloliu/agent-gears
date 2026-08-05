# Bundle format

The bundle is an OKF v0.2 bundle — a directory of Markdown files with YAML frontmatter. This file defines the subset memex writes; [OKF.md](OKF.md) is the vendored format reference, covering what OKF permits beyond that subset. Both are local: **this skill never needs network access.**

OKF consumers are permissive by rule — unknown keys are tolerated and missing optional keys are never an error — which is what lets the `digest` and `revision` keys below stay conformant.

## Layout

```
.knowledges/
  index.md            okf_version + swept_at, then the unit list
  log.md              dated history, newest first
  CONVENTIONS.md      this repo's settings (read first, never by consumers)
  <unit>/
    index.md          responsibility, boundary, entries, entry points
    <mechanic>.md     one entry per mechanic
  crosscutting/       mechanics spanning units, same shape
```

Reserved names are `index.md` and `log.md`. Every other `.md` file is an entry and **must** carry frontmatter with a non-empty `type`.

A unit's directory **mirrors its path in the repo**: `hsmsss/` → `<bundle>/hsmsss/`, `src/api/auth/` → `<bundle>/src/api/auth/`. Never flatten a nested path into a single name — mirroring keeps unit directories collision-free and makes a trail readable as a repo path. `unit` in `CONVENTIONS.md` records the repo-relative path.

## Entry

```markdown
---
type: Mechanic
title: Linktest suppression
description: Why and when HSMS-SS suppresses the linktest timer.
tags: [hsmsss, timers, session]
status: stable
generated: {by: "claude/opus-5", at: 2026-08-05T14:20:00Z}
verified:
  - {by: "human:arlo", at: 2026-08-05T15:02:00Z}
sources:
  - {resource: hsmsss/linktest.go, digest: sha256:f02fe29be77d3794, revision: bc97919}
  - {resource: hsmsss/connection.go, digest: sha256:41c9a0b2ee5d7710, revision: bc97919}
---

# What it does
One to three sentences. The thing a reader needs before anything else — stated as what the
public docs do *not* already say. Where a doc comment covers the same ground, say so and
record the delta, including any place the doc is imprecise.

# How it works
The mechanic itself — sequence, state, ownership, who calls whom.

# Invariants
- Conditions that must hold, phrased so a violation is recognisable in a diff.

# Failure modes
- What breaks, and the symptom it produces.

# Where to look
- suppression gate: `hsmsss/linktest.go` → `(*linktestTimer).shouldSuppress`
- reset on activity: `hsmsss/connection.go` → `(*Connection).noteActivity`
```

`# Gotchas` is optional. Everything else stays in this order — a fixed shape is what makes targeted repair possible.

## Fields

| Field | Meaning |
|---|---|
| `type` | Required. `Mechanic` for entries, `Unit` for unit indexes, `Bundle Conventions` for CONVENTIONS.md. |
| `title`, `description`, `tags` | Display name, one sentence, and search terms. |
| `status` | `stable`, `draft` (do not trust the body), or `deprecated` (history only). Always write it explicitly; a missing `status` is *read* as `stable`. **A newly authored entry is born `draft`** and becomes `stable` only after a verify pass — digests prove an entry is *current*, never that it is *correct*. |
| `generated` | `{by: <actor>, at: <ISO-8601>}` — who wrote the current text and when. |
| `verified` | List of `{by: <actor>, at: <ISO-8601>}` — one record per verify pass that found no fault. A `human:<id>` actor is the top tier; a model actor is real but weaker evidence, and weakest of all when it is the same actor as `generated.by`. Cleared on any material edit. |
| `sources` | Every file the entry was derived from. See below. |

Actors follow OKF: `<producer>/<version>` for models, `human:<id>` for people, `process:<id>` for automation.

`stale_after` is deliberately unused — digests are a strictly better staleness signal than a calendar.

## sources

```yaml
sources:
  - {resource: hsmsss/linktest.go, digest: sha256:f02fe29be77d3794, revision: bc97919}
```

- **Flow style, one line per source, exactly these three keys in this order.** This is not cosmetic: the freshness pass greps this exact shape, so a reordered or multi-line mapping is *invisible* to it and the entry reads as permanently fresh. A malformed source line is worse than a missing one — it fails silently, in the direction of false trust. The conformance check below is what catches it.
- `resource` — path relative to the repo root.
- `digest` — `sha256:` plus the **first 16 characters** of `sha256sum <file>`. This is the authority on whether the entry is stale, and it is unaffected by commits, amends, rebases, and squashes.
- `revision` — HEAD at the moment the entry was derived, used only as a diff base. Record HEAD even when the working tree is dirty and the derivation read uncommitted content — the digest already covers what was read, and HEAD is the base a later diff needs. Use `uncommitted` only when there is no HEAD to record: a repo with no commits, or no git at all.

## Trust ladder

| Signal | What a consumer does |
|---|---|
| `status: stable`, digests match | Trust the entry. Do not re-read the source. |
| `status: draft`, or a digest mismatch | Hint only. Confirm each claim against source before relying on it. |
| `verified` by an actor other than `generated.by` | An independent pass found no fault. |
| `verified` with a `human:` actor | Strongest tier — a person read this text against the code. |
| `status: deprecated` | History. Do not use it to reason about current behaviour. |

## Links and pointers

- **Trails** between entries are bundle-relative and start with `/`: `[linktest suppression](/hsmsss/linktest-suppression.md)`. They survive files moving between subdirectories.
- **Code pointers** are `file` → `qualified symbol`, never a line number. A symbol that grep cannot find is a signal the entry is stale — a wrong line number is silent.
- Broken trails are tolerated, not errors. Never delete an entry to fix one.

## index.md

Bundle root, the only index carrying frontmatter:

```markdown
---
okf_version: "0.2"
swept_at: bc97919
---

# Units

* [hsmsss](hsmsss/) - HSMS-SS session layer: connection lifecycle, linktest, reconnect.
* [secs2](secs2/) - Item model and wire encoding.
```

Unit index:

```markdown
---
type: Unit
title: hsmsss
description: HSMS-SS session layer — connection lifecycle, linktest, reconnect.
---

# Responsibility
What this unit owns, in one or two sentences.

# Boundary
What it does *not* own, and which unit does.

# Entries
* [Linktest suppression](/hsmsss/linktest-suppression.md) - why and when the timer is suppressed.

# Entry points
- session setup: `hsmsss/session.go` → `(*Session).Open`
```

## log.md

Newest first, ISO dates, one line per change:

```markdown
# Log

## 2026-08-05
* **Update**: [Linktest suppression](/hsmsss/linktest-suppression.md) suppression window keyed to T5 activity.
* **Creation**: [Send pipeline](/hsmsss/send-pipeline.md) batching path introduced in pipeline.go.
* **Deprecation**: [Legacy framing](/secs1/legacy-framing.md) source removed.
```

## CONVENTIONS.md

```markdown
---
type: Bundle Conventions
title: <repo> memex conventions
unit: go-package
unit_globs: ["*/"]
exclude: [tmp, .worktrees, testdata, vendor]
scope: mechanics
pointer_style: symbol
tracked: true
hook: .agents/rules/150-memex.md
propose_threshold_loc: 150
declined: []
---

# What earns an entry
<the repo-specific answer settled during bootstrap>

# What does not
<what other files already own — READMEs, doc comments, ADRs>
```

## Conformance check

Run before finishing any run that wrote to the bundle. It prints every source line that does not match the canonical shape — the lines the freshness pass would silently skip:

```bash
grep -rn 'resource:' --include='*.md' "$BUNDLE" \
| grep -v '{resource: [^,]*, digest: sha256:[0-9a-f]\{16\}, revision: [^,}]*}'
```

Silence means clean. Any line it prints must be repaired before the run ends, because that entry is currently unfalsifiable.

Then check that every file an entry points at is also a file it cites — an uncited pointer target is unwatched, so a claim resting on it rots while the entry still reads as fresh:

```bash
find "$BUNDLE" -name '*.md' ! -name index.md ! -name log.md | while read -r f; do
  srcs=$(grep -o '{resource: [^,]*' "$f" | sed 's/{resource: //')
  sed -n '/^# Where to look/,$p' "$f" | grep -o '`[^`]*\.[a-z]*`' | tr -d '`' | grep '/' | sort -u \
  | while read -r p; do
      printf '%s' "$srcs" | grep -qx "$p" || echo "UNCITED  $f  ->  $p"
    done
done
```

Then check that every symbol a pointer names still exists in the files that entry cites. This is the promise symbol pointers make — that a rename breaks them *loudly* — and nothing else enforces it:

```bash
find "$BUNDLE" -name '*.md' ! -name index.md ! -name log.md | while read -r f; do
  files=$(grep -o '{resource: [^,]*' "$f" | sed 's/{resource: //' | tr '\n' ' ')
  [ -z "$files" ] && continue
  sed -n '/^# Where to look/,$p' "$f" | grep -o '`[^`]*`' | tr -d '`' | grep -v '/' | sort -u \
  | while read -r sym; do
      name="${sym##*.}"; name="${name%%(*}"; [ -z "$name" ] && continue
      grep -qw -- "$name" $files 2>/dev/null || echo "NOSYM  $f  ->  $sym"
    done
done
```

A `NOSYM` line means the symbol was renamed or removed: repoint it, or the entry is describing code that no longer exists under that name.

Then confirm by reading: every non-reserved `.md` parses as YAML frontmatter plus body and carries a non-empty `type`; every unit index lists the entries it owns; every unit directory mirrors its repo path.
