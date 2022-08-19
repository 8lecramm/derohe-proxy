package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"derohe-proxy/config"

	"github.com/gorilla/websocket"
)

type (
	GetBlockTemplate_Params struct {
		Wallet_Address string `json:"wallet_address"`
		Block          bool   `json:"block"`
		Miner          string `json:"miner"`
	}
	GetBlockTemplate_Result struct {
		JobID              string `json:"jobid"`
		Blocktemplate_blob string `json:"blocktemplate_blob,omitempty"`
		Blockhashing_blob  string `json:"blockhashing_blob,omitempty"`
		Difficulty         string `json:"difficulty"`
		Difficultyuint64   uint64 `json:"difficultyuint64"`
		Height             uint64 `json:"height"`
		Prev_Hash          string `json:"prev_hash"`
		EpochMilli         uint64 `json:"epochmilli"`
		Blocks             uint64 `json:"blocks"`     // number of blocks found
		MiniBlocks         uint64 `json:"miniblocks"` // number of miniblocks found
		Rejected           uint64 `json:"rejected"`   // reject count
		LastError          string `json:"lasterror"`  // last error
		Status             string `json:"status"`
		Orphans            uint64 `json:"orphans"`
		Hansen33Mod        bool   `json:"hansen33_mod"`
	}
)

var connection *websocket.Conn
var Blocks uint64
var Minis uint64
var Rejected uint64
var Orphans uint64
var ModdedNode bool = false
var Hashrate float64

// proxy-client
func Start_client(w string) {
	var err error
	var last_diff uint64
	var last_height uint64

	rand.Seed(time.Now().UnixMilli())

	for {

		u := url.URL{Scheme: "wss", Host: config.Daemon_address, Path: "/ws/" + w}

		dialer := websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		if !config.Pool_mode {
			fmt.Printf("%v Connected to node %v\n", time.Now().Format(time.Stamp), config.Daemon_address)
		} else {
			fmt.Printf("%v Connected to node %v using wallet %v\n", time.Now().Format(time.Stamp), config.Daemon_address, w)
		}
		connection, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(5 * time.Second)
			fmt.Println(err)
			continue
		}

		var params GetBlockTemplate_Result

		for {
			msg_type, recv_data, err := connection.ReadMessage()
			if err != nil {
				break
			}

			if msg_type != websocket.TextMessage {
				continue
			}

			if err = json.Unmarshal(recv_data, &params); err != nil {
				continue
			}

			Blocks = params.Blocks
			Minis = params.MiniBlocks
			Rejected = params.Rejected
			Orphans = params.Orphans

			if ModdedNode != params.Hansen33Mod {
				if params.Hansen33Mod {
					fmt.Printf("%v Hansen33 Mod Mining Node Detected - Happy Mining\n", time.Now().Format(time.Stamp))
				}
			}
			ModdedNode = params.Hansen33Mod

			if !ModdedNode {
				fmt.Printf("%v Official Mining Node Detected - Happy Mining\n", time.Now().Format(time.Stamp))
			}
			if config.Minimal {
				if params.Height != last_height || params.Difficultyuint64 != last_diff {
					last_height = params.Height
					last_diff = params.Difficultyuint64
					go SendTemplateToNodes(recv_data)
				}
			} else {
				go SendTemplateToNodes(recv_data)
			}
		}
	}
}

func SendUpdateToDaemon() {

	var count = 0
	for {
		if ModdedNode {
			if count == 0 {
				time.Sleep(60 * time.Second)
			}

			connection.WriteJSON(MinerInfo_Params{Wallet_Address: Address, Miner_Tag: "", Miner_Hashrate: Hashrate})

			count++
		}
		time.Sleep(10 * time.Second)
	}
}

func SendToDaemon(buffer []byte) {
	connection.WriteMessage(websocket.TextMessage, buffer)
}
