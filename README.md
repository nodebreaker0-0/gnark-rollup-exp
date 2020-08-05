# circuit gen
```bash
cd rollup_circuit
go run circuit.go
mv circuit.r1cs ../.
```

# r1cs setup
```bash
cd setup
go run setup.go
```

# L2 transfer_on_prove
```bash
cd L2chain
go run .
mv circuit.proof ../.
mv input.public ../.
```

# L1 transfer_on_verify
```bash
cd onchain
go run .
```
