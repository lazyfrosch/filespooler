package receiver

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Receiver struct {
	bind     string
	listener *net.TCPListener
	writer   *FileWriter
	running  bool
	quit     chan bool
	exited   chan bool
}

func NewReceiver(bind string, writer *FileWriter) *Receiver {
	return &Receiver{
		bind, nil, writer, false,
		make(chan bool), make(chan bool)}
}

func (r *Receiver) Open() error {
	addr, err := net.ResolveTCPAddr("tcp", r.bind)
	if err != nil {
		return fmt.Errorf("could not resolve TCP listen address: %s", err.Error())
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	r.listener = listener

	r.quit = make(chan bool)
	r.exited = make(chan bool)
	return nil
}

func (r *Receiver) Serve() {
	var handlers sync.WaitGroup

	for {
		r.running = true

		select {
		case <-r.quit:
			log.Println("Shutting down listener")
			_ = r.listener.Close()
			handlers.Wait()
			close(r.exited)
			r.running = false
			return
		default:
			err := r.listener.SetDeadline(time.Now().Add(1e9))
			if err != nil {
				log.Fatal("Could not setup deadline: ", err.Error())
			}

			conn, err := r.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				log.Println("Failed to accept connection:", err.Error())
				continue
			}

			handlers.Add(1)
			go func() {
				r.handleConnection(conn)
				handlers.Done()
			}()
		}
	}
}

func (r *Receiver) handleConnection(conn net.Conn) {
	log.Println("Accepted connection from", conn.RemoteAddr())

	defer func() {
		log.Println("Closing connection from", conn.RemoteAddr())
		_ = conn.Close()
	}()

	// TODO: do something useful
	_, _ = io.Copy(conn, conn)
}

func (r *Receiver) Close() {
	if r.running == true {
		log.Println("Stopping receiver")
		close(r.quit)
		<-r.exited
	}
	_ = r.listener.Close()
}
