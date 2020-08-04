/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/consensys/gnark/backend"
	groth16_bn256 "github.com/consensys/gnark/backend/bn256/groth16"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gurvy"
)

// verifyCmd represents the verify command
func main() {
	fVkPath := filepath.Clean("../circuit.vk")
	proofPath := filepath.Clean("../circuit.proof")
	var vk groth16_bn256.VerifyingKey
	if err := gob.Read(fVkPath, &vk, gurvy.BN256); err != nil {
		fmt.Println("can't load verifying key")
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Printf("%-30s %-30s\n", "loaded verifying key", fVkPath)

	// parse input file
	r1csInput := backend.NewAssignment()
	err := r1csInput.ReadFile(fInputPath)
	if err != nil {
		fmt.Println("can't parse input", err)
		os.Exit(-1)
	}
	fmt.Printf("%-30s %-30s %-d inputs\n", "loaded input", fInputPath, len(r1csInput))

	// load proof
	var proof groth16_bn256.Proof
	if err := gob.Read(proofPath, &proof, gurvy.BN256); err != nil {
		fmt.Println("can't parse proof", err)
		os.Exit(-1)
	}

	// verify proof
	start := time.Now()
	result, err := groth16_bn256.Verify(&proof, &vk, r1csInput)
	if err != nil || !result {
		fmt.Printf("%-30s %-30s %-30s\n", "proof is invalid", proofPath, time.Since(start))
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(-1)
	}
}
