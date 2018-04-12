// +build linux darwin

package tun

import (
	"fmt"
	"log"
	"os"
)

func write(fd *os.File, buf []byte) (int, error) {
        return fd.Write(buf)
}

func read(fd *os.File, mtu int, buf []byte) (int, error) {
	log.Println("mtu = ", mtu)
	if mtu < 1500 {
		mtu = 2048
	}
	
        n, err := fd.Read(buf)

        if n>0 {
                // check length.
                totalLen := 0
                switch buf[0] & 0xf0 {
                case 0x40:
                        totalLen = (int(buf[2])<<8) + int(buf[3])
                case 0x60:
                        totalLen = (int(buf[4])<<8) + int(buf[5]) + 40
                }
                if totalLen != n {
                        return 0, fmt.Errorf("read n(%v)!=total(%v)", n, totalLen)
                }
                return n, nil
        }
        return 0, err
}

func close(fd *os.File) error {
	return fd.Close()
}
