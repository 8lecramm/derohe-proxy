# derohe-proxy

Proxy to combine miners and to reduce network load
The proxy waits for first incoming connection and uses this wallet address for mining.
Long To-Do list, but this release is working.

**Features**
- random nonces
- notification of incoming / lost connections
- statistics every 5 minutes

**Usage**

derohe-proxy [--listen-address=<127.0.0.1:10100>] --daemon-address=<1.2.3.4:10100>

--listen-address (optional): bind to address:port for incoming miner connections. By default, proxy listens on 0.0.0.0:10200

--daemon-address: address:port of daemon
