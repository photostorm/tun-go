package main

import (
	"fmt"
	"flag"
	"log"
	"net"
	"sync"
	"time"

	"tun-go"
)

func startUDP(addr *net.UDPAddr) {
	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("listen udp failed", addr)
	}
	buf := make([]byte, 2048)
	for {
		n, peer, err := l.ReadFrom(buf)
		if n>20 {
			fmt.Println(peer, n, buf[:20])
		}
		if err!=nil {
			log.Println("read error", err)
		}
	}
}
var laddr string
var raddr string
func init() {
	flag.StringVar(&laddr, "l", "127.0.0.1:37988", "local address")
	flag.StringVar(&raddr, "r", "192.168.8.111:37988", "remote address")
	flag.Parse()
}
// first start server tun server.
func main() {
	wg := sync.WaitGroup{}
	remoteAddr, err := net.ResolveUDPAddr("udp", raddr)
	if err != nil {
		panic(err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", laddr)
	if err != nil {
		panic(err)
	}
	go startUDP(remoteAddr)
	time.Sleep(time.Second)
	// local tun interface read and write channel.
	rBuf := make([]byte, 2048)
	wBuf := make([]byte, 2048)

	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		panic(err)
	}
        
	tuntap, err := tun.OpenTunTap(net.IPv4(10, 0, 0, 1), net.IPv4(10, 0, 0, 0), net.IPv4(255, 255, 255, 0))
	if err != nil {
		panic(err)
	}
	defer tuntap.Close()
        
	// read from udp conn, and write into tun.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			n, err := conn.Read(wBuf)
			// log.Println("tun<-conn:", n)
			// write into local tun interface channel.
			if n>0 {
                                tuntap.Write(wBuf[:n])
                        }
                        if err != nil {
				panic(err)
			}
		}
	}()
	// read from local tun interface channel, and write into remote udp channel.
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			n, err := tuntap.Read(rBuf)
                        if n>0 {
                                log.Println("tun->conn:", n)
                                log.Println("read!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
                                log.Println(rBuf[0]>>4, " src:", net.IP(rBuf[12:16]), "dst:", net.IP(rBuf[16:20]))
                                conn.Write(rBuf[:n])
                        }
                        if err != nil {
                                panic(err)
                        }
		}
	}()

	wg.Wait()
}
