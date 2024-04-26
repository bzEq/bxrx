// Copyright (c) 2024 Kai Luo <gluokai@gmail.com>. All rights reserved.

package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"

	"github.com/bzEq/bxrx/core"
	"github.com/bzEq/bxrx/relayer"
)

var options relayer.Options

func startRelayer() {
	pipeline := &relayer.Pipeline{}
	ln, err := net.Listen("tcp", options.LocalAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()
	var fe core.Frontend
	var be core.Backend
	if options.NextHop == "" {
		fe = relayer.NewWrapFE(ln.(*net.TCPListener), pipeline)
		be = &relayer.TCPBE{}
	} else {
		fe = relayer.NewSocks5FE(ln.(*net.TCPListener))
		be = relayer.NewWrapBE(options.NextHop, pipeline)
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
	flag.Parse()
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	startRelayer()
}
