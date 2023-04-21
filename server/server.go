package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/OmarTariq612/go-wstunnel/util"
	"nhooyr.io/websocket"
)

type Server struct {
	localAddr string
}

func NewServer(localAddr string) *Server {
	return &Server{localAddr: localAddr}
}

func (s *Server) ListenAndServe() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dst, err := util.ParseURLDst(r.URL)
		if err != nil {
			if err = s.reject(w, r, http.StatusBadRequest, err.Error()); err != nil {
				log.Println(err)
			}
			return
		}

		tcpConn, err := net.Dial("tcp", dst)
		if err != nil {
			if err = s.reject(w, r, http.StatusInternalServerError, err.Error()); err != nil {
				log.Println(err)
			}
			return
		}
		defer tcpConn.Close()

		wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true, Subprotocols: []string{util.WSProtocol}})
		if err != nil {
			log.Println(err)
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
	})

	return http.ListenAndServe(s.localAddr, nil)
}

func (s *Server) reject(w http.ResponseWriter, r *http.Request, statusCode int, errorMsg string) error {
	w.Header().Add("Connection", "close")
	w.Header().Add(util.RejectReasonHeader, strconv.Quote(errorMsg))
	w.WriteHeader(statusCode)
	return r.Body.Close()
}
