package prove

import (
	"testing"

	"github.com/consensys/gnark/frontend"
)

// doubler proves 2*A == B, with B public.
type doubler struct {
	A frontend.Variable
	B frontend.Variable `gnark:",public"`
}

func (c *doubler) Define(api frontend.API) error {
	api.AssertIsEqual(api.Add(c.A, c.A), c.B)
	return nil
}

func TestRun(t *testing.T) {
	if err := Run(&doubler{}, &doubler{A: 3, B: 6}); err != nil {
		t.Fatalf("expected proof to verify: %v", err)
	}
}

func TestStepsAndPublicMismatch(t *testing.T) {
	ccs, err := Compile(&doubler{})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	keys, err := Setup(ccs)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	proof, public, err := Prove(ccs, keys.PK, &doubler{A: 3, B: 6})
	if err != nil {
		t.Fatalf("prove: %v", err)
	}
	if err := Verify(proof, keys.VK, public); err != nil {
		t.Fatalf("verify: %v", err)
	}

	// A proof for B=8 must not verify against the public witness for B=6.
	otherProof, _, err := Prove(ccs, keys.PK, &doubler{A: 4, B: 8})
	if err != nil {
		t.Fatalf("prove other: %v", err)
	}
	if err := Verify(otherProof, keys.VK, public); err == nil {
		t.Fatal("verify should fail when public inputs differ")
	}
}
