package wrap

type TCPRequest struct {
	// Do not use net.TCPAddr here, since we intend to let the remote peer to
	// resolve the domain name.
	Addr string
}
