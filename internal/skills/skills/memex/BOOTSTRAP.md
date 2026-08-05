# Bootstrap

Runs once per repo. Produces `CONVENTIONS.md`, a skeleton over every unit, full entries for the hot units, and the hook that makes other agents use the bundle.

Read [FORMAT.md](FORMAT.md) before writing anything.

## Discover units first

Propose from what the repo already declares; never impose a layout on it.

| Signal | Units |
|---|---|
| `go.mod` | directories containing `*.go`, depth 1–2 |
| `package.json` workspaces, `pnpm-workspace.yaml` | the workspace globs |
| `Cargo.toml` `[workspace] members` | the members |
| `pyproject.toml`, `setup.cfg` | declared packages, else top-level dirs with `__init__.py` |
| none of the above | top-level directories holding source, minus vendor, build output, fixtures, and dot-directories |

Also read any existing map of the repo — README structure sections, an `AGENTS.md`, a docs index — and let it override the heuristics.

Each unit's bundle directory mirrors its repo path (`src/api/auth/` → `<bundle>/src/api/auth/`), so a glob matching nested directories produces nested unit directories, never flattened names.

## Interview

Ask **one question at a time**, each with your recommended answer first and the trade-off stated, and wait for the answer before asking the next. Later questions depend on earlier ones. This is a design conversation, not a form.

1. **Bundle directory** — default `.knowledges/`.
2. **Units** — present the discovered list, with what you excluded and why. Settle additions, removals, and exclusions.
3. **Scope** — mechanics only (default: flow, state, invariants, failure modes, where to look). Ask whether anything else belongs, and say plainly what already owns it: signatures belong to doc comments, decisions and rationale belong to ADRs or `CONTEXT.md` where the repo keeps them.
4. **Hot units** — which units to seed with full entries now. Everything else fills on miss via `capture`.
5. **Tracked or ignored** — committed travels with the repo, survives clones and worktrees, and shows knowledge changing alongside code; ignored keeps agent-authored docs out of the history.
6. **Hook location** — present what you detected (below) and confirm.

Record every answer in `CONVENTIONS.md`, including the repo-specific "what earns an entry" prose. An answer that is not written down will be re-invented by the next sync.

## Build

1. **`CONVENTIONS.md`** — the interview's outcome.
2. **Skeleton** — bundle `index.md` (`okf_version: "0.2"`, `swept_at: <HEAD>`), an empty `log.md`, and a `<unit>/index.md` for every unit: responsibility, boundary, entry points. Derive these from package docs, READMEs, and directory shape — this pass is structural and must stay cheap. Do not read units whole.
3. **Seed the hot units.** For each candidate entry, first pass the novelty gate: state the question it answers, grep the package's doc comments, README, and `docs/` for that question, and drop or narrow the candidate if they already answer it. A repo with thorough doc comments needs *fewer* entries, not more — the entries it does need sit below the documented contract.

   Write the **first entry**, show it, and ask for corrections to depth, tone, and length before writing any others. Apply the correction to everything that follows.

   Seeded entries are written `draft`. Before finishing, run the verify branch over them: check every claim against the implementation, correct what's wrong, and promote only what survives. Never hand back a bundle of `stable` entries nobody checked — the first run is exactly when a wrong claim is cheapest to catch and most expensive to leave.
4. **Hook** — write it, then show the user the exact text you added.
5. **Report** — the bundle you built, the units left as skeletons, and the two invocations they will actually use: `/memex` to sync and `/memex check` to inspect.

Done when every unit has an `index.md`, every hot unit has entries whose sources carry digests, the hook is installed, and the conformance check in FORMAT.md passes.

## Hook

Detect, in order:

1. A rules directory indexed by `AGENTS.md` (or `CLAUDE.md`) — add a rule file numbered to fit the existing sequence, plus one row in the index table.
2. Otherwise `CLAUDE.md` — append a short section.
3. Otherwise ask where repo-wide agent instructions belong. Do not create a convention the repo does not have.

The hook is exactly this block — a heading, one description line, three bullets, and one closing pointer line. Substitute the bundle directory chosen in question 1 for `<bundle>`, and the repo's own word for a unit (`package`, `crate`, `module`, `service`) for `<unit>`. Match the heading style of the repo's existing rule files if they have one; keep everything below the heading verbatim. Add no other prose — the hook is loaded into every agent's context, so every word is paid for on every turn.

After writing it, read the file back and diff it against this block. A hook that drifted from the template is the one piece of the install nothing else checks.

```markdown
# Memex

`<bundle>/` holds durable notes on how this repo works, one entry per mechanic.

- **Before** reading source to work out how something behaves, read `<bundle>/<unit>/index.md` and the entries it links.
- **Trust** `status: stable` entries as written. Treat `status: draft` — or a "where to look" symbol grep cannot find — as a hint to confirm against source.
- **After** deriving a mechanic no entry covered, run the memex skill's `capture` before you finish; after changing a mechanic an entry documents, run a sync for that unit.

Entries record mechanics *below* the public contract; see `<bundle>/CONVENTIONS.md` for what earns one.
```
