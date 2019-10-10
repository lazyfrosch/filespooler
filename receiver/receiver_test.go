package receiver

import (
	"testing"
	"time"
)

func testBind(t *testing.T, addr string, open bool) *Receiver {
	// TODO: temp path
	w, err := NewFileWriter("/tmp/testing")
	if err != nil {
		t.Fatal("Could not setup FileWriter", err)
	}

	r := NewReceiver(addr, w)
	if open {
		err = r.Open()
		if err != nil {
			t.Fatal("Could not open receiver", err)
		}
	}

	return r
}

func TestReceiver(t *testing.T) {
	r := testBind(t, ":12345", true)

	go r.Serve()
	time.Sleep(2)
	r.Close()
}

func TestBindFail(t *testing.T) {
	r1 := testBind(t, ":12345", false)
	err := r1.Open()
	if err != nil {
		t.Fatal("First bind failed", err.Error())
	}
	defer r1.Close()

	r2 := testBind(t, ":12345", false)
	err = r2.Open()
	if err == nil {
		t.Fatal("Second bind did not fail")
	}
}

func TestMultipleAddresses(t *testing.T) {
	addresses := []string{"[::1]:12345", "0.0.0.0:12346"}
	for _, addr := range addresses {
		r := testBind(t, addr, true)
		time.Sleep(1)
		r.Close()
	}
}