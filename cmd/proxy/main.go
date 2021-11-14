package main

import (
	"flag"
	"os"

	"github.com/miguelmota/go-rpc-provider-proxy/pkg/proxy"
)

func main() {
	var port string
	var proxyURL string
	var proxyMethod string
	var logLevel string
	var authorizationSecret string
	var leakyBucketLimitPerSecond int
	var softCapIPRequestsPerMinute int
	var hardCapIPRequestsPerMinute int
	var slackWebhookURL string
	var slackChannel string

	portEnv := os.Getenv("PORT")
	if portEnv != "" {
		port = portEnv
	}

	authSecretEnv := os.Getenv("AUTH_SECRET")

	flag.StringVar(&port, "port", "8000", "Server port")
	flag.StringVar(&proxyURL, "proxy-url", "", "Proxy URL")
	flag.StringVar(&proxyMethod, "proxy-method", "", "Proxy method")
	flag.StringVar(&logLevel, "log-level", "", "Log level")
	flag.StringVar(&authorizationSecret, "auth-secret", authSecretEnv, "Authorization secret")
	flag.IntVar(&leakyBucketLimitPerSecond, "limit-per-second", leakyBucketLimitPerSecond, "Leaky bucket limit per second")
	flag.IntVar(&softCapIPRequestsPerMinute, "soft-cap-ip-requests-per-minute", softCapIPRequestsPerMinute, "Soft cap requests per minute for IP")
	flag.IntVar(&hardCapIPRequestsPerMinute, "hard-cap-ip-requests-per-minute", hardCapIPRequestsPerMinute, "Hard cap requests per minute for IP")
	flag.StringVar(&slackWebhookURL, "slack-webhook-url", slackWebhookURL, "Slack Webhook URL")
	flag.StringVar(&slackChannel, "slack-channel", slackChannel, "Slack channel for notifications")
	flag.Parse()

	if proxyURL == "" {
		panic("Flag -proxy-url is required")
	}

	// add always allowed IPs here
	alwaysAllowedIps := []string{
		"3.215.160.175",  // dev server
		"34.193.216.56",  // production server
		"47.147.201.15",  // miguel
		"47.147.192.199", // miguel
		"172.17.0.1",     // miguel
		"172.114.143.76", // chris
		"127.0.0.1",
	}

	// add blocked IPs here
	blockedIps := []string{
		"70.185.111.46", // this ip keeps hitting hard cap on kovan proxy
	}

	rpcProxy := proxy.NewProxy(&proxy.Config{
		ProxyURL:                   proxyURL,
		ProxyMethod:                proxyMethod,
		Port:                       port,
		LogLevel:                   logLevel,
		AuthorizationSecret:        authorizationSecret,
		BlockedIps:                 blockedIps,
		AlwaysAllowedIps:           alwaysAllowedIps,
		LeakyBucketLimitPerSecond:  leakyBucketLimitPerSecond,
		SoftCapIPRequestsPerMinute: softCapIPRequestsPerMinute,
		HardCapIPRequestsPerMinute: hardCapIPRequestsPerMinute,
		SlackWebhookURL:            slackWebhookURL,
		SlackChannel:               slackChannel,
	})

	panic(rpcProxy.Start())
}
