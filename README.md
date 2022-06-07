# derohe-proxy

Proxy to combine miners and to reduce network load.
Long To-Do list, but this is a working release.

**Features**
- random nonces
- muliple wallets are supported
- notification of incoming and lost connections / submitted results / stats
- user-defined logging interval

**Usage**

```derohe-proxy [--listen-address=<127.0.0.1:11111>] [--log-interval=<60>] [--minimal] [--nonce] --daemon-address=<1.2.3.4:10100>```

```--listen-address (optional): bind to address:port for incoming miner connections. By default, proxy listens on 0.0.0.0:10200
--daemon-address: address:port of daemon
--log-interval (optional): logging every X seconds, where X >= 60. Default is 60 seconds
--minimal (optional): forward only 2 jobs per block (1 for first 9 miniblocks, 1 for final miniblock)
--nonce (optional): enable random nonces, disabled by default```
