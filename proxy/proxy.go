package proxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Proxy ...
type Proxy struct {
	httpClient         *http.Client
	proxyURL           string
	proxyMethod        string
	port               string
	maxIdleConnections int
	requestTimeout     int
	method             string
}

// Config ...
type Config struct {
	ProxyURL    string
	ProxyMethod string
	Port        string
}

// NewProxy ...
func NewProxy(config *Config) *Proxy {
	if config == nil {
		panic("Proxy config is required")
	}

	port := "8000"
	if config.Port != "" {
		port = config.Port
	}

	method := "GET"
	if config.ProxyMethod != "" {
		method = strings.ToUpper(config.ProxyMethod)
	}

	return &Proxy{
		port:               port,
		proxyURL:           config.ProxyURL,
		proxyMethod:        method,
		maxIdleConnections: 100,
		requestTimeout:     3600,
	}
}

// PingHandler ...
func (p *Proxy) PingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}

// ProxyHandler ...
func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Del("Access-Control-Allow-Credentials")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept,Origin,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Range,Range")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE,PATCH")
		w.Header().Set("Access-Control-Max-Age", "1728000")
		w.Header().Set("Content-Type", "text/plain charset=UTF-8")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(204)
		return
	}

	if r.Method != p.proxyMethod {
		http.Error(w, "Not supported", http.StatusNotFound)
		return
	}

	req, err := http.NewRequest(p.proxyMethod, p.proxyURL, r.Body)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// copy headers to request
	for k, v := range r.Header {
		req.Header.Set(k, v[0])
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Host")
	req.Header.Del("Content-Length")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// re-use connection
	defer resp.Body.Close()

	// response body
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}

	w.Header().Del("Access-Control-Allow-Credentials")
	w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept,Origin,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Range,Range")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE,PATCH")

	w.WriteHeader(200)
	w.Write(body)
}

// Start ...
func (p *Proxy) Start() error {
	httpClient, err := p.createHTTPClient()
	if err != nil {
		return err
	}

	p.httpClient = httpClient

	host := "0.0.0.0:" + p.port
	http.HandleFunc("/ping", p.PingHandler)
	http.HandleFunc("/", p.ProxyHandler)

	fmt.Printf("Proxying %s %s\n", p.proxyMethod, p.proxyURL)

	fmt.Println("Listening on port " + p.port)
	return http.ListenAndServe(host, nil)
}

func (p *Proxy) createHTTPClient() (*http.Client, error) {
	transport := &http.Transport{
		MaxIdleConnsPerHost: p.maxIdleConnections,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(p.requestTimeout) * time.Second,
	}

	return client, nil
}
