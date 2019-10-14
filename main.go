package main

import (
	"flag"
	"fmt"
	"github.com/lazyfrosch/filespooler/receiver"
	"github.com/lazyfrosch/filespooler/sender"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	sourcePath := cmd.String("source", "", "Source path to read from")

	_ = cmd.Parse(os.Args[2:])

	if flag.NArg() > 0 {
		log.Fatalf("Found extra arguments: %v", flag.Args())
	}

	if *connect == "" {
		log.Fatal("Please specify --connect")
	}
	if *sourcePath == "" {
		log.Fatal("Please specify --source")
	}

	log.Printf("Starting sender to %s", *connect)
	log.Printf("Reading data from %s", *sourcePath)

	r, err := sender.NewFileReader(*sourcePath)
	if err != nil {
		log.Fatal("Could not set up FileReader: ", err)
	}

	signals := make(chan os.Signal, 1)
	quit := make(chan bool)
	done := make(chan bool)

	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	s := sender.NewSender(*connect, r)

	go func() {
		sig := <-signals
		log.Printf("Got signal %v from OS", sig)
		s.Stop()
		close(quit)
		done <- true
	}()

	s.Run()
	_ = s.Close()

	<-done
	log.Println("Exiting sender")
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
