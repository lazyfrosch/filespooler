package sender

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"github.com/lazyfrosch/filespooler/util"
	"log"
	"net"
	"time"
)

const (
	ConnectTimeout    = 5
	KeepaliveInterval = 10
	FileCheckInterval = 5
	DataTimeout       = 5
)

type Sender struct {
	addr      string
	conn      net.Conn
	reader    *FileReader
	quit      chan bool
	TlsConfig *tls.Config
}

func NewSender(addr string, reader *FileReader) *Sender {
	return &Sender{addr, nil, reader, nil, nil}
}

func (s *Sender) Open() error {
	log.Printf("Connecting to %s", s.addr)

	_, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("could not parse address %s: %s", s.addr, err)
	}

	conn, err := net.DialTimeout("tcp", s.addr, ConnectTimeout*time.Second)
	if err != nil {
		return fmt.Errorf("could not connect to %s: %s", s.addr, err)
	}

	if s.TlsConfig != nil {
		var tlsConn *tls.Conn

		err := conn.SetReadDeadline(time.Now().Add(ConnectTimeout * time.Second))
		if err != nil {
			return fmt.Errorf("could not set deadline: %s", err)
		}

		s.TlsConfig.ServerName = util.GetNameFromTCPAddr(s.addr)

		tlsConn = tls.Client(conn, s.TlsConfig)

		err = tlsConn.Handshake()
		if err != nil {
			_ = conn.Close()
			return fmt.Errorf("TLS Handshake failed: %s", err)
		}

		s.conn = tlsConn
	} else {
		s.conn = conn
	}

	return nil
}

func (s *Sender) Reconnect() {
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
	}

	if err := s.Open(); err != nil {
		log.Printf("error connecting to server: %s", err)
	}
}

func (s *Sender) setTimeout() {
	_ = s.conn.SetDeadline(time.Now().Add(DataTimeout * time.Second))
}

func (s *Sender) Run() {
	s.quit = make(chan bool)

	keepalive := time.NewTicker(KeepaliveInterval * time.Second)
	checkFiles := time.NewTicker(FileCheckInterval * time.Second)

	for {
		if s.conn == nil {
			s.Reconnect()
		}

		if err := s.SendFiles(); err != nil {
			log.Printf("error sending files: %s", err)
			s.Reconnect()
		}

		select {
		case <-s.quit:
			return
		case <-keepalive.C:
			if s.conn == nil {
				continue
			}

			s.setTimeout()
			if _, err := s.conn.Write([]byte("KEEPALIVE\n")); err != nil {
				log.Printf("error sending keepalive: %s", err)
				s.Reconnect()
			}
		case <-checkFiles.C:
			continue
		}
	}
}

func (s *Sender) SendFiles() error {
	files, err := s.reader.ReadDir()
	if err != nil {
		return err
	}

	for _, file := range files {
		s.setTimeout()

		log.Printf("Sending file %s", file.RawName)

		if _, err := s.conn.Write([]byte("SEND_FILE\n")); err != nil {
			return fmt.Errorf("could not sent command: %s", err)
		}

		enc := gob.NewEncoder(s.conn)
		if err := enc.Encode(file); err != nil {
			return fmt.Errorf("could not send encoded data: %s", err)
		}

		// TODO: handle response

		// Delete file when it was sent
		err = s.reader.Delete(file.RawName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sender) Stop() {
	close(s.quit)
}

func (s *Sender) Close() error {
	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		if err != nil {
			return err
		}
	}

	return nil
}
