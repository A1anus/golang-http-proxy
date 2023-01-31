package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func copyHeader(to, from http.Header) {
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

type Proxy struct {}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		handleHTTPS(w, r)
	} else {
		handleHTTP(w, r)
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	r.RequestURI = ""
	r.Header.Del("Proxy-Connection")

	httpClient := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	proxyResponse, err := httpClient.Do(r)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer proxyResponse.Body.Close()

	copyHeader(w.Header(), proxyResponse.Header)
	w.WriteHeader(proxyResponse.StatusCode)
	io.Copy(w, proxyResponse.Body)
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	connDest, err := connectHandshake(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	connSrc, _, err := hijacker.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	go exchangeData(connDest, connSrc)
	go exchangeData(connSrc, connDest)
}

func exchangeData(to io.WriteCloser, from io.ReadCloser) {
	defer func() {
		to.Close()
		from.Close()
	}()

	io.Copy(to, from)
}

func connectHandshake(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", r.Host, 10000 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	w.WriteHeader(http.StatusOK)

	return conn, nil
}
