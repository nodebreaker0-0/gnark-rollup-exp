# tasks.md — 001-zkkit

Status legend: [ ] pending · [~] in progress · [x] done

## P1 — Scaffolding & docs
- [x] T001 Archive legacy v0.2-alpha; gitignore build artifacts
- [x] T002 README (direction), CHARTER, delegation_matrix, constitution
- [x] T003 spec.md / plan.md / tasks.md
- [x] T004 decisions/ entry for the rebuild pivot + the no-unlink FS constraint

## P2 — Toolchain proof  (blocks everything after)
- [x] T010 `go mod init` + add gnark v0.15.0; `go mod download`
- [x] T011 `prove` harness: Compile/Setup/Prove/Verify (Groth16/BN254)
- [x] T012 `examples/cubic` circuit + positive & negative test
- [x] T013 `go test ./...` green; commit

## P3 — Gadget examples
- [x] T020 `examples/mimc` preimage circuit + tests
- [x] T021 `examples/eddsa` in-circuit signature verify + tests
- [x] T022 commit

## P4 — Rollup native types
- [x] T030 `rollup.Account` serialize/deserialize/hash + tests
- [x] T031 `rollup.Transfer` sign/verify (native EdDSA+MiMC) + tests
- [x] T032 `rollup.Operator` state, Merkle proofs, updateState + tests
- [x] T033 commit

## P5 — Rollup circuit + e2e
- [x] T040 `rollup.Circuit` Define on modern std gadgets (membership, sig, update)
- [x] T041 operator witness assignment matching circuit input names
- [x] T042 e2e test: batch=1 prove+verify; negative test
- [x] T043 batch=N parametrization + test
- [x] T044 commit

## P6 — Harden
- [x] T050 Makefile `verify` (fmt, vet, test, secrets grep)
- [x] T051 .github/workflows/ci.yml
- [x] T052 godoc on all exported symbols; examples/README
- [x] T053 go mod tidy; final make verify green; commit

## P7 — Persistence (FR-002 complete)
- [x] T060 prove/io.go: stream + file helpers for CS, pk, vk, proof, public witness
- [x] T061 Keys.Save / LoadKeys
- [x] T062 cross-stage persisted prove/verify test (temp files)
- [x] T063 commit

## Post-MVP backlog (not yet started)
- [ ] T070 PLONK backend option alongside Groth16 (FR-001 centralization)
- [x] T071 extract reusable gadgets (account commitment, leaf+membership) into gadget/
- [ ] T072 benchmarks (constraint counts, prove/verify timings) per circuit
