package prove

import (
	"fmt"
	"io"
	"os"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
)

// This file adds persistence so the proving flow can span processes: run Setup
// once, store the keys, then prove and verify later from the serialized
// artifacts — the way a real deployment separates the (expensive, one-time)
// setup from routine proving and verification.

// --- stream helpers (io.Writer / io.Reader) ---

// WriteCCS serializes a constraint system.
func WriteCCS(w io.Writer, ccs constraint.ConstraintSystem) (int64, error) {
	return ccs.WriteTo(w)
}

// ReadCCS deserializes a constraint system for the zkkit curve.
func ReadCCS(r io.Reader) (constraint.ConstraintSystem, error) {
	ccs := groth16.NewCS(Curve)
	if _, err := ccs.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("read constraint system: %w", err)
	}
	return ccs, nil
}

// WriteProvingKey serializes a proving key.
func WriteProvingKey(w io.Writer, pk groth16.ProvingKey) (int64, error) {
	return pk.WriteTo(w)
}

// ReadProvingKey deserializes a proving key for the zkkit curve.
func ReadProvingKey(r io.Reader) (groth16.ProvingKey, error) {
	pk := groth16.NewProvingKey(Curve)
	if _, err := pk.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("read proving key: %w", err)
	}
	return pk, nil
}

// WriteVerifyingKey serializes a verifying key.
func WriteVerifyingKey(w io.Writer, vk groth16.VerifyingKey) (int64, error) {
	return vk.WriteTo(w)
}

// ReadVerifyingKey deserializes a verifying key for the zkkit curve.
func ReadVerifyingKey(r io.Reader) (groth16.VerifyingKey, error) {
	vk := groth16.NewVerifyingKey(Curve)
	if _, err := vk.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("read verifying key: %w", err)
	}
	return vk, nil
}

// WriteProof serializes a proof.
func WriteProof(w io.Writer, proof groth16.Proof) (int64, error) {
	return proof.WriteTo(w)
}

// ReadProof deserializes a proof for the zkkit curve.
func ReadProof(r io.Reader) (groth16.Proof, error) {
	proof := groth16.NewProof(Curve)
	if _, err := proof.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("read proof: %w", err)
	}
	return proof, nil
}

// WriteWitness serializes a witness (full or public).
func WriteWitness(w io.Writer, wit witness.Witness) (int64, error) {
	return wit.WriteTo(w)
}

// ReadPublicWitness deserializes a public witness for the zkkit curve.
func ReadPublicWitness(r io.Reader) (witness.Witness, error) {
	wit, err := witness.New(Curve.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("new witness: %w", err)
	}
	if _, err := wit.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("read witness: %w", err)
	}
	return wit, nil
}

// --- file-path convenience wrappers ---

func writeToFile(path string, write func(io.Writer) (int64, error)) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := write(f); err != nil {
		return err
	}
	return f.Close()
}

func readFromFile[T any](path string, read func(io.Reader) (T, error)) (T, error) {
	var zero T
	f, err := os.Open(path)
	if err != nil {
		return zero, err
	}
	defer f.Close()
	return read(f)
}

// Save writes the key pair to pkPath and vkPath.
func (k Keys) Save(pkPath, vkPath string) error {
	if err := writeToFile(pkPath, func(w io.Writer) (int64, error) { return WriteProvingKey(w, k.PK) }); err != nil {
		return fmt.Errorf("save proving key: %w", err)
	}
	if err := writeToFile(vkPath, func(w io.Writer) (int64, error) { return WriteVerifyingKey(w, k.VK) }); err != nil {
		return fmt.Errorf("save verifying key: %w", err)
	}
	return nil
}

// LoadKeys reads a key pair from pkPath and vkPath.
func LoadKeys(pkPath, vkPath string) (Keys, error) {
	pk, err := readFromFile(pkPath, ReadProvingKey)
	if err != nil {
		return Keys{}, fmt.Errorf("load proving key: %w", err)
	}
	vk, err := readFromFile(vkPath, ReadVerifyingKey)
	if err != nil {
		return Keys{}, fmt.Errorf("load verifying key: %w", err)
	}
	return Keys{PK: pk, VK: vk}, nil
}

// SaveCCS writes a constraint system to path.
func SaveCCS(path string, ccs constraint.ConstraintSystem) error {
	return writeToFile(path, func(w io.Writer) (int64, error) { return WriteCCS(w, ccs) })
}

// LoadCCS reads a constraint system from path.
func LoadCCS(path string) (constraint.ConstraintSystem, error) {
	return readFromFile(path, ReadCCS)
}

// SaveProof writes a proof to path.
func SaveProof(path string, proof groth16.Proof) error {
	return writeToFile(path, func(w io.Writer) (int64, error) { return WriteProof(w, proof) })
}

// LoadProof reads a proof from path.
func LoadProof(path string) (groth16.Proof, error) {
	return readFromFile(path, ReadProof)
}

// SavePublicWitness writes a (public) witness to path.
func SavePublicWitness(path string, wit witness.Witness) error {
	return writeToFile(path, func(w io.Writer) (int64, error) { return WriteWitness(w, wit) })
}

// LoadPublicWitness reads a public witness from path.
func LoadPublicWitness(path string) (witness.Witness, error) {
	return readFromFile(path, ReadPublicWitness)
}

// --- on-chain verifier export ---

// ExportSolidityVerifier writes a Solidity verifier contract for a Groth16
// verifying key to w. Only BN254 is supported (which is zkkit's curve). The
// emitted contract verifies proofs produced for the same circuit on-chain.
func ExportSolidityVerifier(w io.Writer, vk groth16.VerifyingKey) error {
	if err := vk.ExportSolidity(w); err != nil {
		return fmt.Errorf("export solidity verifier: %w", err)
	}
	return nil
}

// SaveSolidityVerifier writes the Solidity verifier contract to path.
func SaveSolidityVerifier(path string, vk groth16.VerifyingKey) error {
	return writeToFile(path, func(w io.Writer) (int64, error) {
		return 0, ExportSolidityVerifier(w, vk)
	})
}
