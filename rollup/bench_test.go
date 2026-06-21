package rollup

import (
	"testing"

	"github.com/nodebreaker0-0/gnark-rollup-exp/prove"
)

// TestReportConstraints logs the R1CS constraint count of the rollup circuit for
// a few batch sizes. Run with `go test -run TestReportConstraints -v ./rollup`.
func TestReportConstraints(t *testing.T) {
	for _, batch := range []int{1, 2, 3} {
		witnesses, pathLen := buildBatch(t, 16, batch)
		ccs, err := prove.Compile(New(len(witnesses), pathLen))
		if err != nil {
			t.Fatalf("compile batch %d: %v", batch, err)
		}
		t.Logf("rollup batch=%d: %d R1CS constraints (%d public)", batch, ccs.GetNbConstraints(), ccs.GetNbPublicVariables())
	}
}

// BenchmarkProveGroth16 measures proving time for a single-transfer batch. Setup
// is done once outside the timed loop.
func BenchmarkProveGroth16(b *testing.B) {
	witnesses, pathLen := buildBatch(b, 16, 1)
	ccs, err := prove.Compile(New(1, pathLen))
	if err != nil {
		b.Fatalf("compile: %v", err)
	}
	keys, err := prove.Setup(ccs)
	if err != nil {
		b.Fatalf("setup: %v", err)
	}
	assignment := Assign(witnesses, pathLen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := prove.Prove(ccs, keys.PK, assignment); err != nil {
			b.Fatalf("prove: %v", err)
		}
	}
}

// BenchmarkProvePLONK measures proving time under the PLONK backend.
func BenchmarkProvePLONK(b *testing.B) {
	witnesses, pathLen := buildBatch(b, 16, 1)
	ccs, err := prove.CompilePLONK(New(1, pathLen))
	if err != nil {
		b.Fatalf("compile: %v", err)
	}
	keys, err := prove.SetupPLONK(ccs)
	if err != nil {
		b.Fatalf("setup: %v", err)
	}
	assignment := Assign(witnesses, pathLen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := prove.ProvePLONK(ccs, keys.PK, assignment); err != nil {
			b.Fatalf("prove: %v", err)
		}
	}
}
