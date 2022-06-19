package main

import (
	"derohe-proxy/proxy"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/docopt/docopt-go"
)

func main() {
	var err error

	Arguments, err = docopt.Parse(command_line, nil, true, "pre-alpha", false)

	if err != nil {
		return
	}

	if Arguments["--listen-address"] != nil {
		addr, err := net.ResolveTCPAddr("tcp", Arguments["--listen-address"].(string))
		if err != nil {
			return
		} else {
			if addr.Port == 0 {
				return
			} else {
				listen_addr = addr.String()
			}
		}
	}

	if Arguments["--daemon-address"] == nil {
		return
	} else {
		daemon_address = Arguments["--daemon-address"].(string)
	}

	if Arguments["--log-interval"] != nil {
		interval, err := strconv.ParseInt(Arguments["--log-interval"].(string), 10, 32)
		if err != nil {
			return
		} else {
			if interval < 60 || interval > 3600 {
				log_intervall = 60
			} else {
				log_intervall = int(interval)
			}
		}
	}

	if Arguments["--nonce"].(bool) {
		nonce = true
		minimal = true
		fmt.Printf("%v Nonce editing is enabled\n", time.Now().Format(time.Stamp))
		fmt.Printf("%v Switch to >minimal< mode\n", time.Now().Format(time.Stamp))
	}

	if Arguments["--minimal"].(bool) && !Arguments["--nonce"].(bool) {
		minimal = true
		fmt.Printf("%v Forward only 2 jobs per block\n", time.Now().Format(time.Stamp))
	}

	fmt.Printf("%v Logging every %d seconds\n", time.Now().Format(time.Stamp), log_intervall)

	go proxy.Start_server(listen_addr)

	// Wait for first miner connection to grab wallet address
	for proxy.CountMiners() < 1 {
		time.Sleep(time.Second * 1)
	}
	go proxy.Start_client(daemon_address, proxy.Address, minimal, nonce)

	for {
		time.Sleep(time.Second * time.Duration(log_intervall))
		fmt.Printf("%v %d miners connected, Bl: %d, Mbl: %d, Rej: %d\n", time.Now().Format(time.Stamp), proxy.CountMiners(), proxy.Blocks, proxy.Minis, proxy.Rejected)
		for i := range proxy.Wallet_count {
			if proxy.Wallet_count[i] > 1 {
				fmt.Printf("%v Wallet %v, %d miners\n", time.Now().Format(time.Stamp), i, proxy.Wallet_count[i])
			}
		}
	}
}
