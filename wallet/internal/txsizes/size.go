// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txsizes

import (
	"github.com/HcashOrg/hcashd/wire"

	h "github.com/HcashOrg/hcashwallet/internal/helpers"
	"github.com/HcashOrg/hcashwallet/wallet/udb"
)

// Worst case script and input/output size estimates.
const (
	// RedeemP2PKHSigScriptSize is the worst case (largest) serialize size
	// of a transaction input script that redeems a compressed P2PKH output.
	// It is calculated as:
	//
	//   - OP_DATA_73
	//   - 72 bytes DER signature + 1 byte sighash
	//   - OP_DATA_33
	//   - 33 bytes serialized compressed pubkey
	RedeemP2PKHSigScriptSize = 1 + 73 + 1 + 33

	RedeemP2PKHBlissSigScriptSize = 3 + 751 + 3 + 897 + 1

	//TODO specify the size of signature
	RedeemP2PKHLmsSigScriptSize = 3 + 1230 + 3 + 999 + 1

	// P2PKHPkScriptSize is the size of a transaction output script that
	// pays to a compressed pubkey hash.  It is calculated as:
	//
	//   - OP_DUP
	//   - OP_HASH160
	//   - OP_DATA_20
	//   - 20 bytes pubkey hash
	//   - OP_EQUALVERIFY
	//   - OP_CHECKSIG
	P2PKHPkScriptSize = 1 + 1 + 1 + 20 + 1 + 1

	P2PKHAltScriptSize = 1 + 1 + 1 + 20 + 1 + 1 + 1

	// RedeemP2PKHInputSize is the worst case (largest) serialize size of a
	// transaction input redeeming a compressed P2PKH output.  It is
	// calculated as:
	//
	//   - 32 bytes previous tx
	//   - 4 bytes output index
	//   - 1 byte tree
	//   - 8 bytes amount
	//   - 4 bytes block height
	//   - 4 bytes block index
	//   - 1 byte compact int encoding value 107
	//   - 107 bytes signature script
	//   - 4 bytes sequence
	RedeemP2PKHInputSize = 32 + 4 + 1 + 8 + 4 + 4 + 1 + RedeemP2PKHSigScriptSize + 4

	//RedeemP2PKHAltInputSize = 32 + 4 + 1 + 8 + 4 + 4 + 2 + RedeemP2PKHAltSigScriptSize + 4
	RedeemP2PKHBlissInputSize = 32 + 4 + 1 + 8 + 4 + 4 + 3 + RedeemP2PKHBlissSigScriptSize + 4

	RedeemP2PKHLmsInputSize = 32 + 4 + 1 + 8 + 4 + 4 + 3 + RedeemP2PKHLmsSigScriptSize + 4
	
	// P2PKHOutputSize is the serialize size of a transaction output with a
	// P2PKH output script.  It is calculated as:
	//
	//   - 8 bytes output value
	//   - 2 bytes version
	//   - 1 byte compact int encoding value 25
	//   - 25 bytes P2PKH output script
	P2PKHOutputSize = 8 + 2 + 1 + P2PKHPkScriptSize

	P2PKHAltOutputSize = 8 + 2 + 1 + P2PKHAltScriptSize
)

// EstimateSerializeSize returns a worst case serialize size estimate for a
// signed transaction that spends inputCount number of compressed P2PKH outputs
// and contains each transaction output from txOuts.  The estimated size is
// incremented for an additional P2PKH change output if addChangeOutput is true.
func EstimateSerializeSize(inputCount int, txOuts []*wire.TxOut, addChangeOutput bool, accType uint8) int {
	changeSize := 0
	inputSize := 0
	outputCount := len(txOuts)
	if addChangeOutput {
		if accType == udb.AcctypeEc {
			changeSize = P2PKHOutputSize
		} else if accType == udb.AcctypeBliss || accType == udb.AcctypeLms{
			changeSize = P2PKHAltOutputSize
		}
		outputCount++
	}
	if accType == udb.AcctypeEc {
		inputSize = RedeemP2PKHInputSize
	}else if accType == udb.AcctypeBliss {
		inputSize = RedeemP2PKHBlissInputSize
	}else if accType == udb.AcctypeLms {
		inputSize = RedeemP2PKHLmsInputSize
	}


	// 12 additional bytes are for version, locktime and expiry.
	return 12 + (2 * wire.VarIntSerializeSize(uint64(inputCount))) +
		wire.VarIntSerializeSize(uint64(outputCount)) +
		inputCount*inputSize +
		h.SumOutputSerializeSizes(txOuts) +
		changeSize
}
