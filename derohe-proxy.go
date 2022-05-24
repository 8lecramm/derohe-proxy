package main

import (
	"derohe-proxy/proxy"
	"fmt"
	"net"
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

	go proxy.Start_server(listen_addr)

	// Wait for first miner connection to grab wallet address
	for proxy.CountMiners() < 1 {
		time.Sleep(time.Second * 1)
	}
	go proxy.Start_client(daemon_address, proxy.Address)

	for {
		time.Sleep(time.Minute * 5)
		fmt.Printf("%v %4d miners connected\t\tBlocks:%4d\tMiniblocks:%4d\tRejected:%4d\n", time.Now().Format(time.Stamp), proxy.CountMiners(), proxy.Blocks, proxy.Minis, proxy.Rejected)
	}
}
