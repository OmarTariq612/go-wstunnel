package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/OmarTariq612/go-wstunnel/util"
	"nhooyr.io/websocket"
)

type Client struct {
	localAddr  string
	remoteAddr string
	serverAddr string
}

func NewClient(tunnelAddrOptions, serverAddr string) *Client {
	tokens := strings.Split(tunnelAddrOptions, ":")
	var localHost, localPort, remoteHost, remotePort string

	switch len(tokens) {
	case 4:
		localHost = tokens[0]
		localPort = tokens[1]
		remoteHost = tokens[2]
		remotePort = tokens[3]
	case 3:
		localHost = "127.0.0.1"
		localPort = tokens[0]
		remoteHost = tokens[1]
		remotePort = tokens[2]
	}

	return &Client{
		localAddr:  net.JoinHostPort(localHost, localPort),
		remoteAddr: net.JoinHostPort(remoteHost, remotePort),
		serverAddr: serverAddr,
	}
}

func (c *Client) Start() error {
	listener, err := net.Listen("tcp", c.localAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go c.handleConnection(conn)
	}
}

func (c *Client) handleConnection(tcpConn net.Conn) {
	defer tcpConn.Close()
	// wsConn, _, err := websocket.Dial(context.Background(), fmt.Sprintf("ws://%s/?dst=%s", c.serverAddr, c.remoteAddr), &websocket.DialOptions{Subprotocols: []string{util.WSProtocol}})
	wsConn, resp, err := websocket.Dial(context.Background(), fmt.Sprintf("%s?dst=%s", c.serverAddr, c.remoteAddr), &websocket.DialOptions{Subprotocols: []string{util.WSProtocol}})
	if err != nil {
		var msg string
		if resp != nil {
			msg = fmt.Sprintf("err: %v , reason: %s", err, resp.Header.Get(util.RejectReasonHeader))
		} else {
			msg = fmt.Sprintf("err: %v", err)
		}
		log.Print(msg)
		return
	}

	wsNetConn := websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary)
	defer wsNetConn.Close()
	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(tcpConn, wsNetConn)
		if err != nil {
			err = fmt.Errorf("could not write from wsNetConn to tcpConn: %w", err)
		}
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(wsNetConn, tcpConn)
		if err != nil {
			err = fmt.Errorf("could not write from tcpConn to wsNetConn: %w", err)
		}
		errCh <- err
	}()

	if err := <-errCh; err != nil && !errors.Is(err, io.EOF) {
		log.Println(err)
	}
}
