# CHARTER — zkkit

**One line:** a tested, documented Go library for writing zk circuits and
producing/verifying proofs, built on modern `gnark`, with a working zk-rollup
reference and copyable examples.

## Purpose

Turn the 2020-era `gnark-rollup-exp` proof of concept into a maintainable Go
module that a developer can `go get` and use to (a) write their own circuits
against a clean prove/verify harness and (b) study a complete account-based
zk-rollup on the current API.

## Users

- **Primary:** B-Harvest engineers prototyping zk applications in Go.
- **Secondary:** any Go developer learning gnark who wants runnable, modern,
  tested examples instead of stale gists.

## Success criteria (measurable)

- **SC-001** `go test ./...` builds and passes on a clean checkout with only
  `go` + network for module download. No manual key files required.
- **SC-002** Every example circuit (`cubic`, `mimc`, `eddsa`, `rollup`) has a
  test that runs a real Groth16 setup → prove → verify and asserts the proof
  verifies, plus a negative test asserting a tampered witness fails.
- **SC-003** The rollup reference proves a batch of ≥1 transfer end to end and
  verifies, reproducing the PoC's behaviour on modern `gnark`.
- **SC-004** `make verify` (fmt, vet, test, secrets scan) is green and runs in
  CI on every push.
- **SC-005** Every exported identifier in `rollup/` and the harness has a godoc
  comment; `go vet` and `gofmt -l` report nothing.

## MVP definition

A single Go module that compiles, with: the prove/verify harness, the four
example circuits with passing positive + negative tests, the rollup library
(native types + circuit) with an end-to-end test, a green `make verify`, and CI.
Gadget extraction and PLONK are post-MVP.

## Out of scope

- Reimplementing any `gnark` primitive (field math, constraint compiler, backends).
- On-chain Solidity verifier *deployment* (gas-tuned contracts, network deploy
  scripts). Generating the verifier contract is supported (see
  `prove.ExportSolidityVerifier`); deploying and integrating it on a real chain
  is left to the consumer. (Revised 2026-06-21; see decisions/.)
- A general CLI / daemon. The deliverable is a library + tests + examples.
- Trusted-setup ceremony tooling. Tests use throwaway in-process setup.
- New cryptography. We use BN254 + Groth16 + MiMC + EdDSA as the PoC did.

## Safety constraints

- **No secrets in the repo.** No private keys, API keys, or tokens in code,
  fixtures, logs, or committed artifacts. Test keys are derived deterministically
  in-process from fixed seeds and never written to disk.
- **No committed proving keys.** `.pk/.vk/.r1cs/.proof` are build artifacts,
  gitignored and regenerable.
- **Reproducible.** Pinned `gnark` version in `go.mod`; tests deterministic.

## Trade-off preferences

- Correctness and clarity over proving speed (this is a reference/library).
- Standard `gnark` gadgets over hand-rolled constraints wherever they exist.
- Small dependency surface: `gnark` + its transitive deps, nothing else of
  substance. New direct deps are Notify-only; anything wallet/secret-related is
  Block.
- Prefer breaking cleanly from the legacy API over compatibility shims.
