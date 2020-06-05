# go-rpc-provider-proxy

> A simple Go HTTP server that proxies RPC provider requests.

## Getting started

```bash
# terminal 1
$ go run cmd/proxy/main.go -proxy-url="https://kovan.rpc.authereum.com" -proxy-method=POST -port=800
Proxying POST https://kovan.rpc.authereum.com/
Listening on port 8000

# terminal 2
$ curl http://localhost:8000 -X POST -H "content-type: application/json" -d '{"method":"eth_getCode","params":["0xf2b139bd79e08f9273e6a3dc2702051e1b16cdf8","latest"],"id":13009,"jsonrpc":"2.0"}'
```

Authorization header example:

```bash
# terminal 1
$ go run cmd/proxy/main.go -proxy-url="https://kovan.rpc.authereum.com" -proxy-method=POST -port=800 -auth-secret=mysecret

# terminal 2
$ echo -n mysecret | openssl base64
bXlzZWNyZXQ=

$ curl http://localhost:8000 -X POST -H "content-type: application/json" -H "Authorization: Bearer bXlzZWNyZXQ=" -d '{"method":"eth_getCode","params":["0xf2b139bd79e08f9273e6a3dc2702051e1b16cdf8","latest"],"id":13009,"jsonrpc":"2.0"}'
```

## License

[MIT](LICENSE)
