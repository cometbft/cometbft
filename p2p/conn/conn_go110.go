package conn

import "net"

func NetPipe() (net.Conn, net.Conn) {
	return net.Pipe()
}
