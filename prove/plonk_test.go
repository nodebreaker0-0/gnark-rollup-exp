package prove

import "testing"

func TestRunPLONK(t *testing.T) {
	if err := RunPLONK(&doubler{}, &doubler{A: 5, B: 10}); err != nil {
		t.Fatalf("expected PLONK proof to verify: %v", err)
	}
}

func TestPLONKStepsAndPublicMismatch(t *testing.T) {
	ccs, err := CompilePLONK(&doubler{})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	keys, err := SetupPLONK(ccs)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	proof, public, err := ProvePLONK(ccs, keys.PK, &doubler{A: 5, B: 10})
	if err != nil {
		t.Fatalf("prove: %v", err)
	}
	if err := VerifyPLONK(proof, keys.VK, public); err != nil {
		t.Fatalf("verify: %v", err)
	}

	otherProof, _, err := ProvePLONK(ccs, keys.PK, &doubler{A: 6, B: 12})
	if err != nil {
		t.Fatalf("prove other: %v", err)
	}
	if err := VerifyPLONK(otherProof, keys.VK, public); err == nil {
		t.Fatal("verify should fail when public inputs differ")
	}
}
