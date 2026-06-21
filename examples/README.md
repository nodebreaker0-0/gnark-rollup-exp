# Examples

Each example is a self-contained package with a circuit and its tests. The tests
double as documentation: they show how to build a witness and run a real Groth16
prove/verify through the [`prove`](../prove) harness.

Run them all:

```bash
go test ./examples/...
```

| Example | What it proves | Gadgets used |
|---|---|---|
| [`cubic`](./cubic) | knowledge of secret `x` with `x³ + x + 5 == y` | raw arithmetic (`api.Mul`, `api.Add`) |
| [`mimc`](./mimc) | knowledge of a MiMC hash preimage | `std/hash/mimc` |
| [`eddsa`](./eddsa) | a valid EdDSA signature over a message | `std/signature/eddsa`, `std/algebra/native/twistededwards` |

For a full application that composes these ideas — account commitments, Merkle
membership, signature checks, and state updates in one SNARK — see the
[`rollup`](../rollup) package.

## Anatomy of an example

1. Define a struct whose fields are the circuit inputs. Tag public inputs with
   `gnark:",public"`; everything else is secret.
2. Implement `Define(api frontend.API) error` with the constraints.
3. In the test, build a satisfying assignment and call `prove.Run(&Circuit{}, &assignment)`.
   A negative test feeds a bad assignment to `test.IsSolved` and asserts it fails.
