package docker

import (
	"net/http"
	"net"
	"../config"
)

func fakeDial (proto, addr string) (conn net.Conn, err error) {
	return net.Dial("unix", config.Configuration.Docker_ip)
}

func GetClient () *http.Client {
	if config.Configuration.Docker_socket == true {
		tr := &http.Transport{
			Dial: fakeDial,
		}

		client := &http.Client{
			Transport: tr,
		}
		
		return client
	}
	
	return &http.Client{}
}