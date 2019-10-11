package sender

/*
import (
	"fmt"
	"io"
	"net"
	"os"
	"testing"
)

const TestListen = "[127.0.0.1]:54111"

func dummyListener(t *testing.T) {
	l, err := net.Listen("tcp", TestListen)
	if err != nil {
		t.Fatal("Could not listen: ", err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			t.Fatal("Error accepting: ", err)
		}

		io.Copy()
	}
}

func TestNewSender(t *testing.T) {
	s := NewSender()
}
*/
