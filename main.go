package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"castspeak/internal/cli"
	"castspeak/internal/server"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}

	var err error
	args := os.Args[2:]

	switch os.Args[1] {
	case "serve":
		runServe(args)
		return
	case "speak":
		err = cli.RunSpeak(args)
	case "devices":
		err = cli.RunDevices(args)
	case "volume":
		err = cli.RunVolume(args)
	case "mute":
		err = cli.RunMute(args, true)
	case "unmute":
		err = cli.RunMute(args, false)
	case "stop":
		err = cli.RunStop(args)
	case "status":
		err = cli.RunStatus(args)
	case "scan":
		err = cli.RunScan(args)
	case "play":
		err = cli.RunPlay(args)
	case "version", "--version", "-v":
		fmt.Println(version)
		return
	case "--help", "-h", "help":
		cli.PrintUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		cli.PrintUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	portFlag := fs.String("port", "", "Port to listen on (default 8080, or PORT env)")
	fs.Parse(args)

	port := os.Getenv("PORT")
	if *portFlag != "" {
		port = *portFlag
	}
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: server.New(),
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
