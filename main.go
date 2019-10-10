package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lazyfrosch/filespool/receiver"
)

const (
	// DefaultListen is the default TCP port to use
	DefaultListen = ":5664"
)

func receiverCli() {
	cmd := flag.NewFlagSet(os.Args[0]+" receiver", flag.ExitOnError)
	listen := cmd.String("listen", DefaultListen, "Listen to this address")
	targetPath := cmd.String("target", "", "Target path to write to")

	_ = cmd.Parse(os.Args[2:])

	if flag.NArg() > 0 {
		log.Fatalf("Found extra arguments: %v", flag.Args())
	}

	if *targetPath == "" {
		log.Fatal("Please specify --target")
	}

	log.Printf("Starting listener on %s", *listen)
	log.Printf("Spooling data to %s", *targetPath)

	writer, err := receiver.NewFileWriter(*targetPath)
	if err != nil {
		log.Fatal("Could not setup FileWriter:", err.Error())
	}

	r := receiver.NewReceiver(*listen, writer)
	err = r.Open()
	if err != nil {
		log.Fatal("Could not open listener:", err.Error())
	}

	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-signals
		log.Printf("Got signal %v from OS", sig)
		r.Close()
		done <- true
	}()

	go func() {
		r.Serve()
		r.Close()
		done <- true
	}()

	<-done
	log.Println("Exiting daemon")
}

func senderCli() {
	cmd := flag.NewFlagSet(os.Args[0]+" sender", flag.ExitOnError)
	connect := cmd.String("connect", "", "Send to this TCP address")
	sourcePath := cmd.String("source", "/somepath", "Source path to read from")

	_ = cmd.Parse(os.Args[2:])

	if flag.NArg() > 0 {
		log.Fatalf("Found extra arguments: %v", flag.Args())
	}

	// TODO: fail when connect is empty!

	log.Printf("Starting sender to %s", *connect)
	log.Printf("Reading data from %s", *sourcePath)

	// TODO: implement
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "receiver|sender [options]")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "receiver":
		receiverCli()
	case "sender":
		senderCli()
	default:
		log.Fatal("Unknown mode:", os.Args[1])
	}
}
