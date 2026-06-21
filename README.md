# zkkit — write your zk circuits in Go

`zkkit` is a Go module for building zero-knowledge circuits and proofs. You write
the circuit as a plain Go struct, and the toolkit compiles it, runs a Groth16
setup, produces a proof, and verifies it — all from Go, with no external prover
toolchain.

It is built **on top of [`gnark`](https://github.com/consensys/gnark)** (the
mature ConsenSys zk-SNARK library), not as a replacement for it. `gnark` gives us
the field arithmetic, the constraint compiler, and the Groth16/PLONK backends.
`zkkit` adds the layer most projects end up re-implementing by hand: reusable
in-circuit gadgets, a clean prove/verify harness, a complete and modern zk-rollup
reference implementation, and a library of runnable example circuits you can copy
from.

> **Status:** active rebuild. This repository started as a 2020-era proof of
> concept on `gnark v0.2.1-alpha` (preserved under [`legacy/`](./legacy)). It is
> being rebuilt into a tested, documented, importable Go library on modern
> `gnark`. See [`CHARTER.md`](./CHARTER.md) for scope and
> [`specs/001-zkkit/`](./specs/001-zkkit) for the working spec.

## Why this exists

If you want to prove something in Go today, `gnark` is the answer — and `zkkit`
does not try to compete with it. What `gnark` deliberately leaves to you is the
application layer:

- **Gadgets** — composable circuit building blocks (account commitments, Merkle
  membership, signature checks) that you wire together instead of writing from
  raw constraints every time.
- **A prove/verify harness** — one call that compiles a circuit, runs setup,
  proves a witness, and verifies, with the keys and proofs serialized to disk.
- **A reference application** — a working account-based zk-rollup showing how the
  gadgets compose into something real, on the current API.
- **Examples** — small, self-contained circuits (`cubic`, `mimc`, `eddsa`,
  `rollup`) that double as the test suite.

The goal is that a Go developer who has never written a circuit can read one
example, copy it, and have a verified proof in minutes.

## Install

```bash
go get github.com/nodebreaker0-0/gnark-rollup-exp@v0.2.0
```

Requires Go ≥ 1.25 (gnark v0.15 needs ≥ 1.25.7). See [`CHANGELOG.md`](./CHANGELOG.md)
for releases.

## Quick start

```bash
go test ./...          # build every example + run the full suite
make verify            # fmt + vet + test + secrets scan (the CI gate)
```

A circuit is a Go struct that declares its inputs and implements `Define`:

```go
// x^3 + x + 5 == y   (the zk "hello world")
type CubicCircuit struct {
    X frontend.Variable `gnark:"x"`              // secret by default
    Y frontend.Variable `gnark:"y,public"`       // public input
}

func (c *CubicCircuit) Define(api frontend.API) error {
    x3 := api.Mul(c.X, c.X, c.X)
    api.AssertIsEqual(c.Y, api.Add(x3, c.X, 5))
    return nil
}
```

Proving and verifying it is a single helper call (see
[`examples/cubic`](./examples/cubic)).

## Layout

```
zkkit/
├── examples/        runnable example circuits (cubic, mimc, eddsa, rollup)
├── rollup/          the zk-rollup reference library (accounts, transfers, operator, circuit)
├── prove/           the compile → setup → prove → verify harness (+ key/proof persistence)
├── gadget/          reusable in-circuit gadgets (account commitment, Merkle membership)
├── legacy/          the original v0.2.1-alpha PoC, kept for reference
├── specs/           spec-kit working documents
└── decisions/       autonomous decision log
```

## Benchmarks

```bash
go test -run TestReportConstraints -v ./rollup   # circuit sizes
go test -run '^$' -bench BenchmarkProve ./rollup # proving timings
```

The rollup circuit is ~29.8k R1CS constraints per transfer (linear in batch
size). On the reference machine a single-transfer Groth16 proof takes ~0.24s;
the same circuit under PLONK takes ~1.1s.

## Relationship to `gnark`

`zkkit` re-exports nothing and hides nothing. Your circuits use `frontend.API`
and the `gnark` standard gadgets directly; `zkkit` only adds higher-level pieces
on top. If you outgrow the toolkit, you drop down to plain `gnark` with zero
migration. The curve is BN254 and the default backend is Groth16, matching the
original PoC; PLONK support is tracked in the spec.

## License

The reference code derived from the ConsenSys `gnark` examples retains its
Apache-2.0 headers. See individual files.
