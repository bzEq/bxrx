// Copyright (c) 2024 Kai Luo <gluokai@gmail.com>. All rights reserved.

package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"

	"github.com/bzEq/bxrx/core"
	h1p "github.com/bzEq/bxrx/proxy/http"
	"github.com/bzEq/bxrx/relayer"
)

var options relayer.Options

func proxyLocalHTTP(be core.Backend) {
	socksProxyURL, err := url.Parse("socks5://" + options.LocalAddr)
	if err != nil {
		log.Println(err)
		return
	}
	fe := relayer.NewHTTPProxyFE()
	proxy := &h1p.HTTPProxy{
		Transport: &http.Transport{Proxy: http.ProxyURL(socksProxyURL)},
		Relay:     fe.Capture,
	}
	server := &http.Server{
		Addr:    options.LocalHTTPProxy,
		Handler: proxy,
	}
	log.Println("Starting http proxy on", options.LocalHTTPProxy)
	go server.ListenAndServe()
	if err := core.NewRelayer(fe, be).Relay(); err != nil {
		log.Println(err)
	}
}

func relay() {
	log.Println("Listening on", options.LocalAddr)
	ln, err := net.Listen("tcp", options.LocalAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()
	var fe core.Frontend
	var be core.Backend
	pipeline := &relayer.Pipeline{}
	if options.NextHop == "" {
		fe = relayer.NewWrapFE(ln.(*net.TCPListener), pipeline)
		be = &relayer.TCPBE{}
	} else {
		fe = relayer.NewSocks5FE(ln.(*net.TCPListener))
		log.Println("Backend is connecting to", options.NextHop)
		be = relayer.NewWrapBE(options.NextHop, pipeline)
		if options.LocalHTTPProxy != "" {
			go proxyLocalHTTP(be)
		}
	}
	r := core.NewRelayer(fe, be)
	if err := r.Relay(); err != nil {
		log.Println(err)
		return
	}
}

func main() {
	var seed int64
	binary.Read(crand.Reader, binary.BigEndian, &seed)
	rand.Seed(seed)
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&options.LocalAddr, "l", "localhost:1080", "Listen address of this relayer")
	flag.StringVar(&options.NextHop, "n", "", "Address of next-hop relayer")
	flag.StringVar(&options.LocalHTTPProxy, "http_proxy", "", "Enable this relayer serving as http proxy")
	flag.Parse()
	if !debug {
		log.SetOutput(io.Discard)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	relay()
}
