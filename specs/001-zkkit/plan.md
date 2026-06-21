# plan.md — 001-zkkit

## Tech stack (decided)

- Go 1.23, single module `github.com/nodebreaker0-0/gnark-rollup-exp`.
- `gnark v0.15.0` (latest as of 2026-05) pinned in `go.mod`.
- Curve BN254, backend Groth16. Std gadgets: `std/hash/mimc`,
  `std/signature/eddsa`, `std/accumulator/merkle`,
  `std/algebra/native/twistededwards`.
- Native crypto: `gnark-crypto` (`ecc/bn254/...`) for off-circuit hashing,
  signing, and Merkle trees.

## Architecture

- `internal/zkconf` (or a small `prove` package) centralizes curve + backend so
  examples don't hardcode them (FR-001).
- `prove` package: `Compile`, `Setup`, `Prove`, `Verify`, plus serialization
  helpers. Thin wrappers over gnark, but the single place a user learns the flow.
- `examples/<name>`: each its own package with a `Circuit` type and a test. Kept
  importable and runnable.
- `rollup` package: `Account`, `Transfer`, `Operator` (native), and `Circuit`
  (in-circuit) + constants for witness names. Operator builds the witness.

## Phases

1. **P1 Scaffolding & docs** — CHARTER, constitution, spec, README, delegation,
   gitignore, legacy archive. (this commit set)
2. **P2 Toolchain proof** — go.mod + `examples/cubic` + harness skeleton, real
   prove/verify test green. Validates gnark v0.15 API end to end.
3. **P3 Gadget examples** — `mimc`, `eddsa` with tests.
4. **P4 Rollup native** — account/transfer/operator + unit tests.
5. **P5 Rollup circuit** — `Define` + operator witness + e2e prove/verify test.
6. **P6 Harden** — Makefile verify, CI, godoc, tidy, final review.

## Risks

- gnark v0.15 API may differ from training-era knowledge (v0.7–v0.9). Mitigation:
  let `go build`/`go test` drive correction; pin version; consult pkg.go.dev if a
  symbol is missing.
- Merkle proof helper API changed between PoC and modern std/accumulator/merkle.
  Mitigation: build the rollup circuit against the modern gadget from scratch,
  not by line-porting the PoC.
- Build/compile time of gnark in the sandbox may exceed single-call timeouts.
  Mitigation: pre-warm module cache with `go mod download` in its own step.
