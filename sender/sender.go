package sender

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"github.com/lazyfrosch/filespooler/util"
	"log"
	"net"
	"strings"
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
	rw        *bufio.ReadWriter
}

func NewSender(addr string, reader *FileReader) *Sender {
	return &Sender{
		addr:   addr,
		reader: reader,
	}
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

	s.rw = bufio.NewReadWriter(bufio.NewReader(s.conn), bufio.NewWriter(s.conn))

	return nil
}

func (s *Sender) Reconnect() {
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
		s.rw = nil
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
		} else {
			if err := s.SendFiles(); err != nil {
				log.Printf("error sending files: %s", err)
				s.Reconnect()
			}
		}

		select {
		case <-s.quit:
			return
		case <-keepalive.C:
			if s.conn == nil {
				continue
			}

			s.setTimeout()
			if _, err := s.rw.WriteString("KEEPALIVE\n"); err != nil {
				log.Printf("error sending keepalive: %s", err)
				s.Reconnect()
			} else {
				_ = s.rw.Flush()
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

		if _, err := s.rw.WriteString("SEND_FILE\n"); err != nil {
			return fmt.Errorf("could not sent command: %s", err)
		}

		enc := gob.NewEncoder(s.rw)
		if err := enc.Encode(file); err != nil {
			return fmt.Errorf("could not send encoded data: %s", err)
		}

		err = s.rw.Flush()
		if err != nil {
			return fmt.Errorf("could not flush data: %s", err)
		}

		s.setTimeout()
		response, err := s.rw.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error waiting for response for sent file: %s", err)
		}

		response = strings.Trim(response, "\n")
		if response == "OK" {
			// Delete file when it was sent
			err = s.reader.Delete(file.RawName)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("peer did not acknowledge file and returned: %s", response)
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
		s.rw = nil
		if err != nil {
			return err
		}
	}

	return nil
}
