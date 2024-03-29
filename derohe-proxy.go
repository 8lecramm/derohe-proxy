package main

import (
	"derohe-proxy/config"
	"derohe-proxy/proxy"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/deroproject/derohe/globals"
	"github.com/docopt/docopt-go"
)

func main() {
	var err error

	config.Arguments, err = docopt.Parse(config.Command_line, nil, true, "pre-alpha", false)

	if err != nil {
		return
	}

	if config.Arguments["--listen-address"] != nil {
		addr, err := net.ResolveTCPAddr("tcp", config.Arguments["--listen-address"].(string))
		if err != nil {
			return
		} else {
			if addr.Port == 0 {
				return
			} else {
				config.Listen_addr = addr.String()
			}
		}
	}

	if config.Arguments["--daemon-address"] == nil {
		return
	} else {
		config.Daemon_address = config.Arguments["--daemon-address"].(string)
	}

	if config.Arguments["--wallet-address"] != nil {

		// check for worker suffix
		var parseWorker []string
		var address string

		if strings.Contains(config.Arguments["--wallet-address"].(string), ".") {
			parseWorker = strings.Split(config.Arguments["--wallet-address"].(string), ".")
			config.Worker = parseWorker[1]
			address = parseWorker[0]
		} else {
			address = config.Arguments["--wallet-address"].(string)
		}

		addr, err := globals.ParseValidateAddress(address)
		if err != nil {
			fmt.Printf("%v Wallet address is invalid!\n", time.Now().Format(time.Stamp))
		}
		config.WalletAddr = addr.String()
		if config.Worker != "" {
			fmt.Printf("%v Using wallet %s and name %s for all connections\n", time.Now().Format(time.Stamp), config.WalletAddr, config.Worker)
		} else {
			fmt.Printf("%v Using wallet %s for all connections\n", time.Now().Format(time.Stamp), config.WalletAddr)
		}
	}

	if config.Arguments["--log-interval"] != nil {
		interval, err := strconv.ParseInt(config.Arguments["--log-interval"].(string), 10, 32)
		if err != nil {
			return
		} else {
			if interval < 60 || interval > 3600 {
				config.Log_intervall = 60
			} else {
				config.Log_intervall = int(interval)
			}
		}
	}

	/*if config.Arguments["--minimal"].(bool) {
		config.Minimal = true
		fmt.Printf("%v Forward only 2 jobs per block\n", time.Now().Format(time.Stamp))
	}*/

	if config.Arguments["--nonce"].(bool) {
		config.Nonce = true
		fmt.Printf("%v Nonce editing is enabled\n", time.Now().Format(time.Stamp))
	}

	if config.Arguments["--pool"].(bool) {
		config.Pool_mode = true
		config.Minimal = false
		fmt.Printf("%v Pool mode is enabled\n", time.Now().Format(time.Stamp))
	}

	fmt.Printf("%v Logging every %d seconds\n", time.Now().Format(time.Stamp), config.Log_intervall)

	go proxy.Start_server()

	// Wait for first miner connection to grab wallet address
	for proxy.CountMiners() < 1 {
		time.Sleep(time.Second * 1)
	}
	if config.Worker == "" {
		go proxy.Start_client(proxy.Address)
	} else {
		go proxy.Start_client(proxy.Address + "." + config.Worker)
	}

	for {
		time.Sleep(time.Second * time.Duration(config.Log_intervall))

		miners := proxy.CountMiners()

		hash_rate_string := ""

		if miners > 0 {
			switch {
			case proxy.Hashrate > 1000000000000:
				hash_rate_string = fmt.Sprintf("%.3f TH/s", float64(proxy.Hashrate)/1000000000000.0)
			case proxy.Hashrate > 1000000000:
				hash_rate_string = fmt.Sprintf("%.3f GH/s", float64(proxy.Hashrate)/1000000000.0)
			case proxy.Hashrate > 1000000:
				hash_rate_string = fmt.Sprintf("%.3f MH/s", float64(proxy.Hashrate)/1000000.0)
			case proxy.Hashrate > 1000:
				hash_rate_string = fmt.Sprintf("%.3f KH/s", float64(proxy.Hashrate)/1000.0)
			case proxy.Hashrate > 0:
				hash_rate_string = fmt.Sprintf("%d H/s", int(proxy.Hashrate))
			}
		}

		if !config.Pool_mode {
			fmt.Printf("%v %d miners connected, IB:%d MB:%d MBR:%d MBO:%d\n", time.Now().Format(time.Stamp), miners, proxy.Blocks, proxy.Minis, proxy.Rejected, proxy.Orphans)
		} else {
			fmt.Printf("%v %d miners connected, Pool stats: IB:%d MB:%d MBR:%d MBO:%d\n", time.Now().Format(time.Stamp), miners, proxy.Blocks, proxy.Minis, proxy.Rejected, proxy.Orphans)
			fmt.Printf("%v Shares submitted: %d, Hashrate: %s\n", time.Now().Format(time.Stamp), proxy.Shares, hash_rate_string)
		}
		proxy.CountWallets()
	}
}
