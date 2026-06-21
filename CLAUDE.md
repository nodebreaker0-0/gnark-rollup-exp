# CLAUDE.md — agent entry point

`zkkit`: a Go zk-circuit toolkit on modern `gnark` (v0.15, BN254, Groth16). See
`README.md` for the pitch, `CHARTER.md` for scope, `specs/001-zkkit/` for the
working spec, and `decisions/` for autonomous decisions.

## Folder map

```
prove/         compile→setup→prove→verify harness (start here)
examples/      cubic, mimc, eddsa — circuits + tests (copyable patterns)
rollup/        account/transfer/operator (native) + circuit/assign (in-circuit) + e2e tests
legacy/        original v0.2.1-alpha PoC, separate go.mod, NOT built by ./...
specs/         spec.md / plan.md / tasks.md
scripts/       secrets-scan.sh (part of make verify)
```

## Working here

- `make verify` is the gate: `fmt` + `vet` + `test` + `secrets`. Must be green
  before committing.
- Tests run real Groth16 setup/prove/verify; the full suite takes ~10–15s.
- Go 1.25 toolchain is required (gnark v0.15 needs ≥1.25.7); `go.mod` pins it.

## Environment gotcha (this workspace)

The mounted filesystem **denies file deletion (`unlink`)**. Never rely on `rm`.
Recover a stale `.git/index.lock` by renaming it aside (`mv ... .stale`). Untrack
files with `git rm --cached` + `.gitignore` rather than deleting. See
`decisions/2026-06-21-rebuild-as-zkkit-library.md`.

## Key design notes

- In-circuit `MiMC.Sum()` keeps the chaining value; `eddsa.Verify`/`IsValid` do
  **not** reset the hasher. Always `Reset()` before handing a hasher to them
  (see `rollup/circuit.go: verifySignature`).
- The modern `std/accumulator/merkle.MerkleProof` takes `{RootHash, Path}` with
  `Path[0]` = leaf; the leaf *index* is passed to `VerifyProof` and bit-decomposed
  internally — no separate proof-helper array (unlike the legacy PoC).
