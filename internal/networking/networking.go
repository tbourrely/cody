package networking

import (
	"fmt"
	"math/rand"
	"net"
)

func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}

	defer ln.Close()

	return true
}

func FindRandomPort() (port int, err error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func random(start int, end int) int {
	return rand.Intn(end-start) + start
}

func FindRandomPortInRange(start int, end int) (port int, err error) {
	port = random(start, end)
	for IsPortAvailable(port) == false {
		port = random(start, end)
	}
	return
}
