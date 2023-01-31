package main

import (
	"crypto/tls"
	"flag"
	"http-proxy/proxy"
	"log"

	"net/http"
)

func main() {
	p := &proxy.Proxy{}

	var protocol, pem, key string
	flag.StringVar(&pem, "pem", "RootCA.pem", "")
	flag.StringVar(&key, "key", "RootCA.key", "")
	flag.StringVar(&protocol, "protocol", "http", "")
	flag.Parse()

	server := &http.Server{
		Addr: ":8080",
		Handler: p,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	switch protocol {
	case "http":
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf(err.Error())
		}

		break
	case "https":
		if err := server.ListenAndServeTLS(pem, key); err != nil {
			log.Fatalf(err.Error())
		}

		break
	default:
		log.Println("http or https, not anything else")

		break
	}
}
