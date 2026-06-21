# Rebuild the PoC as a modern gnark library (zkkit)

**Date**: 2026-06-21
**Context**: User asked to turn `gnark-rollup-exp` into a "production-grade
usable Go zk library where you write circuits in Go, with examples." The repo is
a 2020-era PoC on `gnark v0.2.1-alpha`, whose API (`frontend.CS`,
`*frontend.Constraint`, `MUSTBE_EQ`, `gurvy`) no longer exists.

**Options considered**:
- A. Reimplement a zk-SNARK library from scratch in Go. — Enormous; would be
  strictly inferior to `gnark`; reinvents field math, compiler, backends.
- B. Modernize and re-architect the PoC into a tested, importable Go library
  *on top of* modern `gnark` (v0.15): prove/verify harness + rollup reference +
  example circuits. — Achievable, genuinely useful, honest positioning.
- C. Minimal port: just make the existing files compile on new gnark. — Leaves a
  messy multi-module PoC, no tests, not a "library".

**Decision**: B.

**Why**: `gnark` already *is* "a production Go library for writing circuits in
Go." Competing with it (A) wastes effort for a worse result. The real unmet need
is the application layer — reusable gadgets, a clean harness, a working modern
rollup, and copyable tested examples — which is exactly what a B-Harvest team
prototyping zk apps would reach for. This matches the user's intent ("write
circuits in Go, a zk library, examples") while being truthful about what is worth
building. Constitution III (minimize deps) and the CHARTER out-of-scope (no
reimplementing gnark primitives) follow from this.

**Backport**: CHARTER.md (whole), specs/001-zkkit/{spec,plan,tasks}.md, README.md
positioning section.

---

## Sub-decision: working around a no-unlink filesystem

**Context**: The mounted workspace allows file create and rename but denies
`unlink`/delete (EPERM on `rm`, even for owned files). A crashed `git rm` left a
stale `.git/index.lock` that could not be deleted, wedging git.

**Decision**: Never depend on deletion. Recover locks by `mv`-ing them aside
(rename works). Untrack files via `git rm --cached` + `.gitignore` instead of
deleting. Legacy code is moved (renamed) into `legacy/`, not removed. Generated
artifacts remain physically on disk but are untracked and gitignored.

**Why**: It is the only reliable path given the FS constraint; it satisfies
Constitution V (no committed build artifacts) without requiring deletes.

**Backport**: delegation_matrix.md "Notes for this environment"; .gitignore.
