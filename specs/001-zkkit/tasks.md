# tasks.md — 001-zkkit

Status legend: [ ] pending · [~] in progress · [x] done

## P1 — Scaffolding & docs
- [x] T001 Archive legacy v0.2-alpha; gitignore build artifacts
- [x] T002 README (direction), CHARTER, delegation_matrix, constitution
- [x] T003 spec.md / plan.md / tasks.md
- [x] T004 decisions/ entry for the rebuild pivot + the no-unlink FS constraint

## P2 — Toolchain proof  (blocks everything after)
- [ ] T010 `go mod init` + add gnark v0.15.0; `go mod download`
- [ ] T011 `prove` harness: Compile/Setup/Prove/Verify (Groth16/BN254)
- [ ] T012 `examples/cubic` circuit + positive & negative test
- [ ] T013 `go test ./...` green; commit

## P3 — Gadget examples
- [ ] T020 `examples/mimc` preimage circuit + tests
- [ ] T021 `examples/eddsa` in-circuit signature verify + tests
- [ ] T022 commit

## P4 — Rollup native types
- [ ] T030 `rollup.Account` serialize/deserialize/hash + tests
- [ ] T031 `rollup.Transfer` sign/verify (native EdDSA+MiMC) + tests
- [ ] T032 `rollup.Operator` state, Merkle proofs, updateState + tests
- [ ] T033 commit

## P5 — Rollup circuit + e2e
- [ ] T040 `rollup.Circuit` Define on modern std gadgets (membership, sig, update)
- [ ] T041 operator witness assignment matching circuit input names
- [ ] T042 e2e test: batch=1 prove+verify; negative test
- [ ] T043 batch=N parametrization + test
- [ ] T044 commit

## P6 — Harden
- [ ] T050 Makefile `verify` (fmt, vet, test, secrets grep)
- [ ] T051 .github/workflows/ci.yml
- [ ] T052 godoc on all exported symbols; examples/README
- [ ] T053 go mod tidy; final make verify green; commit
