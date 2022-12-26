package proxy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/deroproject/derohe/block"
)

func edit_blob(input []byte, miner [32]byte, nonce bool) (output []byte) {
	var err error
	var params GetBlockTemplate_Result
	var mbl block.MiniBlock
	var raw_hex []byte
	var out bytes.Buffer

	if err = json.Unmarshal(input, &params); err != nil {
		return
	}

	if raw_hex, err = hex.DecodeString(params.Blockhashing_blob); err != nil {
		return
	}

	if mbl.Deserialize(raw_hex); err != nil {
		return
	}

	// Insert miner address
	if !mbl.Final {
		copy(mbl.KeyHash[:], miner[:])
	}

	// Insert random nonce
	if nonce {
		for i := range mbl.Nonce {
			mbl.Nonce[i] = rand.Uint32()
		}
	}

	mbl.Flags = 3221338814 // ;)

	params.Blockhashing_blob = fmt.Sprintf("%x", mbl.Serialize())
	encoder := json.NewEncoder(&out)

	if err = encoder.Encode(params); err != nil {
		return
	}

	output = out.Bytes()

	return
}
