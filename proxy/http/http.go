package http

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

// See https://www.rfc-editor.org/rfc/rfc9110.html#field.connection
var HopByHopFields = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"TE",
	"Transfer-Encoding",
	"Upgrade",
}

func RemoveHopByHopFields(header http.Header) {
	for _, f := range HopByHopFields {
		header.Del(f)
	}
}

type HTTPProxy struct {
	Transport http.RoundTripper
	Relay     func(c net.Conn, raddr string)
}

func (self *HTTPProxy) handleConnect(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	h, ok := w.(http.Hijacker)
	if !ok {
		log.Println(fmt.Errorf("Hijacking not supported"))
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	c, _, err := h.Hijack()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if self.Relay == nil {
		log.Println("Nil relay function, failed relaying to", req.Host)
		return
	}
	self.Relay(c, req.Host)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (self *HTTPProxy) handleOther(w http.ResponseWriter, req *http.Request) {
	// To avoid 'Request.RequestURI can't be set in client requests' error.
	req.RequestURI = ""
	RemoveHopByHopFields(req.Header)
	client := &http.Client{Transport: self.Transport}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	RemoveHopByHopFields(resp.Header)
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// Modified from
// https://www.sobyte.net/post/2021-09/https-proxy-in-golang-in-less-than-100-lines-of-code/
func (self *HTTPProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodConnect {
		self.handleConnect(w, req)
	} else {
		self.handleOther(w, req)
	}
}
