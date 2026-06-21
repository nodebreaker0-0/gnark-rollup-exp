# Bring Solidity verifier export into scope (deployment stays out)

**Date**: 2026-06-21
**Context**: The original PoC's "onchain" step verified proofs in Go, not on a
real chain. The CHARTER listed any on-chain verifier as out of scope. For a
rollup library, being able to verify proofs on-chain is a core value, and gnark
can emit a Solidity verifier from a Groth16 BN254 verifying key.

**Options considered**:
- A. Keep all on-chain work out of scope. — Leaves an obvious gap for a rollup.
- B. Support *generating* the verifier contract, but keep deployment/gas-tuning
  out of scope. — Delivers the reusable artifact without committing to chain ops.

**Decision**: B.

**Why**: `vk.ExportSolidity` is a thin, well-tested gnark feature; exposing it
(`prove.ExportSolidityVerifier`/`SaveSolidityVerifier`) lets users take a proof
on-chain with no extra dependency, while deployment specifics (network, gas,
proxy patterns) remain genuinely consumer-specific and out of scope. Curve is
BN254, which is what Ethereum's precompiles support.

**Backport**: CHARTER.md Out-of-scope; prove/io.go; prove/solidity_test.go;
CHANGELOG.
