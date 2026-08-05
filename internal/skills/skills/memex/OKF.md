# OKF v0.2 — vendored reference

Everything memex needs to read or write a conformant bundle, vendored so the skill works with **no network access**. Nothing in this skill requires fetching a URL.

This is a working digest of the rules memex relies on, not a verbatim copy of the specification. Upstream normative text, if you ever need to settle a question this file does not answer: `GoogleCloudPlatform/knowledge-catalog`, path `okf/SPEC.md`. Treat that as provenance, not as a runtime dependency — and do not go fetch it merely to confirm something already stated below.

Where memex deliberately narrows OKF, [FORMAT.md](FORMAT.md) is the authority: it defines the exact subset written here. This file explains what the format *permits*, so a bundle from another producer is still readable.

## The model

A bundle is a directory tree of Markdown files with YAML frontmatter, distributed as a git repo, an archive, or a subdirectory of a larger repo. Two design commitments explain most of the rest: if you can `cat` a file you can read OKF, and consumers record *signals* rather than verdicts.

```
bundle/
  index.md          optional — directory listing
  log.md            optional — chronological history
  <concept>.md      concept documents
  <subdir>/
    index.md
    <concept>.md
```

`index.md` and `log.md` are reserved names. Every other `.md` file is a **concept document**: YAML frontmatter, then a free-form Markdown body.

## Frontmatter

**Required on every concept document:**

| Field | Meaning |
|---|---|
| `type` | Short string naming the kind of concept — "BigQuery Table", "Metric", "Playbook", "Mechanic". Not centrally registered; producers pick descriptive values. |

**Recommended:** `title` (display name), `description` (one sentence), `resource` (URI identifying the underlying asset), `tags` (list of strings).

**Provenance, trust, and lifecycle** — all optional families:

```yaml
sources:
  - id: unique-key
    resource: https://url-or-path
    title: Human-readable label
    author: actor-identifier
    usage_count: 12
    last_modified: 2026-08-05
usage_window: {from: 2026-01-01, to: 2026-08-05}

generated: {by: actor, at: 2026-08-05T12:00:00Z}
verified:
  - {by: actor, at: 2026-08-05T12:30:00Z}

status: draft | stable | deprecated
stale_after: 2026-11-05
```

- `generated.by` names who produced the current content; `generated.at` is when it last meaningfully changed.
- `verified` is a list of verification events. Trust tiers derive from it: no `verified` → **unverified**; non-human actors only → **machine-confirmed**; any `human:<id>` actor → **human-reviewed**.
- `status` defaults to `stable`. `draft` means not yet reviewed and possibly incomplete; `deprecated` means kept for links and history but no longer current.
- `stale_after` is an absolute date — content is stale on or after it. There is no relative TTL. *memex does not use this field: a content digest is a strictly better staleness signal than a calendar.*
- Per-claim attribution uses Markdown footnotes keyed to `sources[].id`, not positional indexes.

## Actor convention

Three identity forms for `generated.by` and `verified[].by`:

- `<producer>/<version>` for agents and tools — `reference_agent/gemini-2.5-pro`, `claude/opus-5`
- `human:<id>` for people — `human:arlo`
- `process:<id>` for automation — `process:finance-nightly`

Consumers classify trust on the `human:` prefix, so the form matters.

## Cross-linking

- **Bundle-relative links** start with `/` and resolve from the bundle root: `[customers](/tables/customers.md)`. Preferred, because they survive documents moving between subdirectories.
- Ordinary relative paths are also valid.
- Path-valued fields (`resource`, `sources[].resource`, and the computation fields) accept absolute URLs, bundle-relative paths, or relative paths.
- **Broken links are not malformed.** Consumers must tolerate a link whose target does not exist.
- A `references/` subdirectory conventionally mirrors external material, run instructions, or code. Convention, not requirement.

## Reserved files

`index.md` provides progressive disclosure at any directory level. It needs no frontmatter, with one exception: the bundle-root `index.md` may carry `okf_version`. Body groups entries under headings:

```markdown
# Group Heading

* [Title](relative-url) - description from the linked concept
* [Subdirectory](subdir/) - description
```

Consumers may synthesize an index on the fly when one is absent.

`log.md` records change history, newest first, with ISO 8601 date headings:

```markdown
# Update Log

## 2026-08-05
* **Update**: [Concept Title](path) what changed.
* **Creation**: [Concept Title](path) what was added.
```

The bold keywords (`**Update**`, `**Creation**`, `**Deprecation**`) are conventional, not required.

## Attested Computation

OKF defines a second concept type, `Attested Computation`, carrying both a meaning and a sanctioned computation method — `runtime`, typed `parameters`, an immutable `computation` (inline or by path), an `executor` returning a receipt, and a deterministic `attester` validating that receipt. An agent supplies parameter *values* only and can never author the computation.

**memex does not produce these.** It is recorded here so an agent reading someone else's bundle recognizes the type rather than treating it as malformed.

## Conformance

A bundle is conformant when every non-reserved `.md` file has parseable YAML frontmatter, every frontmatter block has a non-empty `type`, and reserved files follow their structure where present.

Consumption is **permissive by rule**. A consumer must not reject a bundle for:

- missing optional frontmatter
- unknown `type` values
- unknown additional keys
- broken cross-links
- missing `index.md`

A bare `verified` mapping must be treated as a one-element list. Failing attestations should be surfaced, never silently dropped.

This permissiveness is what lets memex add its own keys — `digest`, `revision`, and the `CONVENTIONS.md` settings — while staying conformant.

## Versioning

`<major>.<minor>`, currently **0.2**. Minor bumps add backward-compatible optional fields; major bumps change required fields or reserved filenames. A bundle may declare its target with `okf_version: "0.2"` in the root `index.md`.

Changes from v0.1, in case you meet an older bundle: `timestamp` became `generated.at`, and a `# Citations` body list became the `sources` frontmatter family. Consumers may still read the legacy forms.
