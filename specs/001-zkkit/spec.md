# spec.md — 001-zkkit

## What we are building

A Go module (`zkkit`) that lets a developer write zk circuits in Go and
prove/verify them, built on `gnark`, with a zk-rollup reference and example
circuits. See `CHARTER.md` for purpose and out-of-scope.

## User stories

- **US-001** As a Go dev new to zk, I copy `examples/cubic`, change the
  constraint, and get a verified proof without learning a new toolchain.
- **US-002** As a dev, I call one harness function to compile → setup → prove →
  verify any circuit implementing `frontend.Circuit`.
- **US-003** As a dev studying rollups, I read `rollup/` and run its e2e test to
  see how account state, Merkle proofs, signatures, and balance updates compose
  into one SNARK.
- **US-004** As a maintainer, I run `make verify` and trust a green result to
  mean the library is shippable.

## Functional requirements

- **FR-001** Module targets BN254 + Groth16 by default; curve/backend choices are
  centralized so PLONK can be added later without touching example circuits.
- **FR-002** A `prove` harness exposes: compile a circuit to a constraint system,
  run setup (pk/vk), prove a witness, verify a proof. Keys/proofs serialize to
  and load from `io.Writer`/`io.Reader` (and helper file paths).
- **FR-003** Example circuits: `cubic` (arithmetic), `mimc` (hash preimage),
  `eddsa` (in-circuit signature verification), `rollup` (full reference).
- **FR-004** Rollup library provides native (out-of-circuit) types — account,
  transfer, operator with Merkle-backed state — and the in-circuit `Circuit`
  with `Define`, matching the PoC's semantics: membership before/after, signature
  validity, nonce+1, amount ≤ balance, balance transfer.
- **FR-005** Witness assignment for the rollup is produced by the operator from a
  processed transfer batch, so the e2e test needs no hand-built witness.
- **FR-006** Every circuit has positive and negative (`test.IsSolved` /
  tampered-witness) tests.

## Success criteria

See `CHARTER.md` SC-001 … SC-005. Tracked there to avoid drift.

## Open questions / deferred

- PLONK backend (deferred, post-MVP).
- Gadget package extraction from the rollup circuit (deferred until the rollup
  circuit is green, so we extract from working code).
- Real Solidity verifier (out of scope).
