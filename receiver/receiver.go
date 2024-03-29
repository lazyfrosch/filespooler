package receiver

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/lazyfrosch/filespooler/sender"
	"github.com/lazyfrosch/filespooler/util"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	ReadTimeout          = 5
	CommunicationTimeout = 60
)

type Receiver struct {
	bind      string
	listener  *net.TCPListener
	writer    *FileWriter
	running   bool
	quit      chan bool
	exited    chan bool
	TlsConfig *tls.Config
	PeerNames []string
}

func NewReceiver(bind string, writer *FileWriter) *Receiver {
	return &Receiver{
		bind:   bind,
		writer: writer,
	}
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
			r.running = false
			handlers.Wait()
			close(r.exited)
			return
		default:
			err := r.listener.SetDeadline(time.Now().Add(1e9))
			if err != nil {
				log.Print("Could not setup deadline: ", err.Error())
				return
			}

			conn, err := r.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				log.Println("Failed to accept connection:", err.Error())
				continue
			}

			var tlsConn *tls.Conn
			if r.TlsConfig != nil {
				remote := conn.RemoteAddr().String()
				err := conn.SetReadDeadline(time.Now().Add(ReadTimeout * time.Second))
				if err != nil {
					log.Printf("[%s] Could not set deadline: %s", remote, err)
					return
				}

				tlsConn = tls.Server(conn, r.TlsConfig)

				err = tlsConn.Handshake()
				if err != nil {
					log.Printf("[%s] TLS Handshake failed: %s", remote, err)
					_ = conn.Close()
					continue
				}

				// authenticate peer against whitelist
				state := tlsConn.ConnectionState()
				if len(state.PeerCertificates) > 0 {
					clientCert := state.PeerCertificates[0]

					if ok, name := util.ValidateNamesOnCertificate(clientCert, r.PeerNames); ok {
						log.Printf("[%s] client cert accepted with name %s", remote, name)
					} else {
						log.Printf("[%s] client cert names did not match whitelist: %s", remote, r.PeerNames)
						_ = conn.Close()
						continue
					}
				}
			}

			handlers.Add(1)
			go func() {
				if tlsConn != nil {
					r.handleConnection(tlsConn)
				} else {
					r.handleConnection(conn)
				}
				handlers.Done()
			}()
		}
	}
}

func (r *Receiver) handleConnection(conn net.Conn) {
	remote := conn.RemoteAddr()
	log.Printf("[%s] accepted new connection", remote)

	timer := time.NewTimer(CommunicationTimeout * time.Second)

	defer func() {
		log.Printf("[%s] closing connection", remote)
		_ = conn.Close()
		timer.Stop()
	}()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		select {
		case <-r.quit:
			// The receiver wants to stop
			return
		case <-timer.C:
			// Timeout on connection
			log.Printf("[%s] No data received in %d seconds. Disconnecting", remote, CommunicationTimeout)
			return
		default:
			err := conn.SetReadDeadline(time.Now().Add(ReadTimeout * time.Second))
			if err != nil {
				log.Printf("[%s] Could not set deadline: %s", remote, err)
				return
			}

			cmd, err := rw.ReadString('\n')
			switch {
			case err == io.EOF:
				log.Printf("[%s] Connection EOF", remote)
				return
			case err != nil:
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					// Timeout from deadline, retry read in next loop
					continue
				}
				log.Printf("[%s] Could not read from stream: %s", remote, err)
				return
			}
			cmd = strings.Trim(cmd, "\n ")
			switch cmd {
			case "SEND_FILE":
				err := r.handleSendFile(conn, rw)
				if err != nil {
					log.Print(err)
					return
				}
			case "NOOP":
			case "KEEPALIVE":
				// resetting timeout
				timer.Reset(CommunicationTimeout * time.Second)
			default:
				log.Printf("[%s] Unknown command: %s", remote, cmd)
			}
		}
	}
}

func (r *Receiver) handleSendFile(conn net.Conn, rw *bufio.ReadWriter) error {
	remote := conn.RemoteAddr()
	file, err := sender.DecodeGobFileData(rw)
	if err != nil {
		return fmt.Errorf("[%s] Could not decode file: %s", remote, err)
	}

	log.Printf("[%s] Received file %s", remote, file.Name())

	err = r.writer.WriteFile(file)
	if err != nil {
		_, err = rw.WriteString("ERR\n")
		if err != nil {
			return fmt.Errorf("could not write ERR response: %s", err)
		}
		return fmt.Errorf("[%s] Could not write file: %s", remote, err)
	}

	_, err = rw.WriteString("OK\n")
	if err != nil {
		return fmt.Errorf("could not write OK response: %s", err)
	}

	err = rw.Flush()
	if err != nil {
		return fmt.Errorf("error flushing buffer: %s", err)
	}

	return nil
}

func (r *Receiver) Close() {
	if r.running == true {
		log.Println("Stopping receiver")
		close(r.quit)
		<-r.exited
	}
	_ = r.listener.Close()
}
