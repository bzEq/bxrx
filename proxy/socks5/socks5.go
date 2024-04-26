package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const VER = 5

const (
	CMD_CONNECT = iota + 1
	CMD_BIND
	CMD_UDP_ASSOCIATE
)

const (
	ATYP_IPV4 = iota + 1
	_
	ATYP_DOMAINNAME
	ATYP_IPV6
)

const (
	REP_SUCC = iota
	REP_GENERAL_SERVER_FAILURE
	REP_CONNECTION_NOT_ALLOWED
	REP_NETWORK_UNREACHABLE
	REP_HOST_UNREACHABLE
	REP_CONNECTION_REFUSED
	REP_TTL_EXPIRED
	REP_COMMAND_NOT_SUPPORTED
	REP_ADDRESS_TYPE_NOT_SUPPORTED
	REP_UNASSIGNED_START
)

type Request struct {
	VER, CMD, ATYP byte
	DST_ADDR       []byte
	DST_PORT       [2]byte
}

type Reply struct {
	VER, REP, ATYP byte
	BND_ADDR       []byte
	BND_PORT       [2]byte
}

const HANDSHAKE_TIMEOUT = 8

func ExchangeMetadata(rw net.Conn) (err error) {
	buf := make([]byte, 255)
	// VER, NMETHODS.
	rw.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(rw, buf[:2]); err != nil {
		log.Println("Reading VER, NMETHODS failed: ", err)
		return
	}
	// METHODS.
	methods := buf[1]
	rw.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(rw, buf[:methods]); err != nil {
		log.Println("Reading METHODS failed: ", err)
		return
	}
	// No auth for now.
	rw.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = rw.Write([]byte{VER, 0}); err != nil {
		log.Println("Writing VER failed:", err)
		return
	}
	return
}

func ReceiveRequest(r net.Conn) (req Request, err error) {
	buf := make([]byte, net.IPv6len)
	// VER, CMD, RSV, ATYP
	r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(r, buf[:4]); err != nil {
		log.Println("Reading request failed: ", err)
		return req, err
	}
	req.VER = buf[0]
	req.CMD = buf[1]
	req.ATYP = buf[3]
	switch req.ATYP {
	case ATYP_IPV6:
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, buf[:net.IPv6len]); err != nil {
			log.Println("Reading IPv6 address failed: ", err)
			return
		}
		req.DST_ADDR = make([]byte, net.IPv6len)
		copy(req.DST_ADDR, buf[:net.IPv6len])
	case ATYP_IPV4:
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, buf[:net.IPv4len]); err != nil {
			log.Println("Reading IPv4 address failed: ", err)
			return
		}
		req.DST_ADDR = make([]byte, net.IPv4len)
		copy(req.DST_ADDR, buf[:net.IPv4len])
	case ATYP_DOMAINNAME:
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, buf[:1]); err != nil {
			log.Println("Reading length of domain name failed: ", err)
			return
		}
		req.DST_ADDR = make([]byte, buf[0])
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, req.DST_ADDR); err != nil {
			log.Println("Reading domain name failed: ", err)
			return
		}
	default:
		err = fmt.Errorf("Unsupported ATYP: %d", req.ATYP)
		log.Println(err)
		return req, err
	}
	r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	_, err = io.ReadFull(r, buf[:2])
	if err != nil {
		log.Println("Reading port failed: ", err)
		return req, err
	}
	copy(req.DST_PORT[:2], buf[:2])
	return req, nil
}

func SendReply(w net.Conn, r Reply) (err error) {
	// FIXME: Respect Reply.
	w.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = w.Write([]byte{r.VER, r.REP, 0, r.ATYP}); err != nil {
		return
	}
	w.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = w.Write(r.BND_ADDR); err != nil {
		return
	}
	w.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = w.Write(r.BND_PORT[:]); err != nil {
		return
	}
	return
}

func GetDialAddress(req Request) string {
	port := fmt.Sprintf("%d", binary.BigEndian.Uint16(req.DST_PORT[:2]))
	switch req.ATYP {
	case ATYP_IPV4, ATYP_IPV6:
		return net.JoinHostPort(net.IP(req.DST_ADDR).String(), port)
	case ATYP_DOMAINNAME:
		return net.JoinHostPort(string(req.DST_ADDR), port)
	default:
		return ""
	}
}
