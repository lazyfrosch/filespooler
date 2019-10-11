package sender

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	ConnectTimeout    = 5
	KeepaliveInterval = 10
	ReconnectInterval = 30
	FileCheckInterval = 5
	DataTimeout       = 5
)

type Sender struct {
	addr   string
	conn   net.Conn
	reader *FileReader
	quit   chan bool
}

func NewSender(addr string, reader *FileReader) *Sender {
	return &Sender{addr, nil, reader, nil}
}

func (s *Sender) Open() error {
	log.Printf("Connecting to %s", s.addr)

	conn, err := net.DialTimeout("tcp", s.addr, ConnectTimeout*time.Second)
	if err != nil {
		return fmt.Errorf("could not connect to %s: %s", s.addr, err)
	}

	s.conn = conn

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

	reconnect := time.NewTicker(ReconnectInterval * time.Second)
	keepalive := time.NewTicker(KeepaliveInterval * time.Second)
	checkFiles := time.NewTicker(FileCheckInterval * time.Second)

	if s.conn == nil {
		s.Reconnect()
	}

	for {
		select {
		case <-s.quit:
			return
		case <-reconnect.C:
			if s.conn == nil {
				s.Reconnect()
			}
		case <-keepalive.C:
			if s.conn != nil {
				continue
			}

			s.setTimeout()
			if _, err := s.conn.Write([]byte("KEEPALIVE\n")); err != nil {
				log.Printf("error sending keepalive: %s", err)
				s.Reconnect()
			}
		case <-checkFiles.C:
			if s.conn == nil {
				continue
			}

			if err := s.SendFiles(); err != nil {
				log.Printf("error sending files: %s", err)
				s.Reconnect()
			}
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
