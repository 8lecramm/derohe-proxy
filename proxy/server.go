package proxy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"derohe-proxy/config"

	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/graviton"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var server *nbhttp.Server

var memPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 16*1024)
	},
}

type user_session struct {
	blocks        uint64
	miniblocks    uint64
	lasterr       string
	address       rpc.Address
	worker        string
	orphans       uint64
	hashrate      float64
	valid_address bool
	address_sum   [32]byte
}

type ( // array without name containing block template in hex
	MinerInfo_Params struct {
		Wallet_Address string  `json:"wallet_address"`
		Miner_Tag      string  `json:"miner_tag"`
		Miner_Hashrate float64 `json:"miner_hashrate"`
	}
	MinerInfo_Result struct {
	}
)

var client_list_mutex sync.Mutex
var client_list = map[*websocket.Conn]*user_session{}

var miners_count int
var Shares uint64
var shareValue uint64
var Wallet_count map[string]uint
var Address string

var rwmutex sync.RWMutex

func Start_server() {
	var err error

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{generate_random_tls_cert()},
		InsecureSkipVerify: true,
	}

	mux := &http.ServeMux{}
	mux.HandleFunc("/", onWebsocket) // handle everything

	server = nbhttp.NewServer(nbhttp.Config{
		Name:                    "GETWORK",
		Network:                 "tcp",
		AddrsTLS:                []string{config.Listen_addr},
		TLSConfig:               tlsConfig,
		Handler:                 mux,
		MaxLoad:                 10 * 1024,
		MaxWriteBufferSize:      5 * 1024 * 1024,
		ReleaseWebsocketPayload: true,
		KeepaliveTime:           240 * time.Hour, // we expects all miners to find a block every 10 days,
		NPoller:                 runtime.NumCPU(),
	})

	server.OnReadBufferAlloc(func(c *nbio.Conn) []byte {
		return memPool.Get().([]byte)
	})
	server.OnReadBufferFree(func(c *nbio.Conn, b []byte) {
		memPool.Put(b)
	})

	if err = server.Start(); err != nil {
		return
	}

	Wallet_count = make(map[string]uint)

	server.Wait()
	defer server.Stop()

}

func CountMiners() int {
	client_list_mutex.Lock()
	defer client_list_mutex.Unlock()

	miners_count = len(client_list)

	return miners_count
}

// forward all incoming templates from daemon to all miners
func SendTemplateToNodes(data []byte) {

	client_list_mutex.Lock()
	defer client_list_mutex.Unlock()

	for rk, rv := range client_list {

		if client_list == nil {
			break
		}

		if !config.Pool_mode {
			miner_address := rv.address_sum

			if result := edit_blob(data, miner_address, config.Nonce); result != nil {
				data = result
			} else {
				fmt.Println(time.Now().Format(time.Stamp), "Failed to change nonce / miner keyhash")
			}
		}

		go func(k *websocket.Conn, v *user_session) {
			defer globals.Recover(2)
			k.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
			k.WriteMessage(websocket.TextMessage, data)

		}(rk, rv)

	}
}

// handling for incoming miner connections
func onWebsocket(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/ws/") {
		http.NotFound(w, r)
		return
	}
	address := strings.TrimPrefix(r.URL.Path, "/ws/")

	// check for worker suffix
	var parseWorker []string
	var worker string
	if strings.Contains(address, ".") {
		parseWorker = strings.Split(address, ".")
		worker = parseWorker[1]
		address = parseWorker[0]
	}

	addr, err := globals.ParseValidateAddress(address)
	if err != nil {
		fmt.Fprintf(w, "err: %s\n", err)
		return
	}

	upgrader := newUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//panic(err)
		return
	}

	addr_raw := addr.PublicKey.EncodeCompressed()
	wsConn := conn.(*websocket.Conn)

	session := user_session{address: *addr, address_sum: graviton.Sum(addr_raw), worker: worker}
	wsConn.SetSession(&session)

	client_list_mutex.Lock()
	defer client_list_mutex.Unlock()

	client_list[wsConn] = &session
	Wallet_count[client_list[wsConn].address.String()]++

	if config.WalletAddr != "" {
		Address = config.WalletAddr
	} else {
		Address = address
	}

	if !config.Pool_mode {
		fmt.Printf("%v Incoming connection: %v (%v), Wallet: %v\n", time.Now().Format(time.Stamp), wsConn.RemoteAddr().String(), worker, address)
	} else {
		fmt.Printf("%v Incoming connection: %v (%v)\n", time.Now().Format(time.Stamp), wsConn.RemoteAddr().String(), worker)
		if len(client_list) == 1 {
			Connected = time.Now().UnixMilli()
			shareValue = 0
		}
	}
}

// forward results to daemon
func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()

	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {

		if messageType != websocket.TextMessage {
			return
		}

		client_list_mutex.Lock()
		defer client_list_mutex.Unlock()

		SendToDaemon(data)
		if !config.Pool_mode {
			fmt.Printf("%v Submitting result from miner: %v (%v), Wallet: %v\n", time.Now().Format(time.Stamp), c.RemoteAddr().String(), client_list[c].worker, client_list[c].address.String())
		} else {
			Shares++
			shareValue += difficulty
			if Connected > 0 {
				Hashrate = shareValue / (uint64(time.Now().UnixMilli()-Connected) / 1000)
			}
		}
		//}
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		client_list_mutex.Lock()
		defer client_list_mutex.Unlock()
		Wallet_count[client_list[c].address.String()]--
		fmt.Printf("%v Lost connection: %v (%v)\n", time.Now().Format(time.Stamp), c.RemoteAddr().String(), client_list[c].worker)
		delete(client_list, c)
	})

	return u
}

func CountWallets() {

	rwmutex.RLock()
	defer rwmutex.RUnlock()

	for i := range Wallet_count {
		if Wallet_count[i] > 1 {
			fmt.Printf("%v Wallet %v, %d miners\n", time.Now().Format(time.Stamp), i, Wallet_count[i])
		}
	}
}

// taken unmodified from derohe repo
// cert handling
func generate_random_tls_cert() tls.Certificate {

	/* RSA can do only 500 exchange per second, we need to be faster
	     * reference https://github.com/golang/go/issues/20058
	    key, err := rsa.GenerateKey(rand.Reader, 512) // current using minimum size
	if err != nil {
	    log.Fatal("Private key cannot be created.", err.Error())
	}

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
	    Type:  "RSA PRIVATE KEY",
	    Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	*/
	// EC256 does roughly 20000 exchanges per second
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		panic(err)
	}
	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	tml := x509.Certificate{
		SerialNumber: big.NewInt(int64(time.Now().UnixNano())),

		// TODO do we need to add more parameters to make our certificate more authentic
		// and thwart traffic identification as a mass scale

		// you can add any attr that you need
		NotBefore: time.Now().AddDate(0, -1, 0),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		// you have to generate a different serial number each execution
		/*
		   Subject: pkix.Name{
		       CommonName:   "New Name",
		       Organization: []string{"New Org."},
		   },
		   BasicConstraintsValid: true,   // even basic constraints are not required
		*/
	}
	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		panic(err)
	}
	return tlsCert
}
