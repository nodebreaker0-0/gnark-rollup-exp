# Constitution — zkkit

Executable principles. `make verify` and review enforce these. Hard rules are
non-negotiable without a `decisions/` entry.

## I. Build green or it didn't happen
`go build ./...` and `go test ./...` must pass on a clean checkout. A commit that
breaks the build is reverted, not patched on top.

## II. Every example is a test
No example circuit ships without a test that runs setup → prove → verify and
asserts success, plus a negative test asserting a bad witness fails. Examples are
the spec made executable.

## III. Minimize dependencies
Direct dependencies beyond `gnark` (and its transitive set) require a Notify-only
note in `decisions/`. Anything touching wallets, key custody, or secrets is Block.

## IV. No secrets, ever
No private keys, API keys, tokens, or mnemonics in code, tests, fixtures, logs,
or committed artifacts. Test keys are derived in-process from fixed public seeds
and never persisted. `make verify` greps for secret patterns and fails on a hit.

## V. No committed build artifacts
`*.r1cs`, `*.pk`, `*.vk`, `*.proof`, `*.ccs`, `input.public`, binaries — all
gitignored. Anything a build can regenerate is not tracked.

## VI. Document the public surface
Every exported identifier in library packages has a godoc comment. `gofmt -l`
and `go vet ./...` report nothing.

## VII. Determinism
Tests are deterministic: fixed seeds, no wall-clock or network dependence in
assertions, pinned `gnark` version in `go.mod`.

## VIII. Log autonomous decisions
Any non-trivial choice made without the human goes in `decisions/YYYY-MM-DD-*.md`
with options, decision, and rationale.

## IX. Rollback-able
Each change is a focused commit with a clear message. No history rewrites, no
force-push. Reverting one commit undoes one logical change.
