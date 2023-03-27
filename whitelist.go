package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type WhiteListRecord struct {
	IP   string
	Memo string
}
type PageData struct {
	List []WhiteListRecord
}

func ReadWhiteList() ([]WhiteListRecord, error) {
	var list []WhiteListRecord
	whitelistData, err := ioutil.ReadFile(config.WhiteListPath)
	if err != nil {
		return nil, fmt.Errorf("error reading IP and memo file: %v", err)
	}
	lines := strings.Split(string(whitelistData), "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, ";")
			if len(parts) != 2 {
				continue
			}
			record := WhiteListRecord{IP: parts[0], Memo: parts[1]}
			list = append(list, record)
		}
	}
	for _, entry := range list {
		fmt.Printf("IP: %s, Memo: %s\n", entry.IP, entry.Memo)
	}
	return list, nil
}
