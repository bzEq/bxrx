package socks5

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/bzEq/bxrx/core"
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
	VER, CMD, RSV, ATYP byte
	DST_ADDR            []byte
	DST_PORT            [2]byte
}

type Reply struct {
	VER, REP, RSV, ATYP byte
	BND_ADDR            []byte
	BND_PORT            [2]byte
}

type UDPRequest struct {
	RSV        [2]byte
	FRAG, ATYP byte
	DST_ADDR   []byte
	DST_PORT   [2]byte
	DATA       []byte
}

const HANDSHAKE_TIMEOUT = 8

// Read
// +----+----------+----------+
// |VER | NMETHODS | METHODS  |
// +----+----------+----------+
// | 1  |    1     | 1 to 255 |
// +----+----------+----------+
// Write
// +----+--------+
// |VER | METHOD |
// +----+--------+
// | 1  |   1    |
// +----+--------+
func ExchangeMetadata(rw net.Conn) (err error) {
	buf := make([]byte, 255)
	// VER, NMETHODS.
	rw.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(rw, buf[:2]); err != nil {
		err = core.Tr(fmt.Errorf("Reading VER, NMETHODS failed: %w", err))
		return
	}
	// METHODS.
	methods := buf[1]
	rw.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(rw, buf[:methods]); err != nil {
		err = core.Tr(fmt.Errorf("Reading METHODS failed: %w", err))
		return
	}
	// No auth for now.
	rw.SetWriteDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = rw.Write([]byte{VER, 0}); err != nil {
		err = core.Tr(fmt.Errorf("Writing VER failed: %w", err))
		return
	}
	return
}

// +----+-----+-------+------+----------+----------+
// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
// +----+-----+-------+------+----------+----------+
// | 1  |  1  | X'00' |  1   | Variable |    2     |
// +----+-----+-------+------+----------+----------+
func ReceiveRequest(r net.Conn, req *Request) (err error) {
	buf := make([]byte, 1024)
	// VER, CMD, RSV, ATYP
	r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	if _, err = io.ReadFull(r, buf[:4]); err != nil {
		err = core.Tr(fmt.Errorf("Reading request failed: %w", err))
		return err
	}
	req.VER = buf[0]
	req.CMD = buf[1]
	req.RSV = buf[2]
	req.ATYP = buf[3]
	switch req.ATYP {
	case ATYP_IPV6:
		req.DST_ADDR = make([]byte, net.IPv6len)
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, req.DST_ADDR); err != nil {
			err = core.Tr(fmt.Errorf("Reading IPv6 address failed: %w", err))
			return
		}
	case ATYP_IPV4:
		req.DST_ADDR = make([]byte, net.IPv4len)
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, req.DST_ADDR); err != nil {
			err = core.Tr(fmt.Errorf("Reading IPv4 address failed: %w", err))
			return
		}
	case ATYP_DOMAINNAME:
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, buf[:1]); err != nil {
			err = core.Tr(fmt.Errorf("Reading length of domain name failed: %w", err))
			return
		}
		req.DST_ADDR = make([]byte, int(buf[0])+1)
		req.DST_ADDR[0] = buf[0]
		r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
		if _, err = io.ReadFull(r, req.DST_ADDR[1:]); err != nil {
			err = core.Tr(fmt.Errorf("Reading domain name failed: %w", err))
			return
		}
	default:
		err = core.Tr(fmt.Errorf("Unsupported ATYP: %d", req.ATYP))
		return err
	}
	r.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMEOUT * time.Second))
	_, err = io.ReadFull(r, req.DST_PORT[:])
	if err != nil {
		err = core.Tr(fmt.Errorf("Reading port failed: %w", err))
		return
	}
	return nil
}

// +----+-----+-------+------+----------+----------+
// |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
// +----+-----+-------+------+----------+----------+
// | 1  |  1  | X'00' |  1   | Variable |    2     |
// +----+-----+-------+------+----------+----------+
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

func GetDialAddress(atyp byte, addr []byte, port [2]byte) string {
	p := fmt.Sprintf("%d", binary.BigEndian.Uint16(port[:2]))
	switch atyp {
	case ATYP_IPV4, ATYP_IPV6:
		return net.JoinHostPort(net.IP(addr).String(), p)
	case ATYP_DOMAINNAME:
		return net.JoinHostPort(string(addr[1:]), p)
	default:
		panic(fmt.Sprintf("Unsupported ATYP: %d", atyp))
	}
}

// +----+------+------+----------+----------+----------+
// |RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
// +----+------+------+----------+----------+----------+
// | 2  |  1   |  1   | Variable |    2     | Variable |
// +----+------+------+----------+----------+----------+
func ParseUDPRequest(buf []byte, req *UDPRequest) error {
	r := bytes.NewReader(buf)
	_, err := r.Read(req.RSV[:])
	if err != nil {
		return core.Tr(err)
	}
	req.FRAG, err = r.ReadByte()
	if err != nil {
		return core.Tr(err)
	}
	req.ATYP, err = r.ReadByte()
	if err != nil {
		return core.Tr(err)
	}
	switch req.ATYP {
	case ATYP_IPV4:
		req.DST_ADDR = make([]byte, net.IPv4len)
		_, err = r.Read(req.DST_ADDR)
		if err != nil {
			return core.Tr(err)
		}
	case ATYP_IPV6:
		req.DST_ADDR = make([]byte, net.IPv6len)
		_, err = r.Read(req.DST_ADDR)
		if err != nil {
			return core.Tr(err)
		}
	case ATYP_DOMAINNAME:
		l, err := r.ReadByte()
		if err != nil {
			return core.Tr(err)
		}
		req.DST_ADDR = make([]byte, int(l)+1)
		req.DST_ADDR[0] = l
		_, err = r.Read(req.DST_ADDR[1:])
		if err != nil {
			return core.Tr(err)
		}
	}
	_, err = r.Read(req.DST_PORT[:])
	if err != nil {
		return core.Tr(err)
	}
	req.DATA = make([]byte, r.Len())
	_, err = r.Read(req.DATA)
	if err != nil {
		return core.Tr(err)
	}
	return nil
}
