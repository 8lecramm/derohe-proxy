# derohe-proxy

Proxy to combine miners and to reduce network load.
Long To-Do list, but this is a working release.

**Features**
- random nonces
- muliple wallets are supported
- notification of incoming / lost connections
- user-defined logging interval

**Usage**

derohe-proxy [--listen-address=<127.0.0.1:10100>] [--log-interval=<60>] --daemon-address=<1.2.3.4:10100>

--listen-address (optional): bind to address:port for incoming miner connections. By default, proxy listens on 0.0.0.0:10200

--daemon-address: address:port of daemon

--log-interval (optional): logging every X seconds, default is 60 seconds
