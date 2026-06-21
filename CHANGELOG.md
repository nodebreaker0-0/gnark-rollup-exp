# Changelog

All notable changes to this project are documented here. This project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **PLONK backend** in `prove` (`CompilePLONK`, `SetupPLONK`, `ProvePLONK`,
  `VerifyPLONK`, `RunPLONK`) alongside Groth16, using the scs builder and a
  development-only KZG SRS. The rollup circuit is cross-checked under both
  backends.
- **Benchmarks & sizing**: `TestReportConstraints` logs circuit sizes and
  `BenchmarkProveGroth16`/`BenchmarkProvePLONK` time proving for the rollup.
- **Solidity verifier export**: `prove.ExportSolidityVerifier` /
  `SaveSolidityVerifier` emit an on-chain Groth16 verifier contract (BN254).
- Rollup generality test covering a batch of distinct sender/receiver pairs.

## [v0.2.0] — 2026-06-21

First release of the rebuilt `zkkit` library on modern `gnark`. This is a
ground-up rewrite of the original 2020-era proof of concept; the old API is gone
and preserved only under `legacy/`.

### Added
- **`prove`** — a compile → setup → prove → verify harness over Groth16 on BN254
  (`Compile`, `Setup`, `Prove`, `Verify`, `Run`), plus persistence for the
  constraint system, proving/verifying keys, proof, and public witness
  (`Keys.Save`/`LoadKeys`, `SaveProof`/`LoadProof`, `SaveCCS`/`LoadCCS`,
  `SavePublicWitness`/`LoadPublicWitness`) so setup, proving, and verification can
  run as separate steps from on-disk artifacts.
- **`gadget`** — reusable in-circuit building blocks: `Account.Commit` (MiMC
  account commitment) and `VerifyMembership` (binds a commitment to a Merkle leaf
  and proof to a public root, then checks inclusion).
- **`rollup`** — an account-based zk-rollup reference: native `Account`,
  `Transfer`, and `Operator` (MiMC Merkle state, EdDSA transfers, batch
  application producing a `TransferWitness`) plus the in-circuit `Circuit` and its
  `Assign`, proving a batch of transfers was applied correctly.
- **`examples`** — runnable, tested circuits: `cubic`, `mimc`, `eddsa`.
- Tooling: `make verify` gate (fmt, vet, test, secrets scan), GitHub Actions CI
  and release workflows, spec-kit documents under `specs/`, and a decision log.

### Notes
- Curve BN254, backend Groth16. PLONK is planned (see `specs/001-zkkit/tasks.md`).
- Requires Go ≥ 1.25 (gnark v0.15 needs ≥ 1.25.7); the toolchain is pinned in
  `go.mod`.
- The trusted setup performed by `prove.Setup` is an in-process, development-only
  setup — not a ceremony. Do not use its keys in production.

## [v0.1.0] — legacy

A "test release" tag on the original `gnark v0.2.1-alpha` proof of concept
(`Initial commit`). Kept for history; the code lives under `legacy/v0.2-alpha/`
and does not build with modern gnark.

[v0.2.0]: https://github.com/nodebreaker0-0/gnark-rollup-exp/releases/tag/v0.2.0
[v0.1.0]: https://github.com/nodebreaker0-0/gnark-rollup-exp/releases/tag/v0.1.0
