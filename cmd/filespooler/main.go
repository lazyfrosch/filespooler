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

func receiverCli(cmdName string, args []string) error {
	cmd := flag.NewFlagSet(cmdName+" receiver", flag.ContinueOnError)
	listen := cmd.String("listen", DefaultListen, "Listen to this address")
	targetPath := cmd.String("target", "", "Target path to write to")

	if err := cmd.Parse(args); err != nil {
		return err
	}

	if flag.NArg() > 0 {
		return fmt.Errorf("found extra arguments: %v", flag.Args())
	}

	if *targetPath == "" {
		return fmt.Errorf("please specify --target")
	}

	log.Printf("Starting listener on %s", *listen)
	log.Printf("Spooling data to %s", *targetPath)

	writer, err := receiver.NewFileWriter(*targetPath)
	if err != nil {
		return fmt.Errorf("could not setup FileWriter: %s", err)
	}

	r := receiver.NewReceiver(*listen, writer)
	if err = r.Open(); err != nil {
		return fmt.Errorf("could not open listener: %s", err)
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
	return nil
}

func senderCli(cmdName string, args []string) error {
	cmd := flag.NewFlagSet(cmdName+" sender", flag.ContinueOnError)
	connect := cmd.String("connect", "", "Send to this TCP address")
	sourcePath := cmd.String("source", "", "Source path to read from")

	if err := cmd.Parse(args); err != nil {
		return err
	}

	if flag.NArg() > 0 {
		return fmt.Errorf("found extra arguments: %v", flag.Args())
	}

	if *connect == "" {
		return fmt.Errorf("please specify --connect")
	}
	if *sourcePath == "" {
		return fmt.Errorf("please specify --source")
	}

	log.Printf("Starting sender to %s", *connect)
	log.Printf("Reading data from %s", *sourcePath)

	r, err := sender.NewFileReader(*sourcePath)
	if err != nil {
		return fmt.Errorf("could not set up FileReader: %s", err)
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
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "receiver|sender [options]")
		os.Exit(2)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cmd := os.Args[0]
	args := os.Args[2:]
	var err error

	switch os.Args[1] {
	case "receiver":
		err = receiverCli(cmd, args)
	case "sender":
		err = senderCli(cmd, args)
	default:
		err = fmt.Errorf("unknown mode: %s", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}
}
