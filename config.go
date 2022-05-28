package main

var command_line string = `derohe-proxy
Proxy to combine all miners and to reduce network load

Usage:
  derohe-proxy [--listen-address=<127.0.0.1:10100>] [--log-interval=<60>] [--minimal-jobs] --daemon-address=<1.2.3.4:10100>

Options:
 --listen-address=<127.0.0.1:10100>		bind to specific address:port, default is 0.0.0.0:10200
 --daemon-address=<1.2.3.4:10100>		connect to this daemon
 --log-interval=<60>   set logging interval in seconds (range 60 - 3600), default is 60 seconds
 --minimal-jobs   forward only 2 jobs per block (1 for miniblocks and 1 for final miniblock), by default all jobs are forwarded

Example Mainnet: ./derohe-proxy --daemon-address=minernode1.dero.io:10100
`

// program arguments
var Arguments = map[string]interface{}{}

var listen_addr string = "0.0.0.0:10200"
var daemon_address string = "minernode1.dero.io:10100"

// logging interval in seconds
var log_intervall int = 60

// send only 2 jobs per block
var minimal = false
