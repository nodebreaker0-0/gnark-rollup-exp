package prove

import (
	"bytes"
	"strings"
	"testing"
)

// TestExportSolidityVerifier checks that a Groth16 verifying key emits a
// plausible Solidity verifier contract (it does not compile it).
func TestExportSolidityVerifier(t *testing.T) {
	ccs, err := Compile(&doubler{})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	keys, err := Setup(ccs)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	var buf bytes.Buffer
	if err := ExportSolidityVerifier(&buf, keys.VK); err != nil {
		t.Fatalf("export: %v", err)
	}

	sol := buf.String()
	for _, want := range []string{"pragma solidity", "contract", "function verifyProof"} {
		if !strings.Contains(sol, want) {
			t.Fatalf("solidity output missing %q; got %d bytes", want, len(sol))
		}
	}
}
