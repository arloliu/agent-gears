# memex

A memex is your repo's durable memory of how its own code works. Each entry records one mechanic: how control and data flow, which invariants hold, what breaks and how the breakage shows. Agents read the entry instead of re-deriving the mechanic from source, which saves tokens on every task that touches that part of the code.

Think of the bundle as a cache in front of your source tree. An entry is a cache line: filled on a miss, invalidated when the code it describes changes, never trusted while stale. Content digests, not dates, decide freshness. Every entry cites the exact files it was derived from, each with a sha256 digest, so a one-line `grep` pass tells you which entries drifted without spending a single model token.

## Why not just READMEs and doc comments?

Those document the public contract: what a function promises, how to call an API. A memex records what sits below that contract, the part you only learn by reading the implementation: why a timer gets suppressed, which goroutine owns a state transition, what symptom a violated invariant produces. The skill enforces this boundary with a novelty gate: before writing an entry, the agent greps existing docs for the same question and only records the delta. If a doc comment already answers it, no entry gets written.

## What's in this skill

| File | Role |
|---|---|
| `SKILL.md` | Entry point. Routes each invocation to a branch: sync, capture, verify, rebuild, check. |
| `BOOTSTRAP.md` | First-run setup: discovers units, interviews you, builds the bundle, installs the agent hook. |
| `FORMAT.md` | The exact file format memex writes, plus the conformance checks that keep entries falsifiable. |
| `OKF.md` | Vendored reference for the underlying OKF v0.2 format, so the skill works offline. |

The bundle itself lives in your repo (default `.knowledges/`), not here. This directory only holds the skill that maintains it.

## Getting started

Run `/memex` in a repo that has no bundle yet. The skill notices and switches to bootstrap, which walks you through a short interview, one question at a time:

1. Where the bundle lives (default `.knowledges/`)
2. Which directories count as units, based on what your repo already declares (Go packages, workspace globs, Cargo members)
3. What earns an entry in this repo, and what other docs already own
4. Which "hot" units get full entries now (everything else fills in on demand)
5. Whether the bundle is committed or git-ignored
6. Where the agent hook goes

Your answers land in `<bundle>/CONVENTIONS.md`, which every later run reads first and obeys. Bootstrap then writes a skeleton for every unit, seeds the hot units with entries, verifies those entries against the implementation, and installs the hook (see below).

## Everyday commands

| Invocation | What it does |
|---|---|
| `/memex` | Sync. Finds stale entries via digest mismatch, repairs them from the diff, proposes entries for uncovered code. |
| `/memex <unit>` | Sync scoped to one unit. |
| `/memex check` | Report freshness and coverage. Writes nothing. |
| `/memex capture` | Record one mechanic an agent just derived. Narrow by design: one entry, no sweep. |
| `/memex verify [unit]` | Re-check every claim against the implementation. Promotes correct entries to `stable`, demotes wrong ones. |
| `/memex rebuild [unit]` | Re-derive entries from scratch. For when drift has outrun repair, not routine refresh. |

## Keeping the bundle current

Digests do the watching for you. When a commit changes a file an entry cites, the entry's stored digest no longer matches, and the next sync flags exactly that entry. Nothing else gets touched, so a sync after a small change is cheap.

The working rhythm:

- **After landing a change** that touches documented mechanics, run `/memex` (or `/memex <unit>` for a focused pass). The skill reads the diff since the entry's recorded revision and decides per entry: cosmetic change, refresh the digest and leave the prose; material change, edit only the affected sections and demote the entry to `draft`; unclear, flag it for you.
- **Periodically**, or when a sync left drafts behind, run `/memex verify`. A material edit voids prior verification on purpose: whoever vouched for the entry vouched for text that no longer exists. Verify is what earns the text back its `stable` status.
- **Occasionally** run `/memex check` to see coverage and how long since the last sweep, with no writes.

Every entry carries a status you can trust at a glance: `stable` means read it as fact, `draft` means confirm claims against source before relying on them, `deprecated` means history only.

Every run that writes to the bundle ends by verifying its own output. Four scripted conformance passes (defined in `FORMAT.md`) check that each entry's frontmatter carries the fields the trust model reads (`type`, `title`, `description`, a valid `status`, a well-formed `generated`, at least one source), that source lines match the exact shape the freshness pass greps, that every file a pointer names is also cited, and that every pointed-at symbol still exists in the code. A violation blocks the run from finishing, because a malformed entry is the worst kind of defect: the freshness pass can't see it, so it reads as trustworthy forever.

## Making agents use and update it automatically

Bootstrap installs a short hook into your repo's agent instructions (a rules file indexed by `AGENTS.md`/`CLAUDE.md`, or `CLAUDE.md` directly). Because that file loads into every agent's context, every agent working in the repo follows the loop without being asked:

- **Before** reading source to work out how something behaves, it reads the unit's index in the bundle and the entries linked there.
- **During** the task, it trusts `stable` entries as written and treats `draft` entries as hints to confirm.
- **After** deriving a mechanic no entry covered, it runs `capture` to write the entry before finishing. After changing a mechanic an entry documents, it runs a sync for that unit.

This closes the loop: agents consume the bundle, and the same agents keep it filled and current as a side effect of normal work. Human-driven syncs after feature work catch whatever the loop misses.

If the hook is missing (say, someone trimmed `CLAUDE.md`), agents fall back to reading source and the bundle silently rots. `BOOTSTRAP.md` contains the canonical hook block; paste it back or re-run bootstrap's hook step.

## Trust model, briefly

An entry's frontmatter records who generated it, who verified it, and when. Verification by a different actor than the author counts for more than self-verification, and a `human:` verifier is the top tier. Digest freshness proves an entry matches the bytes it cites; verification proves the claims were derived correctly. You need both before an entry deserves blind trust, which is exactly what `stable` plus matching digests certifies.
