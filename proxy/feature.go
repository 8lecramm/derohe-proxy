package proxy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/deroproject/derohe/block"
	"github.com/deroproject/derohe/rpc"
)

func edit_blob(input []byte) (output []byte) {
	var err error
	var params rpc.GetBlockTemplate_Result
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

	for i := range mbl.Nonce {
		mbl.Nonce[i] = rand.Uint32()
	}
	mbl.Flags = rand.Uint32()

	params.Blockhashing_blob = fmt.Sprintf("%x", mbl.Serialize())
	encoder := json.NewEncoder(&out)

	if err = encoder.Encode(params); err != nil {
		return
	}

	output = out.Bytes()

	return
}
