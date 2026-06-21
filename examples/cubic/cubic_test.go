package cubic

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/test"
	"github.com/nodebreaker0-0/gnark-rollup-exp/prove"
)

// TestProveVerify exercises the full Groth16 flow: x=3 gives 27+3+5 = 35.
func TestProveVerify(t *testing.T) {
	if err := prove.Run(&Circuit{}, &Circuit{X: 3, Y: 35}); err != nil {
		t.Fatalf("expected proof to verify: %v", err)
	}
}

// TestConstraints checks witness satisfaction directly: the correct assignment
// solves, a tampered output does not.
func TestConstraints(t *testing.T) {
	field := ecc.BN254.ScalarField()

	if err := test.IsSolved(&Circuit{}, &Circuit{X: 3, Y: 35}, field); err != nil {
		t.Fatalf("valid assignment should solve: %v", err)
	}
	if err := test.IsSolved(&Circuit{}, &Circuit{X: 3, Y: 34}, field); err == nil {
		t.Fatal("tampered assignment (y=34) should not solve, but it did")
	}
}
