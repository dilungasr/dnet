package dnet

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

// GetIP returns the IP address from the client request
func GetIP(r *http.Request) (ip string, err error) {
	// get ip from the X-REAL-IP
	ip = r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	devLogger("X-REAL-IP:", ip)
	if netIP != nil {
		return ip, nil
	}

	//get ip from the X-FORWARDED-FOR
	ips := r.Header.Get("X-FORWARDED-FOR")
	ipsSlice := strings.Split(ips, ",")

	// parse the ip to check if the ip is assigned
	for _, ip := range ipsSlice {
		netIP = net.ParseIP(ip)

		//  return the ip address if ip is valid
		if netIP != nil {
			return ip, nil
		}
	}

	// get ip address from the RemoteAddr
	ip, _, err = net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	log.Println("X-FORWARDED-FOR:", netIP)

	// if no valid ip found
	return "", fmt.Errorf("[dnet] no valid IP Address found ")
}
