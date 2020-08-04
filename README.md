# circuit gen
cd rollup_circuit
go run circuit.go
mv circuit.r1cs ../.

# r1cs setup
//build gnark
gnark setup circuit.r1cs

# L2 transfer_on_prove
//build L2chain
cd L2chain
go build
./L2chain
mv circuit.proof ../.

# L1 transfer_on_verify
//build verify
cd onchain
go build
./verify
