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
	valid_address bool
	address_sum   [32]byte
}

var client_list_mutex sync.Mutex
var client_list = map[*websocket.Conn]*user_session{}

var miners_count int
var Address string

func Start_server(listen string) {
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
		AddrsTLS:                []string{listen},
		TLSConfig:               tlsConfig,
		Handler:                 mux,
		MaxLoad:                 10 * 1024,
		MaxWriteBufferSize:      32 * 1024,
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
func SendTemplateToNodes(input []byte) {
	var data []byte

	for rk, rv := range client_list {

		if client_list == nil {
			break
		}

		if nonce := edit_blob(input); nonce != nil {
			data = nonce
		} else {
			fmt.Println(time.Now().Format(time.Stamp), "Failed to change nonce")
			data = input
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

	session := user_session{address: *addr, address_sum: graviton.Sum(addr_raw)}
	wsConn.SetSession(&session)

	client_list_mutex.Lock()
	defer client_list_mutex.Unlock()
	client_list[wsConn] = &session
	Address = address
	fmt.Println(time.Now().Format(time.Stamp), "Incoming connection from IP:", wsConn.RemoteAddr().String())
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
		fmt.Println(time.Now().Format(time.Stamp), "Submitting result from miner IP:", c.RemoteAddr().String())
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		client_list_mutex.Lock()
		defer client_list_mutex.Unlock()
		delete(client_list, c)
		fmt.Println(time.Now().Format(time.Stamp), "Interrupted or lost connection:", c.RemoteAddr().String())
	})

	return u
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
