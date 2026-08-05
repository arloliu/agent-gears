---
name: memex
description: Maintain a repo's memex — an OKF knowledge bundle recording how the code actually works, so agents read the bundle instead of re-deriving mechanics from source. Use when the user wants a knowledge base built, synced, rebuilt, or checked for staleness; when a bundle entry looks stale; or when you derived a mechanic the bundle did not already hold and must record it.
---

# memex

A **memex** is a repo's durable knowledge of its own mechanics — how control and data flow, which invariants hold, where the state machines live, what breaks and how it shows. Each **entry** is one mechanic, written once and read many times, so an agent turns to the entry instead of re-deriving it from source.

An entry is a cache line, and the discipline is a cache's: fill on miss, invalidate on change, keep the hot set warm, never let a stale line be trusted.

## Locate the bundle

Every branch starts here, because the bundle's own config cannot tell you where the bundle is. Run from the repo root:

```bash
BUNDLE=$(ls -d .knowledges 2>/dev/null \
  || grep -rl --include=CONVENTIONS.md --exclude-dir=.git 'type: Bundle Conventions' . 2>/dev/null \
     | head -1 | xargs -r dirname)
```

An empty `$BUNDLE` means no bundle exists — bootstrap. Otherwise **read `$BUNDLE/CONVENTIONS.md` before anything else**: it is this repo's authority on units, exclusions, scope, and thresholds, and it overrides every default below.

Read [FORMAT.md](FORMAT.md) before writing any file in the bundle. It and [OKF.md](OKF.md) are the complete format reference — everything this skill needs is local, so never fetch a spec from the network.

## Routing

| Invocation | Branch |
|---|---|
| `$BUNDLE` empty | Bootstrap — follow [BOOTSTRAP.md](BOOTSTRAP.md) |
| `capture` | Capture — record one mechanic, no sweep |
| `verify [unit]` | Verify — re-derive each claim from source, promote or demote |
| `rebuild [unit]` | Rebuild — re-derive from source |
| `check` | Report freshness and coverage, write nothing |
| default, or `<unit>` | Sync |

## Freshness pass

Run from the repo root. It costs no model tokens, so run it in full before reading anything.

```bash
grep -rHo '{resource: [^,]*, digest: [^,}]*' "$BUNDLE" --include='*.md' \
| sed 's/:{resource: /|/; s/, digest: /|/' \
| while IFS='|' read -r doc path digest; do
    if [ ! -f "$path" ]; then echo "GONE  $doc  $path"
    else cur="sha256:$(sha256sum "$path" | cut -c1-16)"
         [ "$cur" = "$digest" ] || echo "STALE $doc  $path"
    fi
  done
```

Every entry it does not name is fresh. **Do not read fresh entries** — that is the whole saving.

## Sync

1. **Run the freshness pass.** The STALE and GONE lines are the complete candidate set; nothing outside it is touched this run.

2. **Judge each candidate from its diff, not from the file.** For a STALE path: `git diff <revision>..HEAD -- <path>`, plus `git diff -- <path>` for uncommitted work. Read the source file only when the diff is too large or too tangled to judge. Then place the change in a tier:

   - **Cosmetic** — tests, comments, formatting, or symbols this entry never cites. Refresh `digest` and `revision`. Touch no prose.
   - **Material** — a mechanic, invariant, or failure mode the entry describes has changed. Edit only the affected sections, refresh `digest` and `revision`, bump `generated.at`, **drop `verified`, and set `status: draft`** — whoever vouched for this entry vouched for text that no longer exists, so the edited text is unverified by definition.
   - **Unclear** — something changed that the diff does not explain. Leave the body alone, set `status: draft`, and report it under Needs attention.

   For a GONE path: grep for the symbols the entry cites. Found elsewhere → repoint `resource` and re-digest. Gone entirely → treat as material, or mark the entry `status: deprecated` if its subject no longer exists.

   Done when every STALE and GONE line has landed in **exactly one** of these five outcomes, with none left unaccounted for:

   1. `digest` and `revision` refreshed, prose untouched
   2. prose edited, `digest` and `revision` refreshed, `generated.at` bumped, `verified` dropped, `status: draft`
   3. `status: draft`, body untouched, reported under Needs attention
   4. `resource` repointed and re-digested
   5. `status: deprecated`

3. **Sweep for coverage.**
   - **New units** — directories matching `unit_globs` minus `exclude` that have no `<bundle>/<unit>/`. Author the unit `index.md` automatically; it is structural and cheap.
   - **Uncovered mechanics** — source files in a documented unit that no entry cites, above `propose_threshold_loc` and not in `declined`. **Propose** these with one line of rationale each; author only what the user approves, and append anything declined to `declined` in `CONVENTIONS.md` so it is not proposed again.
   - **Deleted units** — mark their entries `status: deprecated`. Never delete an entry; trails may still point at it.
   - Update `swept_at` in the bundle `index.md`.

4. **Record.** Append a dated line to `log.md` for every entry created, updated, or deprecated. Ensure each unit `index.md` links every entry it owns.

5. **Report** in the shape below.

## Capture

The write-back path: an agent derived a mechanic the bundle lacked and records it before finishing. Runs mid-task, so it stays narrow — one entry, no sweep, no report.

1. Identify the unit from `CONVENTIONS.md`, then read that unit's `index.md`.
2. **Pass the novelty gate and record the result.** State the question this entry answers, grep the unit's doc comments, README, and `docs/` for it, and open `# What it does` with the delta — naming the doc you checked and what it leaves out (or gets wrong). No delta means no entry.
3. Extend the existing entry if one covers this mechanic; otherwise create one.
4. Write it per FORMAT.md as `status: draft`, citing only files actually read, each with `digest` and `revision`.
5. Link it from the unit `index.md`; append a `**Creation**` or `**Update**` line to `log.md`.
6. Confirm in one line, naming it as draft until verified. Done when the entry exists, every cited file carries a digest, every pointer target is among those files, and the unit index links it.

## Verify

Digests prove an entry matches the bytes it was derived from. They cannot prove it was derived *correctly* — a confidently wrong entry stays byte-fresh forever, and a wrong entry is worse than no entry. Verify is the only branch that reads source specifically to contradict the bundle.

For each entry in scope, take its claims one at a time — every invariant, every failure mode, every pointer — and check each against the implementation. **Implementation wins over prose, including the repo's own doc comments**: where a doc comment and the code disagree, the code is the fact and the disagreement is itself worth recording in the entry.

Outcomes per entry:

- every claim holds → `status: stable`, append a `verified` record naming the verifying actor. **Prefer an actor other than the one in `generated.by`** — self-verification is the weakest evidence there is, and the two runs that produced wrong entries both passed it.
- any claim is wrong → correct it, `status: draft` until re-verified, and report it.
- a pointer names a symbol grep cannot find → repoint it or drop it.

Done when every entry in scope has been promoted, corrected, or demoted — none left at its prior status by default.

## Rebuild

Re-derive from source, replacing entry bodies. Recompute all digests, set `status: draft`, clear `verified`, keep `log.md` history intact, and append an `**Update**` line per entry. Rebuilt text is new text: it earns `stable` the same way any other new text does, by surviving a verify pass. Use rebuild when drift has outrun repair — not as a routine refresh.

`rebuild <unit>` covers that unit only and leaves `crosscutting/` alone, since a crosscutting entry answers to several units at once. `rebuild` with no argument covers every unit *and* `crosscutting/`.

## Check

Freshness pass plus coverage sweep, reported and nothing written — not even `swept_at`. Include how long ago the last sweep ran.

## Rules

- One mechanic per entry. It must answer **one question**, own **one boundary**, and have **one thing that would change it**. A title needing "and" fails all three at once.
- Before authoring, name the question the entry answers and grep the docs for it. If a doc comment, README, or guide already answers it, do not write the entry — narrow it to the part they leave out, or drop it. An entry that restates a doc inherits that doc's mistakes on top of costing tokens twice.
- Cite only files you read this run. A digest of a file you did not read is a lie, and the whole trust ladder rests on it.
- Every file a pointer names must also be a cited source, or nothing watches it.
- Point at symbols, never line numbers.
- New entries are born `draft`. Only a verify pass makes one `stable`.
- An entry earns its place when re-deriving it would cost more reading than the entry itself. Otherwise leave it out.
- Repair edits the sections the change touched and nothing else. Churn destroys the bundle's review value.

## Report

```
## memex — .knowledges

### Applied
- hsmsss/linktest-suppression.md: suppression window now keyed to T5 activity; digests refreshed.
- secs2/zero-copy-decode.md: digests only (test-file changes).

### Proposed
- hsmsss/send-pipeline.md — hsmsss/pipeline.go (612 LOC) cited by no entry.

### Needs attention
- secs1/block-transfer.md: state machine restructured, intent unclear — marked draft.

### Unchanged
- 14 entries fresh.
```
