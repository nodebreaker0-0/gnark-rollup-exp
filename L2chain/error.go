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

import "errors"

var (
	// ErrSizeByteSlice memory checking
	ErrSizeByteSlice = errors.New("Byte slice size is inconsistant with Account size")

	// ErrNonExistingAccount account not in the database
	ErrNonExistingAccount = errors.New("The account is not in the rollup database")

	// ErrNonConsistantAccount account not in the database
	ErrNonConsistantAccount = errors.New("The account provided exists but is inconsistant with what's in the database")

	// ErrWrongSignature wrong signature
	ErrWrongSignature = errors.New("Invalid signature")

	// ErrAmountTooHigh the amount is bigger than the balance
	ErrAmountTooHigh = errors.New("Amount is bigger than balance")

	// ErrNonce inconsistant nonce between transfer and account
	ErrNonce = errors.New("Incorrect nonce")
)
