package net

import (
	"crypto/rand"
	"crypto/rsa"
	"net"

	"github.com/zeppelinmc/zeppelin/net/packet/status"
)

type Config struct {
	Status StatusProvider

	IP                   net.IP
	Port                 int
	CompressionThreshold int32
	Encrypt              bool
	Authenticate         bool
}

type StatusProvider func() status.StatusResponseData

func (c Config) New() (*Listener, error) {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP: c.IP, Port: c.Port,
	})
	if err != nil {
		return nil, err
	}
	lis := &Listener{
		cfg:      c,
		Listener: l,

		newConns: make(chan *Conn),
		err:      make(chan error),
	}
	lis.privKey, err = rsa.GenerateKey(rand.Reader, 1024)

	go lis.listen()

	return lis, err
}

func Status(s status.StatusResponseData) StatusProvider {
	return func() status.StatusResponseData {
		return s
	}
}
