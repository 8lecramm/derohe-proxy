package main

var command_line string = `derohe-proxy
Proxy to combine all miners and to reduce network load

Usage:
  derohe-proxy [--listen-address=<127.0.0.1:10100>] --daemon-address=<1.2.3.4:10100>

Options:
 --listen-address=<127.0.0.1:10100>		bind to specific address:port, default is 0.0.0.0:10200
 --daemon-address=<1.2.3.4:10100>		connect to this daemon

Example Mainnet: ./derohe-proxy --daemon-address=minernode1.dero.io:10100
`

// program arguments
var Arguments = map[string]interface{}{}

var listen_addr string = "0.0.0.0:10200"
var daemon_address string = "minernode1.dero.io:10100"
