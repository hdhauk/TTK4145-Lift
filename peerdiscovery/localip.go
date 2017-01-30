package peerdiscovery

import (
	"net"
	"strings"
)

var localIP string

// GetLocalIP return the IP-adress of the local client. It does this by
// dailing the Google DNS service, hence it will fail if it is unable to reach
// the internet.
func GetLocalIP() (string, error) {
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
