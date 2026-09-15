package main

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func pipe(a, b net.Conn) {
	defer a.Close()
	defer b.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { io.Copy(a, b); wg.Done() }()
	go func() { io.Copy(b, a); wg.Done() }()
	wg.Wait()
}

func listen(listenAddr, dest string) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("proxy", listenAddr, "->", dest)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("accept", listenAddr, err)
			continue
		}
		go func(c net.Conn) {
			d, err := net.Dial("tcp", dest)
			if err != nil {
				c.Close()
				return
			}
			pipe(c, d)
		}(c)
	}
}

func main() {
	host := "192.168.0.107"
	dest := "172.31.215.1"
	if len(os.Args) >= 3 {
		host = os.Args[1]
		dest = os.Args[2]
	}
	for _, p := range []string{"3000", "3001", "3002", "3003"} {
		go listen(net.JoinHostPort(host, p), net.JoinHostPort(dest, p))
	}
	// Browsers often use IPv4 localhost; WSL only publishes master on ::1:3000.
	go listen("127.0.0.1:3000", net.JoinHostPort(dest, "3000"))
	select {}
}
