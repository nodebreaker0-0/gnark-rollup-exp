// Package prove is a thin compile → setup → prove → verify harness over gnark's
// Groth16 backend on the BN254 curve. It is the single place a zkkit user learns
// the proving flow; every example circuit drives its tests through these helpers.
//
// The Setup performed here is an in-process, throwaway trusted setup intended for
// development and testing. It is NOT a multi-party ceremony and must not be used
// to generate keys for a production deployment.
package prove

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// Curve is the elliptic curve used across zkkit. BN254 matches the original
// proof of concept and is the most widely supported pairing-friendly curve for
// on-chain verification.
var Curve = ecc.BN254

// Keys bundles a Groth16 proving/verifying key pair produced by Setup.
type Keys struct {
	PK groth16.ProvingKey
	VK groth16.VerifyingKey
}

// Compile builds the R1CS constraint system for a circuit definition.
func Compile(circuit frontend.Circuit) (constraint.ConstraintSystem, error) {
	ccs, err := frontend.Compile(Curve.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		return nil, fmt.Errorf("compile circuit: %w", err)
	}
	return ccs, nil
}

// Setup runs a Groth16 trusted setup for a compiled circuit. See the package
// documentation: this is a development-only setup, not a ceremony.
func Setup(ccs constraint.ConstraintSystem) (Keys, error) {
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return Keys{}, fmt.Errorf("groth16 setup: %w", err)
	}
	return Keys{PK: pk, VK: vk}, nil
}

// Prove builds the witness from an assignment and produces a proof. It returns
// the proof together with the public-only witness needed for verification.
func Prove(ccs constraint.ConstraintSystem, pk groth16.ProvingKey, assignment frontend.Circuit) (groth16.Proof, witness.Witness, error) {
	full, err := frontend.NewWitness(assignment, Curve.ScalarField())
	if err != nil {
		return nil, nil, fmt.Errorf("build witness: %w", err)
	}
	public, err := full.Public()
	if err != nil {
		return nil, nil, fmt.Errorf("extract public witness: %w", err)
	}
	proof, err := groth16.Prove(ccs, pk, full)
	if err != nil {
		return nil, nil, fmt.Errorf("groth16 prove: %w", err)
	}
	return proof, public, nil
}

// Verify checks a proof against a verifying key and the public witness. It
// returns nil if and only if the proof is valid.
func Verify(proof groth16.Proof, vk groth16.VerifyingKey, publicWitness witness.Witness) error {
	if err := groth16.Verify(proof, vk, publicWitness); err != nil {
		return fmt.Errorf("groth16 verify: %w", err)
	}
	return nil
}

// Run performs the full compile → setup → prove → verify flow for a circuit and
// a satisfying assignment. It returns nil when the resulting proof verifies,
// making it a one-call check that a circuit and witness are consistent and
// provable end to end.
func Run(circuit, assignment frontend.Circuit) error {
	ccs, err := Compile(circuit)
	if err != nil {
		return err
	}
	keys, err := Setup(ccs)
	if err != nil {
		return err
	}
	proof, publicWitness, err := Prove(ccs, keys.PK, assignment)
	if err != nil {
		return err
	}
	return Verify(proof, keys.VK, publicWitness)
}
