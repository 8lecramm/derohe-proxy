package proxy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bitfield/qrand"
	"github.com/deroproject/derohe/block"
	"github.com/deroproject/derohe/rpc"
)

type nonce_store struct {
	nonce     [8]byte
	timestamp time.Time
	sync.RWMutex
}

func edit_blob(input []byte, miner [32]byte, nonce bool) (output []byte) {
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

	// Insert miner address
	if !mbl.Final {
		copy(mbl.KeyHash[:], miner[:])
	}

	// Insert random nonce
	if nonce {
		var qnonce [8]byte
		Found.Lock()
		// send nonce pattern to all nodes, lasts for 2 hours or until another nonce has been found
		// TODO: add command argument
		if binary.BigEndian.Uint64(Found.nonce[:]) > 0 && time.Now().Sub(Found.timestamp) < time.Hour*2 {
			copy(qnonce[:], Found.nonce[:])
			qrand.Read(qnonce[0:2])
			mbl.Nonce[1] = binary.BigEndian.Uint32(qnonce[0:4])
			mbl.Nonce[2] = binary.BigEndian.Uint32(qnonce[4:8])
		} else {
			qrand.Read(qnonce[:])
			mbl.Nonce[1] = binary.BigEndian.Uint32(qnonce[0:4])
			mbl.Nonce[2] = binary.BigEndian.Uint32(qnonce[4:8])
		}
		Found.Unlock()
	}

	params.Blockhashing_blob = fmt.Sprintf("%x", mbl.Serialize())
	encoder := json.NewEncoder(&out)

	if err = encoder.Encode(params); err != nil {
		return
	}

	output = out.Bytes()

	return
}
