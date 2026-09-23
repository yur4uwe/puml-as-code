# `UnhandledStatement` — Design & Implementation Methodology

Audience: implementing agent working directly in `pkg/parser` / `pkg/parser/ast` /
`pkg/parser/dialect`. This document explains *why* the shape below was chosen, not
just what to type, so implementation decisions made mid-work stay consistent with
the reasoning.

---

## 1. Problem Statement

PlantUML class diagrams contain valid surface syntax that `pac`'s parser does not
model semantically: `skinparam`, `legend`/`endlegend`, `header`/`footer`, `title`,
`caption`, `sprite`, `together`, `hide`/`show`, and `!`-prefixed preprocessor
directives (`!include`, `!if`, `!define`, etc.).

`pac fmt` must reprint these byte-for-byte without understanding them, while
preserving two invariants defined in the formatter roadmap:

```
Format(Format(src)) == Format(src)      // idempotence
Parse(src) == Parse(Format(src))        // AST stability
```

`UnhandledStatement` is the AST node that carries this content through the
pipeline unmodified.

---

## 2. Core Decision: Whitelist, Not Catch-All

**`UnhandledStatement` must only ever be produced from a known, enumerated set of
valid-but-unmodeled PlantUML constructs (a "tier-2 whitelist"). It must never be
produced as a fallback for "the parser failed to match this."**

This is the single most important rule in this document. Everything else follows
from it.

### Why this matters

Collapsing "valid syntax we don't implement" and "invalid/malformed syntax" into
one node type destroys the ability to do error reporting later. A typo inside a
`class` block and a legitimate `skinparam` line would become indistinguishable to
any downstream consumer — including the LSP diagnostics pass this feature is
ultimately a prerequisite for. Once that information is lost at parse time, it
cannot be reconstructed after the fact.

### The three-tier model

Statement parsing conceptually has three outcomes, not two:

1. **Matches real class-diagram grammar** → normal AST node (`ClassDecl`,
   `RelationshipDecl`, etc.). Unaffected by this work.
2. **Matches a tier-2 whitelist entry** → `UnhandledStatement`. Recognition is
   driven by an explicit table of leading keywords, described in §3.
3. **Matches neither** → a genuine parse error. **Out of scope for this task.**
   The parser continues to fail-fast (bubble the error up) for tier-3 input.
   Do not route tier-3 failures into `UnhandledStatement` as a stopgap — see §6.

---

## 3. Recognition: The Tier-2 Table

Recognition must happen as an explicit, positive match — i.e. "does this
statement's leading token/prefix appear in a known table" — not as the default
branch of a failed match. Implement this as a lookup table, consulted *before*
falling through to a parse-error return.

### 3.1 Important: several constructs once assumed to be tier-2 are already handled elsewhere

`together`, `hide`/`show`/`remove`, `skinparam`, and `title` are **already
modeled by real AST nodes** elsewhere in `pac` (e.g. `together` as an anonymous
package). They do **not** belong in the `UnhandledStatement` whitelist. Do not
add them; if any of them appear in the table during implementation, that's a
regression to remove, not a gap to fill.

The tier-2 whitelist is therefore limited to constructs `pac` genuinely does not
model at all yet:

| Keyword | Closer | Nestable | Dual-form (`BlockModifiers`) | Notes |
|---|---|---|---|---|
| `legend` | `endlegend` | no | yes — assumed `left/right/center/top/bottom`, **not yet confirmed** | inline-vs-block resolved per line via `IsBlock`; see §4.2 open item |
| `header` | `endheader` | no | yes — `left/right/center` | confirmed dual-form (inline text vs. multi-line block) via PlantUML web renderer |
| `footer` | `endfooter` | no | yes — `left/right/center` | confirmed dual-form via PlantUML web renderer |
| `caption` | — | — | — | always single-line |
| `sprite` | — | — | — | always single-line |
| `!function` | `!endfunction` | no | no | unconditionally block; macro body is preprocessor text, not diagram grammar |
| `!definelong` | `!enddefinelong` | no | no | same reasoning as `!function` |
| `!procedure` | `!endprocedure` | no | no | same reasoning as `!function` |
| `!if` | `!endif` | **yes** | no | unconditionally block; see §3.3 — deliberately opaque, confirmed nestable against real PlantUML |
| `!` (fallback) | — | — | — | any other `!`-prefixed line not matched above (`!include`, `!define`, etc.) |

This table is effectively a living changelog of "PlantUML features that exist and
are syntactically valid, but `pac` does not generate code for." Keep it
extensible — new entries should be addable without touching parsing logic beyond
the table itself.

Do **not** attempt to recognize other diagram types' syntax (sequence-diagram
`participant`, `alt`/`loop`/`end`, etc.) as tier-2. That content is not valid
class-diagram syntax and is explicitly tier-3 (a real error), per the decision in
§6.

Note: whether an entry is single-line, unconditionally block, or dual-form is
not a separately authored flag — it is fully determined by `Closer` and
`BlockModifiers`, per the `Tier2Entry`/`IsBlock` design in §4.1–§4.2. `legend`,
`header`, and `footer` must not be implemented as unconditionally block; their
form is resolved per occurrence.

### 3.2 Why grouping ("`Blocked` vs `SingleLine`") isn't a hand-set flag

Early drafts of this design assumed every tier-2 construct could default to
"one `UnhandledStatement` per source line," with grouping treated as a pure
formatting/tooling nicety to defer. That assumption was wrong for any construct
whose body is **not** statement-shaped text: if only the opener line
(`legend`, `!function foo()`, ...) is whitelisted and the parser is fail-fast
on tier-3, the very next line of a `legend` body (arbitrary free text) is
handed to the class-diagram grammar, fails to match anything, and aborts the
entire parse. This isn't a missing-nicety bug, it's a correctness bug — `fmt`
would be unable to process any file containing one of these constructs at all.

So whether an entry produces one line or a spanning block is not a
hand-authored enum — it falls out of two data fields, resolved by a single
`IsBlock` method (full definition in §4.1):

- **Block form** — the construct's body is not diagram grammar (free text, or
  semantics `pac` doesn't evaluate — see §3.3). Signaled by `Closer != ""`
  (and, for dual-form entries, by the line's trailing tokens all being
  recognized modifiers or absent — §4.2). On matching, the parser switches
  into raw-consume mode and keeps emitting subsequent lines into the *same*
  node's `Raw` field, **verbatim, without attempting to parse them as
  statements**, until it encounters the registered `Closer`. This is driven
  entirely by an explicit keyword pair per table entry — not by
  brace-matching, not by "keep going until something parses again."
- **Single-line form** — no body semantics; the whole construct is one line,
  one node. Signaled by `Closer == ""`, or by `Closer != ""` on a dual-form
  entry whose line carries real trailing content (§4.2).

This directly supersedes the "always exactly one line per node, grouping is a
deferrable nicety" position from an earlier draft of this document. That
position is only correct for entries whose `IsBlock` is always false.

### 3.3 Why `!if`/`!elseif`/`!else`/`!endif` is `Blocked`, not transparent

An earlier draft of this design deliberately kept `!if`/`!endif` transparent
(single-line nodes, with content between them left to parse as normal
class-diagram grammar), reasoning that swallowing the span opaquely would
discard legitimate `class`/`field` statements gated behind a platform check.
That reasoning has been superseded — it had the risk backwards.

The analogy that resolves it: this is the same situation as a C compiler that
does not run the C preprocessor. Given
`#ifdef WIN32 ... #else ... #endif`, a tool that can't evaluate the condition
has no principled basis for asserting *either* branch is "the real program" —
parsing both branches transparently doesn't preserve the source's meaning, it
*fabricates* a translation unit with conflicting declarations that no valid
evaluation of the source would ever produce (e.g. two conflicting definitions
of the same class). That is a worse failure than opaquely discarding the
span, not a better one.

`!if`/`!elseif`/`!else`/`!endif` is exactly this, not merely analogous to it:
without evaluating `pac`'s (currently unimplemented) preprocessor-condition
language, there is no basis for treating content in either branch as
unconditionally real. Nothing of provable validity is lost by treating the
whole conditional as opaque, because `pac` has no grounds to assert either
branch's content as ground truth in the first place. This puts `!if` in the
same category as `legend`/`header`/`footer`/`!function` — "a feature whose
semantics `pac` doesn't implement yet" — rather than as an exception to it.

**Mechanical consequences:**

- `!elseif` and `!else` are **not** separate `UnhandledStatement` nodes under
  this model. They are raw lines inside the swallowed body of the enclosing
  `!if`, exactly like any other line in that span, since the whole construct
  from `!if` to its matching `!endif` becomes one opaque unit.
- **Nesting is real** — confirmed against the PlantUML web renderer, e.g.
  `!if (A) ... !if (B) ... !endif ... !endif` renders successfully with both
  levels functioning. A flat "scan forward for the first line equal to
  `!endif`" would incorrectly close on the *inner* `!endif`, leaving the
  outer one dangling. `!if` is therefore the one whitelist entry with
  `Nestable = true`: the raw-consume loop must track a depth counter,
  incrementing on re-encountering the literal `!if` keyword and decrementing
  on `!endif`, closing the node only when depth returns to zero. All other
  `Blocked` entries (`legend`, `header`, `footer`, `!function`,
  `!definelong`, `!procedure`) are non-nestable — a simple scan to the first
  matching `Closer` is sufficient for them, per real-world PlantUML usage.

---

## 4. Node Shape

A single node type, not split into "statement" vs. "block" variants — how many
lines it swallows is decided at parse time by the registry (§4.1), not by the
node's type:

```go
type UnhandledStatement struct {
    // Raw source text, verbatim, including internal newlines when the
    // matched entry is Kind == Blocked and spans multiple lines.
    Raw string

    // Source span for diagnostics/tooling (LSP folding ranges, etc.)
    Span SourceSpan

    // Leading/trailing trivia (comments, blank-line counts), same
    // mechanism as every other Statement implementer.
    Trivia ast.Trivia
}
```

Rationale for a single type: a second `UnhandledBlock` type would have identical
responsibilities (hold raw text, reprint verbatim) differing only in "how many
lines it swallows," which is a parsing-time decision, not a type-level one. One
type with a `Raw string` that *may* contain `\n` covers both cases with strictly
less surface area for `fmt`, highlighting, and LSP to switch over later.

### 4.1 Tier-2 registry

The recognition table in §3.1 is not just documentation — it should exist as a
real, buildable registry the parser consults directly:

```go
type Tier2Kind int

type Tier2Entry struct {
    Keyword  string    // exact leading token, e.g. "legend", "!function"
    Closer   string    // required iff Kind == Blocked or DualForm == true
    Nestable bool       // true only for entries whose opener can recur inside its own body (currently: "!if")

    BlockModifiers []string // trailing tokens that do NOT count as inline content, e.g. {"left","right","center"}
}

var tier2Table = []Tier2Entry{
    {Keyword: "legend",      DualForm: true, Closer: "endlegend", BlockModifiers: []string{"left", "right", "center", "top", "bottom"}}, // BlockModifiers set: see §4.2 open item — verify against renderer before relying on this list
    {Keyword: "header",      DualForm: true, Closer: "endheader", BlockModifiers: []string{"left", "right", "center"}},
    {Keyword: "footer",      DualForm: true, Closer: "endfooter", BlockModifiers: []string{"left", "right", "center"}},
    {Keyword: "caption",     Kind: SingleLine},
    {Keyword: "sprite",      Kind: SingleLine},
    {Keyword: "!function",   Kind: Blocked, Closer: "!endfunction"},
    {Keyword: "!definelong", Kind: Blocked, Closer: "!enddefinelong"},
    {Keyword: "!procedure",  Kind: Blocked, Closer: "!endprocedure"},
    {Keyword: "!if",         Kind: Blocked, Closer: "!endif", Nestable: true},
    // any other line starting with "!" not matched above -> SingleLine fallback
}
```

**Lookup order:** exact-keyword match against the table first; if nothing
matches and the line starts with `!`, fall back to `SingleLine` as the generic
preprocessor-directive category (covers `!include`, `!define`, and anything
else not explicitly enumerated) rather than requiring every directive to be
listed individually.

**Consume loop for `Blocked` (and `DualForm`-resolved-to-block) entries:**
- Non-nestable (`legend`, `header`, `footer`, `!function`, `!definelong`,
  `!procedure`): append lines verbatim to `Raw` until a line's trimmed content
  equals `Closer`; include that closing line in `Raw` and close the node.
- Nestable (`!if` only, currently): same, but track a depth counter. Depth
  starts at 1 on the opening `!if`. Increment on any subsequent line whose
  trimmed content begins with `!if`; decrement on any line whose trimmed
  content equals `!endif`. Close the node only when depth reaches 0. `!elseif`
  and `!else` lines are appended to `Raw` like any other body line — they do
  not affect depth and are not separately recognized as tier-2 keywords under
  this model ($asee §3.3).

### 4.2 Dual-form keywords: block vs. inline, resolved per line

**This is a correction to the model as first drafted, not an addition.** The
original registry treated `Kind` as a fixed property of the keyword. That is
wrong for at least `header` and `footer` (and, pending confirmation — see the
open item below — likely `legend`): PlantUML accepts **two distinct forms**
for these keywords, distinguished only by what follows on the same line:

```plantuml
header My Header Text
```
— inline form: single line, content is the header text itself, **no
`endheader` follows and none should be searched for**.

```plantuml
header
  Multi-line
  header content
endheader
```
— block form: opener line has no real content (or only a positional
modifier), body is free text on subsequent lines, closed explicitly.

A fixed `Kind: Blocked` for `header` is unsafe: given the inline form, the
parser would match `header`, switch into raw-consume mode, and scan forward
for a literal `endheader` line that will never appear — silently swallowing
everything remaining in the file (subsequent real class declarations, other
tier-2 constructs, potentially past `@enduml`) into one `Raw` blob, with no
error raised anywhere. This is strictly worse than the problem `Blocked`
exists to solve: it doesn't lose information about one construct, it corrupts
the rest of the parse.

**Resolution rule, applied at the moment the keyword matches, for any entry
with `DualForm: true`:**

1. Take the remainder of the line after the keyword, tokenized on whitespace.
2. If the remainder is empty, **or every token in it appears in the entry's
   `BlockModifiers`** (pure positioning, e.g. `legend right`, `header
   center` — no actual content) → this is the **block form**. Proceed exactly
   as a `Blocked` entry: raw-consume until `Closer`, per §4.1.
3. Otherwise (remainder contains at least one token that is not a recognized
   modifier) → this is the **inline form**. Treat as `SingleLine`: one node,
   this line only, no closer expected, no forward scan performed.

`BlockModifiers` must be populated correctly per keyword for step 2 to be
reliable — an incomplete modifier list would cause a legitimate block-form
opener like `legend right` to be misclassified as inline (if `right` isn't in
the list) and produce a truncated one-line node instead of consuming the
body. **Open item, not yet confirmed:** the `header`/`footer` dual-form
behavior and their inline single-line usage were confirmed against the
PlantUML web renderer earlier in this design process. Whether `legend` has an
equivalent single-line inline form (as opposed to being block-only, always
requiring `endlegend`), and the exact `BlockModifiers` vocabulary for all
three keywords, was **not** separately confirmed and should be verified on
the renderer before this table is finalized. If `legend` turns out to be
block-only, set `DualForm: false, Kind: Blocked` for it instead — this is a
one-entry, no-logic-change correction if the assumption is wrong.

### 4.3 Bounding the raw-consume scan: unterminated blocks are tier-3, not silently accepted

Both §4.1's `Closer` scan and §4.2's block-form resolution assume a `Closer`
line will eventually appear. Nothing so far bounds what happens if it
doesn't — a `legend` with a missing `endlegend` due to a typo or truncated
file, for instance. Left unbounded, that scan would run to end-of-file and
either (a) succeed in swallowing the entire remainder of the diagram into one
`UnhandledStatement`, which is silent data loss with no diagnostic, or (b)
need special-casing that isn't specified anywhere yet. Neither is acceptable.

**The raw-consume scan for any `Blocked` (or block-form-resolved `DualForm`)
entry must be bounded by both of the following, whichever is reached first:**

- End of file, or
- A `}` that closes the enclosing `package`/diagram scope the `Blocked`
  statement opened inside (i.e. the raw-consume loop must still track brace
  depth *for the purpose of detecting the enclosing scope's end*, even though
  it does not attempt to parse the swallowed lines as statements).

**If either bound is reached without having matched the entry's `Closer`
(respecting `Nestable` depth where applicable), this is a tier-3 failure —
bubble up as a genuine parse error, per §2/§6, not a silently-closed
`UnhandledStatement`.** An unterminated block must be distinguishable from a
correctly-terminated one; treating both as valid input would mean a missing
`endlegend` and a well-formed diagram produce the same successful parse,
which defeats the purpose of keeping tier-3 fail-fast at all.

`UnhandledStatement` implements the `Statement` interface and is valid wherever
`Statement` is valid at `package` and top-level diagram scope. It is **not**
currently reachable inside `class`/`interface`/`enum` member lists — unrecognized
syntax in those scopes is handled separately by `LaxDialect` (`LaxField` /
`LaxMethod`), which is out of scope for this document.

---

## 5. Grouping Policy: Per-Entry, Not Uniform

**Superseded:** an earlier draft of this document specified a uniform default
of "exactly one `UnhandledStatement` per source line for everything, defer all
grouping." That position has been replaced by the `Kind`-driven model in §3.2
and §4.1: grouping is decided per tier-2 entry, not applied uniformly, because
for `Blocked` entries it is not a deferrable nicety — it is required for `fmt`
to be able to parse the file at all (§3.2 explains the failure mode this
avoids).

The rule as it now stands:

- **`SingleLine` entries** (`caption`, `sprite`, generic `!` fallback): exactly
  one `UnhandledStatement` per line, as originally specified. Nothing about
  these changed.
- **`Blocked` entries** (`legend`, `header`, `footer`, `!function`,
  `!definelong`, `!procedure`, `!if`): one `UnhandledStatement` spans from the
  opener line to the matching closer line (inclusive), with the entire span's
  text captured verbatim in `Raw`, newlines included. Matching is driven
  strictly by the registered `Closer` keyword (and depth counter, for the one
  `Nestable` entry) — never by "consume until something parses again." That
  heuristic was considered and explicitly rejected earlier in this design
  process specifically because it risks swallowing real class-diagram
  statements adjacent to unrelated unrecognized lines; keyword-pair matching
  against a fixed, known closer does not have that failure mode, because it
  only ever stops consuming on the one specific token the entry declares.

**AST-stability note:** because the `Closer`/`Nestable` matching rule is fully
deterministic given only the token stream (no dependence on formatting,
whitespace, or blank-line counts), a fresh parse and a round-tripped
(`Format`-then-`Parse`) parse necessarily produce identical grouping for
`Blocked` spans. `Parse(src) == Parse(Format(src))` continues to hold without
extra work, for the same reason it held under the original uniform policy —
determinism, not line-count, is what the invariant actually requires.

---

## 6. Explicitly Out of Scope (Do Not Implement Here)

To keep this change well-bounded, the following are deliberately deferred and
should **not** be built as part of `UnhandledStatement`:

- **Tier-3 error recovery / synchronizing parser.** Genuine syntax errors
  (anything class-diagram-shaped that fails to match, including other diagram
  types' syntax leaking in) continue to bubble up and fail-fast, exactly as the
  parser behaves today. Building panic-mode recovery (discard tokens to a sync
  point — end of line outside unclosed brackets, a closing `}`, or a
  statement-leading keyword — then resume) is a separate, larger workstream that
  is a hard prerequisite for the LSP track specifically, not for `fmt`. Do not
  bundle it into this task, and do not let tier-3 failures fall back to
  `UnhandledStatement` as a way to sidestep building it.
- **A distinct `ErrorStatement`/diagnostics-collecting node.** Needed when the
  recovery work above happens, not now.
- **Any actual evaluation of `!if`/`!elseif`/`!else` conditions**, `!function`
  bodies, `!definelong` macros, etc. These are captured as opaque raw text
  (§3.3, §4.1) and never interpreted — that's the entire point of tier-2.
- **Generic brace/bracket delimiter matching.** All `Blocked` entries in this
  design close on a specific keyword (`endlegend`, `!endif`, etc.), not on
  brace balancing — do not build a general-purpose bracket matcher, it isn't
  needed for any construct currently in the table.
- **Recognition of other diagram types' keywords** (`participant`, `alt`,
  `loop`, sequence-diagram arrows, etc.) as tier-2. These are invalid in a class
  diagram and belong to tier-3 once error recovery exists.

---

## 7. Acceptance Checklist

- [ ] `UnhandledStatement` is only constructed from a positive registry match
      (§3, §4.1), never from a parse-failure fallback.
- [ ] Single node type, `Raw string` (may contain `\n`), implements `Statement`.
- [ ] Valid at package/diagram scope only (not inside member lists).
- [ ] `together`, `hide`/`show`/`remove`, `skinparam`, `title` are **not**
      present in the tier-2 registry — confirm no regression reintroduces them
      (§3.1).
- [ ] `SingleLine` entries produce exactly one node per line, unchanged.
- [ ] `Blocked` entries (`!function`, `!definelong`, `!procedure`, `!if`)
      consume verbatim lines into a single node's `Raw` until their registered
      `Closer` is matched — never via "consume until something parses again."
- [ ] `legend`, `header`, `footer` are implemented as **`DualForm`**, not
      static `Blocked` — form is resolved per occurrence from the line's
      trailing content vs. `BlockModifiers` (§4.2), not hard-coded per
      keyword.
- [ ] `header`/`footer` inline single-line usage (e.g. `header My Text`)
      produces a single `SingleLine`-shaped node and does **not** trigger a
      forward scan for `endheader`/`endfooter`.
- [ ] `legend`'s dual-form-vs-block-only status and its `BlockModifiers`
      vocabulary (`left`/`right`/`center`/`top`/`bottom` assumed, not yet
      confirmed) are verified against the PlantUML renderer before or during
      implementation — open item from §4.2.
- [ ] The raw-consume scan for any `Blocked`/block-resolved entry is bounded
      by EOF or the enclosing scope's closing `}` (§4.3); reaching either
      bound without matching `Closer` (respecting `Nestable` depth) is treated
      as a tier-3 failure, not a silently-closed node.
- [ ] `!if` is the sole `Nestable = true` entry; the consume loop tracks depth
      (increment on nested `!if`, decrement on `!endif`, close at depth 0) —
      verified against a nested-`!if` fixture (see §3.3; nesting confirmed via
      PlantUML web renderer).
- [ ] `!elseif`/`!else` lines are captured as raw body content inside the
      enclosing `!if` node, not recognized as separate tier-2 keywords.
- [ ] Generic `!`-prefixed lines not matched by any specific registry entry
      fall back to `SingleLine` rather than causing a lookup failure.
- [ ] Tier-3 (genuine syntax errors, including any class-diagram-shaped content
      the registry doesn't match) still bubbles up/fail-fast, unchanged from
      current behavior — no new recovery logic introduced.
- [ ] `Format(Format(src)) == Format(src)` and `Parse(src) == Parse(Format(src))`
      hold for files containing every registry entry, including nested `!if`.
- [ ] Codegen path continues to discard `UnhandledStatement` nodes outright
      (no semantic handling expected there).
