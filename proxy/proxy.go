package proxy

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"go.uber.org/ratelimit"
)

// Proxy ...
type Proxy struct {
	httpClient          *http.Client
	proxyURL            string
	proxyMethod         string
	port                string
	maxIdleConnections  int
	requestTimeout      int
	method              string
	sessionID           int
	logLevel            string
	authorizationSecret string
	ratelimit           ratelimit.Limiter
}

// Config ...
type Config struct {
	ProxyURL            string
	ProxyMethod         string
	Port                string
	LogLevel            string
	AuthorizationSecret string
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

	perSecond := 10
	rl := ratelimit.New(perSecond)

	return &Proxy{
		port:                port,
		proxyURL:            config.ProxyURL,
		proxyMethod:         method,
		maxIdleConnections:  100,
		requestTimeout:      3600,
		sessionID:           0,
		logLevel:            config.LogLevel,
		authorizationSecret: config.AuthorizationSecret,
		ratelimit:           rl,
	}
}

// PingHandler ...
func (p *Proxy) PingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}

// HealthCheckHandler ...
func (p *Proxy) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}

// ProxyHandler ...
func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	p.ratelimit.Take()
	p.sessionID++
	sessionID := p.sessionID

	ipAddress := r.RemoteAddr

	forwardedIP := r.Header.Get("X-Forwarded-For")
	if forwardedIP != "" {
		ipAddress = forwardedIP
	}

	blockedIps := map[string]bool{
		"201.1.208.133": true,
	}

	if _, ok := blockedIps[ipAddress]; ok {
		err := errors.New("Too many requests: Ip address blocked")
		fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
		http.Error(w, "", http.StatusTooManyRequests)
		return
	}

	// check base64 encoded bearer token if auth check enabled
	if p.authorizationSecret != "" {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer")
		if (len(splitToken)) != 2 {
			err := errors.New("Unauthorized: Auth token is required")
			fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		reqToken = strings.TrimSpace(splitToken[1])
		decoded, err := base64.StdEncoding.DecodeString(reqToken)
		if err != nil {
			fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		decodedToken := string(decoded)
		if p.authorizationSecret != decodedToken {
			err := errors.New("Unauthorized: Invalid auth token")
			fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
	}

	bodyBuf, _ := ioutil.ReadAll(r.Body)

	// make copies
	bodyRdr1 := ioutil.NopCloser(bytes.NewBuffer(bodyBuf))
	bodyRdr2 := ioutil.NopCloser(bytes.NewBuffer(bodyBuf))

	requestBody, err := ioutil.ReadAll(bodyRdr1)
	if err != nil {
		fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if p.logLevel == "debug" {
		fmt.Printf("REQUEST ID=%v: %s [%s] %s %s %s %s\n", sessionID, ipAddress, time.Now().String(), r.Method, r.URL.String(), r.UserAgent(), string(requestBody))
	}

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

	req, err := http.NewRequest(p.proxyMethod, p.proxyURL, bodyRdr2)
	if err != nil {
		fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// copy headers to request
	for k, v := range r.Header {
		req.Header.Set(k, v[0])
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Host")

	// setting the content length disables chunked transfer encoding,
	// which is required to make proxy work with Alchemy
	req.ContentLength = int64(len(requestBody))

	if p.logLevel == "debug" {
		httpMsg, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		fmt.Println(string(httpMsg))
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// re-use connection
	defer resp.Body.Close()

	// response body
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("ERROR ID=%v: %s\n", sessionID, err)
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

	if p.logLevel == "debug" {
		fmt.Printf("RESPONSE ID=%v: %s [%s] %v %s %s %s\n", sessionID, ipAddress, time.Now().String(), resp.StatusCode, r.Method, r.URL, body)
	}

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
	http.HandleFunc("/health", p.HealthCheckHandler)
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
