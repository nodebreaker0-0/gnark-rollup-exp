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

var (
	fCount uint
)

func main() {

	nbAccounts := 16 // 16 accounts so we know that the proof length is 5

	operator, users := CreateOperator(nbAccounts)

	// read accounts involved in the transfer
	sender, err := operator.readAccount(0)

	receiver, err := operator.readAccount(1)

	// create the transfer and sign it
	var amount uint64
	amount = 10
	transfer := NewTransfer(amount, sender.pubKey, receiver.pubKey, sender.nonce)

	// sign the transfer
	_, err = transfer.Sign(users[0], operator.h)

	// update the state from the received transfer
	err = operator.updateState(transfer, 0)

	fCount = 1
	fProofPath := filepath.Clean("../")
	fPkPath := filepath.Clean("../circuit.pk")
	circuitPath := filepath.Clean("../circuit.r1cs")
	circuitName := filepath.Base(circuitPath)
	circuitExt := filepath.Ext(circuitName)
	circuitName = circuitName[0 : len(circuitName)-len(circuitExt)]

	var bigIntR1cs frontend.R1CS
	if err := gob.Read(circuitPath, &bigIntR1cs, gurvy.BN256); err != nil {
		fmt.Println("error:", err)
		os.Exit(-1)
	}
	r1cs := backend_bn256.Cast(&bigIntR1cs)
	fmt.Printf("%-30s %-30s %-d constraints\n", "loaded circuit", circuitPath, r1cs.NbConstraints)
	// run setup
	var pk groth16_bn256.ProvingKey
	if err := gob.Read(fPkPath, &pk, gurvy.BN256); err != nil {
		fmt.Println("can't load proving key")
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Printf("%-30s %-30s\n", "loaded proving key", fPkPath)

	// parse input file
	// operator.witnesses impory
	//r1csInput := backend.NewAssignment()
	//err = r1csInput.ReadFile(fInputPath)
	if err != nil {
		fmt.Println("can't parse input", err)
		os.Exit(-1)
	}
	//fmt.Printf("%-30s %-30s %-d inputs\n", "loaded input", operator.witnesses, len(operator.witnesses))

	// compute proof
	start := time.Now()
	proof, err := groth16_bn256.Prove(&r1cs, &pk, operator.witnesses)
	if err != nil {
		fmt.Println("Error proof generation", err)
		os.Exit(-1)
	}
	for i := uint(1); i < fCount; i++ {
		_, _ = groth16_bn256.Prove(&r1cs, &pk, operator.witnesses)
	}
	duration := time.Since(start)
	if fCount > 1 {
		duration = time.Duration(int64(duration) / int64(fCount))
	}
	// default proof path
	proofPath := filepath.Join(".", circuitName+".proof")
	if fProofPath != "" {
		proofPath = fProofPath
	}

	if err := gob.Write(proofPath, proof, gurvy.BN256); err != nil {
		fmt.Println("error:", err)
		os.Exit(-1)
	}

	fmt.Printf("%-30s %-30s %-30s\n", "generated proof", proofPath, duration)

}
