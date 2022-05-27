package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/deroproject/derohe/rpc"
	"github.com/gorilla/websocket"
)

var connection *websocket.Conn
var Blocks uint64
var Minis uint64
var Rejected uint64

// proxy-client
func Start_client(v string, w string) {
	var err error

	rand.Seed(time.Now().UnixMilli())

	for {

		u := url.URL{Scheme: "wss", Host: v, Path: "/ws/" + w}

		dialer := websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		fmt.Println(time.Now().Format(time.Stamp), "Connected to node", v)
		connection, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(5 * time.Second)
			fmt.Println(err)
			continue
		}

		var params rpc.GetBlockTemplate_Result

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

			go SendTemplateToNodes(recv_data)
		}
	}
}

func SendToDaemon(buffer []byte) {
	connection.WriteMessage(websocket.TextMessage, buffer)
}
