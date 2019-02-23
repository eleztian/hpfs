package p2p

import (
	"testing"
	"net"
	"fmt"
)

func TestName(t *testing.T) {
	conn, err := net.Dial("udp", "google.com:80")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer conn.Close()
	fmt.Println(conn.LocalAddr().String())
}
