# derohe-proxy

Proxy to combine miners and to reduce network load.
Long To-Do list, but this is a working release.

**Features**
- random nonces
- multiple wallet support
- notification of incoming and lost connections / submitted results / stats
- user-defined logging interval
- pool mining support (no stratum)
- worker support (wallet_address.worker_name)

**Usage**

```derohe-proxy [--listen-address=<127.0.0.1:11111>] [--log-interval=<60>] [--nonce] [--pool] --daemon-address=<1.2.3.4:10100>```

```--listen-address (optional): bind to address:port for incoming miner connections. By default, proxy listens on 0.0.0.0:10200
--daemon-address: address:port of daemon
--log-interval (optional): logging every X seconds, where X >= 60. Default is 60 seconds
--nonce (optional): enable random nonces, disabled by default```
--pool (optional): enable pool mining, disable keyhash replacement
--wallet-address=<dero1....>   use this wallet address for all connections