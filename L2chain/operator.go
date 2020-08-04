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
	"bytes"
	"hash"
	"math/big"
	"strconv"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/crypto/accumulator/merkletree"
	"github.com/consensys/gnark/gadgets/accumulator/merkle"
)

var (
	// basename of the inputs for the proofs before update
	baseNameSenderMerkleBefore        = "merkle_sender_proof_before_"
	baseNameSenderProofHelperBefore   = "merkle_sender_proof_helper_before"
	baseNameReceiverMerkleBefore      = "merkle_receiver_proof_before_"
	baseNameReceiverProofHelperBefore = "merkle_receiver_proof_helper_before"
	baseNameRootHashBefore            = "merkle_rh_before_"

	// basename of the inputs for the proofs after update
	baseNameSenderMerkleAfter        = "merkle_sender_proof_after_"
	baseNameSenderProofHelperAfter   = "merkle_sender_proof_helper_after"
	baseNameReceiverMerkleAfter      = "merkle_receiver_proof_after_"
	baseNameReceiverProofHelperAfter = "merkle_receiver_proof_helper_after"
	baseNameRootHashAfter            = "merkle_rh_after_"

	// basename sender account pubkey
	baseNameSenderAccountPubkeyx = "a_sender_pubkeyx_"
	baseNameSenderAccountPubkeyy = "a_sender_pubkeyy_"

	// basename of the sender account input before update
	baseNameSenderAccountIndexBefore   = "a_sender_index_before_"
	baseNameSenderAccountNonceBefore   = "a_sender_nonce_before_"
	baseNameSenderAccountBalanceBefore = "a_sender_balance_before_"

	// basename of the sender account input adter update
	baseNameSenderAccountIndexAfter   = "a_sender_index_before_after_"
	baseNameSenderAccountNonceAfter   = "a_sender_nonce_before_after_"
	baseNameSenderAccountBalanceAfter = "a_sender_balance_before_after_"

	// basename of the receiver account pubk
	baseNameReceiverAccountPubkeyx = "a_receiver_pubkeyx_"
	baseNameReceiverAccountPubkeyy = "a_receiver_pubkeyy_"

	// basename of the receiver account input before update
	baseNameReceiverAccountIndexBefore   = "a_receiver_index_before_"
	baseNameReceiverAccountNonceBefore   = "a_receiver_nonce_before_"
	baseNameReceiverAccountBalanceBefore = "a_receiver_balance_before_"

	// basename of the receiver account input after update
	baseNameReceiverAccountIndexAfter   = "a_receiver_index_after_"
	baseNameReceiverAccountNonceAfter   = "a_receiver_nonce_after_"
	baseNameReceiverAccountBalanceAfter = "a_receiver_balance_after_"

	// basename of the transfer input
	baseNameTransferAmount = "t_sender_amount_"
	baseNameTransferSigRx  = "t_sig_Rx_"
	baseNameTransferSigRy  = "t_sig_Ry_"
	baseNameTransferSigS   = "t_sig_S_"
)

// BatchSize size of a batch of transactions to put in a snark
var BatchSize = 10

// Queue queue for storing the transfers (fixed size queue)
type Queue struct {
	listTransfers chan Transfer
}

// NewQueue creates a new queue, batchSize is the capaciy
func NewQueue(batchSize int) Queue {
	resChan := make(chan Transfer, batchSize)
	var res Queue
	res.listTransfers = resChan
	return res
}

// Push queues up t in the queue
func (q *Queue) Push(t Transfer) {
	q.listTransfers <- t
}

// Pop dequeues an element from the queue
func (q *Queue) Pop() Transfer {
	t := <-q.listTransfers
	return t
}

// Operator represents a rollup operator
type Operator struct {
	State      []byte              // list of accounts: index || nonce || balance || pubkeyX || pubkeyY, each chunk is 256 bits
	HashState  []byte              // Hashed version of the state, each chunk is 256bits: ... || H(index || nonce || balance || pubkeyX || pubkeyY)) || ...
	AccountMap map[string]uint64   // hashmap of all available accounts (the key is the account.pubkey.X), the value is the index of the account in the state
	nbAccounts int                 // number of accounts managed by this operator
	h          hash.Hash           // hash function used to build the Merkle Tree
	q          Queue               // queue of transfers
	batch      int                 // current number of transactions in a batch
	witnesses  backend.Assignments // witnesses for the snark cicruit
}

// NewOperator creates a new operator.
// nbAccounts is the number of accounts managed by this operator, h is the hash function for the merkle proofs
func NewOperator(nbAccounts int, h hash.Hash) Operator {
	res := Operator{}

	// create a list of empty accounts
	res.State = make([]byte, SizeAccount*nbAccounts)

	// initialize hash of the state
	res.HashState = make([]byte, h.Size()*nbAccounts)
	for i := 0; i < nbAccounts; i++ {
		h.Reset()
		h.Write(res.State[i*SizeAccount : i*SizeAccount+SizeAccount])
		s := h.Sum([]byte{})
		copy(res.HashState[i*h.Size():(i+1)*h.Size()], s)
	}

	res.AccountMap = make(map[string]uint64)
	res.nbAccounts = nbAccounts
	res.h = h
	res.q = NewQueue(BatchSize)
	res.batch = 0
	res.witnesses = backend.NewAssignment()
	return res
}

// readAccount reads the account located at index i
func (o *Operator) readAccount(i uint64) (Account, error) {

	var res Account
	err := Deserialize(&res, o.State[int(i)*SizeAccount:int(i)*SizeAccount+SizeAccount])
	if err != nil {
		return res, err
	}
	res.pubKey.HFunc = o.h
	return res, nil
}

// updateAccount updates the state according to transfer
// numTransfer is the number of the transfer currently handled (between 0 and batchSize)
func (o *Operator) updateState(t Transfer, numTransfer int) error {

	var posSender, posReceiver uint64
	var ok bool

	ext := strconv.Itoa(numTransfer)
	segmentSize := o.h.Size()

	// read sender's account
	if posSender, ok = o.AccountMap[string(t.senderPubKey.A.X.Bytes())]; !ok {
		return ErrNonExistingAccount
	}
	senderAccount, err := o.readAccount(posSender)
	if err != nil {
		return err
	}

	// read receiver's account
	if posReceiver, ok = o.AccountMap[string(t.receiverPubKey.A.X.Bytes())]; !ok {
		return ErrNonExistingAccount
	}
	receiverAccount, err := o.readAccount(posReceiver)
	if err != nil {
		return err
	}

	// set witnesses for the public keys
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountPubkeyx+ext, senderAccount.pubKey.A.X)
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountPubkeyy+ext, senderAccount.pubKey.A.Y)
	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountPubkeyx+ext, receiverAccount.pubKey.A.X)
	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountPubkeyy+ext, receiverAccount.pubKey.A.Y)

	// set witnesses for the accounts before update
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountIndexBefore+ext, senderAccount.index)
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountNonceBefore+ext, senderAccount.nonce)
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountBalanceBefore+ext, senderAccount.balance)

	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountIndexBefore+ext, receiverAccount.index)
	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountNonceBefore+ext, receiverAccount.nonce)
	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountBalanceBefore+ext, receiverAccount.balance)

	//  Set witnesses for the proof of inclusion of sender and receivers account before update
	var buf bytes.Buffer
	_, err = buf.Write(o.HashState)
	if err != nil {
		return err
	}
	merkleRootBefore, proofInclusionSenderBefore, numLeaves, err := merkletree.BuildReaderProof(&buf, o.h, segmentSize, posSender)
	if err != nil {
		return err
	}
	merkletree.VerifyProof(o.h, merkleRootBefore, proofInclusionSenderBefore, posSender, numLeaves)
	merkleProofHelperSenderBefore := merkle.GenerateProofHelper(proofInclusionSenderBefore, posSender, numLeaves)

	buf.Reset() // the buffer needs to be reset
	_, err = buf.Write(o.HashState)
	if err != nil {
		return err
	}
	_, proofInclusionReceiverBefore, _, err := merkletree.BuildReaderProof(&buf, o.h, segmentSize, posReceiver)
	if err != nil {
		return err
	}
	merkleProofHelperReceiverBefore := merkle.GenerateProofHelper(proofInclusionReceiverBefore, posReceiver, numLeaves)
	o.witnesses.Assign(backend.Public, baseNameRootHashBefore+ext, merkleRootBefore)
	for i := 0; i < len(proofInclusionSenderBefore); i++ {
		o.witnesses.Assign(backend.Secret, baseNameSenderMerkleBefore+ext+strconv.Itoa(i), proofInclusionSenderBefore[i])
		o.witnesses.Assign(backend.Secret, baseNameReceiverMerkleBefore+ext+strconv.Itoa(i), proofInclusionReceiverBefore[i])

		if i < len(proofInclusionReceiverBefore)-1 {
			o.witnesses.Assign(backend.Secret, baseNameSenderProofHelperBefore+ext+strconv.Itoa(i), merkleProofHelperSenderBefore[i])
			o.witnesses.Assign(backend.Secret, baseNameReceiverProofHelperBefore+ext+strconv.Itoa(i), merkleProofHelperReceiverBefore[i])
		}
	}

	// set witnesses for the transfer
	o.witnesses.Assign(backend.Secret, baseNameTransferAmount+ext, t.amount)
	o.witnesses.Assign(backend.Secret, baseNameTransferSigRx+ext, t.signature.R.X)
	o.witnesses.Assign(backend.Secret, baseNameTransferSigRy+ext, t.signature.R.Y)
	o.witnesses.Assign(backend.Secret, baseNameTransferSigS+ext, t.signature.S)

	// verifying the signature. The msg is the hash (o.h) of the transfer
	// nonce || amount || senderpubKey(x&y) || receiverPubkey(x&y)
	resSig, err := t.Verify(o.h)
	if err != nil {
		return err
	}
	if !resSig {
		return ErrWrongSignature
	}

	// checks if the amount is correct
	var bAmount, bBalance big.Int
	receiverAccount.balance.ToBigIntRegular(&bBalance)
	t.amount.ToBigIntRegular(&bAmount)
	if bAmount.Cmp(&bBalance) == 1 {
		return ErrAmountTooHigh
	}

	// check if the nonce is correct
	if t.nonce != senderAccount.nonce {
		return ErrNonce
	}

	// update the balance of the sender
	senderAccount.balance.Sub(&senderAccount.balance, &t.amount)

	// update the balance of the receiver
	receiverAccount.balance.Add(&receiverAccount.balance, &t.amount)

	// update the nonce of the sender
	senderAccount.nonce++

	// set the witnesses for the account after update
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountIndexAfter+ext, senderAccount.index)
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountNonceAfter+ext, senderAccount.nonce)
	o.witnesses.Assign(backend.Secret, baseNameSenderAccountBalanceAfter+ext, senderAccount.balance)

	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountIndexAfter+ext, receiverAccount.index)
	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountNonceAfter+ext, receiverAccount.nonce)
	o.witnesses.Assign(backend.Secret, baseNameReceiverAccountBalanceAfter+ext, receiverAccount.balance)

	// update the state of the operator
	copy(o.State[int(posSender)*SizeAccount:], senderAccount.Serialize())
	o.h.Reset()
	o.h.Write(senderAccount.Serialize())
	bufSender := o.h.Sum([]byte{})
	copy(o.HashState[int(posSender)*o.h.Size():(int(posSender)+1)*o.h.Size()], bufSender)

	copy(o.State[int(posReceiver)*SizeAccount:], receiverAccount.Serialize())
	o.h.Reset()
	o.h.Write(receiverAccount.Serialize())
	bufReceiver := o.h.Sum([]byte{})
	copy(o.HashState[int(posReceiver)*o.h.Size():(int(posReceiver)+1)*o.h.Size()], bufReceiver)

	//  Set witnesses for the proof of inclusion of sender and receivers account after update
	buf.Reset()
	_, err = buf.Write(o.HashState)
	if err != nil {
		return err
	}
	merkleRootAfer, proofInclusionSenderAfter, _, err := merkletree.BuildReaderProof(&buf, o.h, segmentSize, posSender)
	if err != nil {
		return err
	}
	merkleProofHelperSenderAfter := merkle.GenerateProofHelper(proofInclusionSenderAfter, posSender, numLeaves)

	buf.Reset() // the buffer needs to be reset
	_, err = buf.Write(o.HashState)
	if err != nil {
		return err
	}
	_, proofInclusionReceiverAfter, _, err := merkletree.BuildReaderProof(&buf, o.h, segmentSize, posReceiver)
	if err != nil {
		return err
	}
	merkleProofHelperReceiverAfter := merkle.GenerateProofHelper(proofInclusionReceiverAfter, posReceiver, numLeaves)

	o.witnesses.Assign(backend.Public, baseNameRootHashAfter+ext, merkleRootAfer)
	for i := 0; i < len(proofInclusionSenderAfter); i++ {
		o.witnesses.Assign(backend.Secret, baseNameSenderMerkleAfter+ext+strconv.Itoa(i), proofInclusionSenderAfter[i])
		o.witnesses.Assign(backend.Secret, baseNameReceiverMerkleAfter+ext+strconv.Itoa(i), proofInclusionReceiverAfter[i])

		if i < len(proofInclusionSenderAfter)-1 {
			o.witnesses.Assign(backend.Secret, baseNameSenderProofHelperAfter+ext+strconv.Itoa(i), merkleProofHelperSenderAfter[i])
			o.witnesses.Assign(backend.Secret, baseNameReceiverProofHelperAfter+ext+strconv.Itoa(i), merkleProofHelperReceiverAfter[i])
		}
	}

	return nil
}
