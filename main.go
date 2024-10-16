package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
)

const (
	// Url          = "http://127.0.0.1:8000/onboard/system/info"  // for local testing
	Url           = "https://unode-backend.techmentor.solutions/onboard/system/info"
	InvalidToken  = "invalid-token"
	ScriptVersion = "1.2"
)

var (
	ErrEmptyMACAddress = errors.New("MAC Address is empty")
	ErrEmptyIPAddress  = errors.New("IP Address is empty")
)

type SystemInfo struct {
	MACAddress    string `json:"macAddress"`
	IPAddress     string `json:"ipAddress"`
	OS            string `json:"os"`
	Token         string `json:"token"`
	ScriptVersion string `json:"scriptVersion"`
}

func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Failed getting system's network interfaces. Error: ", err)
		return "", err
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Failed getting a list of unicast interface addresses for a specific network interface. Error: ", err)
			return "", err
		}

		if len(addrs) > 0 {
			return iface.HardwareAddr.String(), nil
		}

		fmt.Println("List of unicast interface addresses for a all network interface is empty")
	}

	return "", ErrEmptyMACAddress
}

func getIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Failed to get a list of the system's unicast interface addresses", err)
		return "", err
	}

	for _, addr := range addrs {

		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}

	}

	return "", ErrEmptyIPAddress
}

func sendToServer(info SystemInfo, url string) error {

	jsonData, err := json.Marshal(info)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}
	_, err = http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Println("Error connecting to server")
		return err
	}

	return nil
}

func main() {
	macAddr, err := getMACAddress()
	if err != nil {
		fmt.Println("Error getting MAC address. Error: ", err)
		return
	}

	ipAddr, err := getIPAddress()
	if err != nil {
		fmt.Println("Error getting IP address. Error: ", err)
		return
	}

	token := flag.String("token", InvalidToken, "unique token given while adding new node")
	flag.Parse()

	tokenStr := strings.TrimSpace(*token)
	if tokenStr == InvalidToken || tokenStr == "" { // comparing it to InvalidToken because its the default value
		fmt.Println("\"-token\" flag can not be empty")
		return
	}

	info := SystemInfo{
		MACAddress:    macAddr,
		IPAddress:     ipAddr,
		OS:            runtime.GOOS,
		Token:         tokenStr,
		ScriptVersion: ScriptVersion,
	}

	err = sendToServer(info, Url)
	if err != nil {
		return
	}

	fmt.Printf("System info %+v", info)
}
