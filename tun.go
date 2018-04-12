package tun

const (
	IPv6_HEADER_LENGTH = 40
)

type Tun interface {
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	Close() error
}
