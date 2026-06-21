package prove

import (
	"path/filepath"
	"testing"
)

// TestPersistedFlow simulates a real deployment where setup, proving, and
// verification happen in separate steps from on-disk artifacts: compile + setup
// once and save everything, then load and prove, then load and verify — with no
// in-memory objects shared between the stages.
func TestPersistedFlow(t *testing.T) {
	dir := t.TempDir()
	ccsPath := filepath.Join(dir, "circuit.ccs")
	pkPath := filepath.Join(dir, "circuit.pk")
	vkPath := filepath.Join(dir, "circuit.vk")
	proofPath := filepath.Join(dir, "circuit.proof")
	witPath := filepath.Join(dir, "public.witness")

	// --- stage 1: setup, persist artifacts ---
	{
		ccs, err := Compile(&doubler{})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		keys, err := Setup(ccs)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := SaveCCS(ccsPath, ccs); err != nil {
			t.Fatalf("save ccs: %v", err)
		}
		if err := keys.Save(pkPath, vkPath); err != nil {
			t.Fatalf("save keys: %v", err)
		}
	}

	// --- stage 2: load ccs + pk, prove, persist proof + public witness ---
	{
		ccs, err := LoadCCS(ccsPath)
		if err != nil {
			t.Fatalf("load ccs: %v", err)
		}
		keys, err := LoadKeys(pkPath, vkPath)
		if err != nil {
			t.Fatalf("load keys: %v", err)
		}
		proof, public, err := Prove(ccs, keys.PK, &doubler{A: 7, B: 14})
		if err != nil {
			t.Fatalf("prove: %v", err)
		}
		if err := SaveProof(proofPath, proof); err != nil {
			t.Fatalf("save proof: %v", err)
		}
		if err := SavePublicWitness(witPath, public); err != nil {
			t.Fatalf("save witness: %v", err)
		}
	}

	// --- stage 3: load vk + proof + public witness, verify ---
	{
		keys, err := LoadKeys(pkPath, vkPath)
		if err != nil {
			t.Fatalf("load keys: %v", err)
		}
		proof, err := LoadProof(proofPath)
		if err != nil {
			t.Fatalf("load proof: %v", err)
		}
		public, err := LoadPublicWitness(witPath)
		if err != nil {
			t.Fatalf("load witness: %v", err)
		}
		if err := Verify(proof, keys.VK, public); err != nil {
			t.Fatalf("verify from disk should succeed: %v", err)
		}
	}
}
