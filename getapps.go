package main

import (
	"bufio"
	"os"
	"strings"
)

type VPNClient struct {
	OS   string
	Link string
}

func LoadVPNClients(filePath string) ([]VPNClient, error) {
	var clients []VPNClient
	file, err := os.Open(filePath)
	if err != nil {
		return clients, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) == 2 {
			clients = append(clients, VPNClient{OS: parts[0], Link: parts[1]})
		}
	}

	return clients, scanner.Err()
}

func GetClientsByOS(osName string, clients []VPNClient) []string {
	var links []string
	for _, client := range clients {
		if strings.EqualFold(client.OS, osName) {
			links = append(links, client.Link)
		}
	}
	return links
}
