package encrypt

import (
	"crypto/tls"
	"net"
	"time"
)

// Based on "golang.org/x/crypto/acme/autocert".

type Listener struct {
	conf        *tls.Config
	tcpListener net.Listener
}

func NewListener(conf *tls.Config, tcpListener net.Listener) (*Listener, error) {
	return &Listener{
		conf:        conf,
		tcpListener: tcpListener,
	}, nil
}

func (ln *Listener) Accept() (net.Conn, error) {
	conn, err := ln.tcpListener.Accept()
	if err != nil {
		return nil, err
	}
	tcpConn := conn.(*net.TCPConn)

	// Because Listener is a convenience function, help out with
	// this too.  This is not possible for the caller to set once
	// we return a *tcp.Conn wrapping an inaccessible net.Conn.
	// If callers don't want this, they can do things the manual
	// way and tweak as needed. But this is what net/http does
	// itself, so copy that. If net/http changes, we can change
	// here too.
	if err := tcpConn.SetKeepAlive(true); err != nil {
		return nil, err
	}
	if err := tcpConn.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		return nil, err
	}

	return tls.Server(tcpConn, ln.conf), nil
}

func (ln *Listener) Addr() net.Addr {
	if ln.tcpListener != nil {
		return ln.tcpListener.Addr()
	}
	// net.Listen failed. Return something non-nil in case callers
	// call Addr before Accept:
	return &net.TCPAddr{IP: net.IP{0, 0, 0, 0}, Port: 443}
}

func (ln *Listener) Close() error {
	return ln.tcpListener.Close()
}
