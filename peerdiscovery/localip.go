package peerdiscovery

import (
	"net"
	"strings"
	"sync"
)

var mu sync.Mutex
var localIP string

// GetLocalIP return the IP-address of the local client. It does this by
// dialing the Google DNS service, hence it will fail if it is unable to reach
// the internet.
func GetLocalIP() (string, error) {
	mu.Lock()
	defer mu.Unlock()
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}
