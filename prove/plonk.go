package prove

import (
	"fmt"

	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
)

// This file mirrors the Groth16 harness for the PLONK backend. PLONK compiles to
// a SparseR1CS (via the scs builder) and needs a KZG structured reference string.
//
// SetupPLONK builds that SRS with unsafekzg, which is a deterministic,
// development-only generator — it does NOT run a real KZG ceremony and its keys
// must not be used in production.

// PlonkKeys bundles a PLONK proving/verifying key pair.
type PlonkKeys struct {
	PK plonk.ProvingKey
	VK plonk.VerifyingKey
}

// CompilePLONK compiles a circuit to a PLONK (SparseR1CS) constraint system.
func CompilePLONK(circuit frontend.Circuit) (constraint.ConstraintSystem, error) {
	ccs, err := frontend.Compile(Curve.ScalarField(), scs.NewBuilder, circuit)
	if err != nil {
		return nil, fmt.Errorf("compile circuit (plonk): %w", err)
	}
	return ccs, nil
}

// SetupPLONK runs a PLONK setup for a compiled circuit using a development-only
// (unsafe) KZG SRS. See the file documentation.
func SetupPLONK(ccs constraint.ConstraintSystem) (PlonkKeys, error) {
	srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
	if err != nil {
		return PlonkKeys{}, fmt.Errorf("build kzg srs: %w", err)
	}
	pk, vk, err := plonk.Setup(ccs, srs, srsLagrange)
	if err != nil {
		return PlonkKeys{}, fmt.Errorf("plonk setup: %w", err)
	}
	return PlonkKeys{PK: pk, VK: vk}, nil
}

// ProvePLONK builds the witness from an assignment and produces a PLONK proof,
// returning it with the public-only witness needed for verification.
func ProvePLONK(ccs constraint.ConstraintSystem, pk plonk.ProvingKey, assignment frontend.Circuit) (plonk.Proof, witness.Witness, error) {
	full, err := frontend.NewWitness(assignment, Curve.ScalarField())
	if err != nil {
		return nil, nil, fmt.Errorf("build witness: %w", err)
	}
	public, err := full.Public()
	if err != nil {
		return nil, nil, fmt.Errorf("extract public witness: %w", err)
	}
	proof, err := plonk.Prove(ccs, pk, full)
	if err != nil {
		return nil, nil, fmt.Errorf("plonk prove: %w", err)
	}
	return proof, public, nil
}

// VerifyPLONK checks a PLONK proof against a verifying key and public witness.
func VerifyPLONK(proof plonk.Proof, vk plonk.VerifyingKey, publicWitness witness.Witness) error {
	if err := plonk.Verify(proof, vk, publicWitness); err != nil {
		return fmt.Errorf("plonk verify: %w", err)
	}
	return nil
}

// RunPLONK performs the full compile → setup → prove → verify flow with the PLONK
// backend, returning nil when the proof verifies.
func RunPLONK(circuit, assignment frontend.Circuit) error {
	ccs, err := CompilePLONK(circuit)
	if err != nil {
		return err
	}
	keys, err := SetupPLONK(ccs)
	if err != nil {
		return err
	}
	proof, publicWitness, err := ProvePLONK(ccs, keys.PK, assignment)
	if err != nil {
		return err
	}
	return VerifyPLONK(proof, keys.VK, publicWitness)
}
