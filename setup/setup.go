package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	backend_bn256 "github.com/consensys/gnark/backend/bn256"
	groth16_bn256 "github.com/consensys/gnark/backend/bn256/groth16"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gurvy"
)

func main() {
	circuitPath := filepath.Clean("../circuit.r1cs")
	circuitName := filepath.Base(circuitPath)
	vkPath := filepath.Join("../.", circuitName+".vk")
	pkPath := filepath.Join("../.", circuitName+".pk")

	var bigIntR1cs frontend.R1CS
	if err := gob.Read(circuitPath, &bigIntR1cs, gurvy.BN256); err != nil {
		fmt.Println("error:", err)
		os.Exit(-1)
	}
	r1cs := backend_bn256.Cast(&bigIntR1cs)
	fmt.Printf("%-30s %-30s %-d constraints\n", "loaded circuit", circuitPath, r1cs.NbConstraints)
	// run setup
	var pk groth16_bn256.ProvingKey
	var vk groth16_bn256.VerifyingKey
	start := time.Now()
	groth16_bn256.Setup(&r1cs, &pk, &vk)
	duration := time.Since(start)
	fmt.Printf("%-30s %-30s %-30s\n", "setup completed", "", duration)

	if err := gob.Write(vkPath, &vk, gurvy.BN256); err != nil {
		fmt.Println("error:", err)
		os.Exit(-1)
	}
	fmt.Printf("%-30s %s\n", "generated verifying key", vkPath)
	if err := gob.Write(pkPath, &pk, gurvy.BN256); err != nil {
		fmt.Println("error:", err)
		os.Exit(-1)
	}
	fmt.Printf("%-30s %s\n", "generated proving key", pkPath)
}
