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
	rCh := make(chan []byte, 1024)
	wCh := make(chan []byte, 1024)

	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		panic(err)
	}
	// read from udp conn, and write into tun.
	wg.Add(1)
	go func() {
		defer wg.Done()
		buff := make([]byte, 4096)
		for {
			n, err := conn.Read(buff)
			if err != nil {
				panic(err)
			}
			// log.Println("tun<-conn:", n)
			// write into local tun interface channel.
			wCh <- buff[:n]
		}
	}()
	// read from local tun interface channel, and write into remote udp channel.
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			select {
			case data := <-rCh:
				// if data[0]&0xf0 == 0x40 {
				// write into udp conn.
				log.Println("tun->conn:", len(data))
				log.Println("read!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				log.Println("src:", net.IP(data[12:16]), "dst:", net.IP(data[16:20]))
				if _, err := conn.Write(data); err != nil {
					panic(err)
					// }
				}

			}
		}
	}()

	tuntap, err := tun.OpenTunTap(net.IPv4(10, 0, 0, 1), net.IPv4(10, 0, 0, 0), net.IPv4(255, 255, 255, 0))
	if err != nil {
		panic(err)
	}
	defer tuntap.Close()
	// read data from tun into rCh channel.
	wg.Add(1)
	go func() {
		wg.Done()
		if err := tuntap.Read(rCh); err != nil {
			panic(err)
		}
	}()
	// write data into tun from wCh channel.
	wg.Add(1)
	go func() {
		wg.Done()
		if err := tuntap.Write(wCh); err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}
