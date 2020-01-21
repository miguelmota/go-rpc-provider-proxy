package main

import (
	"flag"

	proxy "github.com/authereum/go-rpc-provider-proxy/proxy"
)

func main() {
	var port string
	var proxyURL string
	var proxyMethod string

	flag.StringVar(&port, "port", "8000", "Server port")
	flag.StringVar(&proxyURL, "proxy-url", "", "Proxy URL")
	flag.StringVar(&proxyMethod, "proxy-method", "", "Proxy method")
	flag.Parse()

	if proxyURL == "" {
		panic("Flag -proxy-url is required")
	}

	rpcProxy := proxy.NewProxy(&proxy.Config{
		ProxyURL:    proxyURL,
		ProxyMethod: proxyMethod,
		Port:        port,
	})

	panic(rpcProxy.Start())
}
