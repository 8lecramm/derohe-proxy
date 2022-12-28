package config

var Command_line string = `derohe-proxy
Proxy to combine all miners and to reduce network load

Usage:
  derohe-proxy --daemon-address=<1.2.3.4:10100> [--listen-address=<127.0.0.1:10100>] [--log-interval=<60>] [--minimal] [--nonce] [--pool] [--wallet-address=<dero1...>]

Options:
 --listen-address=<127.0.0.1:10100>		bind to specific address:port, default is 0.0.0.0:10200
 --daemon-address=<1.2.3.4:10100>		connect to this daemon
 --wallet-address=<dero1....>   use this wallet address for all connections
 --log-interval=<60>   set logging interval in seconds (range 60 - 3600), default is 60 seconds
 --minimal   forward only 2 jobs per block (1 for miniblocks and 1 for final miniblock), by default all jobs are forwarded
 --nonce   enable nonce editing, default is off
 --pool    use this option for pool mining; this option avoids changing the keyhash

Example Mainnet: ./derohe-proxy --daemon-address=minernode1.dero.io:10100
`

// program arguments
var Arguments = map[string]interface{}{}

var Listen_addr string = "0.0.0.0:10200"
var Daemon_address string = "minernode1.dero.io:10100"
var WalletAddr string = ""
var Worker string

// logging interval in seconds
var Log_intervall int = 60

// send only 2 jobs per block
var Minimal bool = false

// edit nonce
var Nonce bool = false

// pool mining
var Pool_mode bool = false
