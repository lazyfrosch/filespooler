package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/Showmax/go-fqdn"
	"github.com/lazyfrosch/filespooler/receiver"
	"github.com/lazyfrosch/filespooler/sender"
	"github.com/lazyfrosch/filespooler/util"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	// DefaultPort is the default TCP port to use
	DefaultPort = "5664"
)

func buildFlagSet(command string) *flag.FlagSet {
	return flag.NewFlagSet(os.Args[0]+" "+command, flag.ContinueOnError)
}

func askForTLSSettings(set *flag.FlagSet) (*string, *string, *string) {
	config := "etc/" // TODO: better handling
	hostname := fqdn.Get()

	return set.String("cert", config+"/"+hostname+".crt", "TLS x509 certificate"),
		set.String("key", config+"/"+hostname+".key", "TLS private key for the certificate"),
		set.String("capath", config+"/ca.crt", "CA Root certificates file")
}

func receiverCli(cmdName string, args []string) error {
	cmd := buildFlagSet("receiver")
	listen := cmd.String("listen", ":"+DefaultPort, "Listen to this address")
	targetPath := cmd.String("target", "", "Target path to write to")

	var peerNames util.ArrayFlags
	cmd.Var(&peerNames, "allow", "Allowed client certificate names, can be repeated to build a list")

	tlsCert, tlsKey, caPath := askForTLSSettings(cmd)

	if err := cmd.Parse(args); err != nil {
		return err
	}

	if flag.NArg() > 0 {
		return fmt.Errorf("found extra arguments: %v", flag.Args())
	}

	if *targetPath == "" {
		return fmt.Errorf("please specify --target")
	}

	if *tlsCert == "" {
		return fmt.Errorf("please specify --cert")
	}
	if *tlsKey == "" {
		return fmt.Errorf("please specify --key")
	}

	if len(peerNames) == 0 {
		return fmt.Errorf("please specify one or more --allow")
	}

	config := util.TlsConfig{
		CAPath:   caPath,
		CertPath: tlsCert,
		KeyPath:  tlsKey,
	}

	tlsConfig, err := config.GetConfig()
	if err != nil {
		return err
	}

	// Enable client auth
	// TODO: allow insecure?
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	log.Printf("Starting listener on %s", *listen)
	log.Printf("Spooling data to %s", *targetPath)

	writer, err := receiver.NewFileWriter(*targetPath)
	if err != nil {
		return fmt.Errorf("could not setup FileWriter: %s", err)
	}

	r := receiver.NewReceiver(*listen, writer)
	r.TlsConfig = tlsConfig
	r.PeerNames = peerNames

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
	cmd := flag.NewFlagSet("sender", flag.ContinueOnError)
	connect := cmd.String("connect", "", "Send to this TCP address")
	sourcePath := cmd.String("source", "", "Source path to read from")

	tlsCert, tlsKey, caPath := askForTLSSettings(cmd)

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

	if *tlsCert == "" {
		return fmt.Errorf("please specify --cert")
	}
	if *tlsKey == "" {
		return fmt.Errorf("please specify --key")
	}

	if !strings.Contains(*connect, ":") {
		*connect = *connect + ":" + DefaultPort
	}

	config := util.TlsConfig{
		CAPath:   caPath,
		CertPath: tlsCert,
		KeyPath:  tlsKey,
	}

	tlsConfig, err := config.GetConfig()
	if err != nil {
		return err
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
	s.TlsConfig = tlsConfig

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
