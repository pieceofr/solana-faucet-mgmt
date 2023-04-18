package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type WhiteListRecord struct {
	IP   string
	Memo string
}
type WhitelistEntry struct {
	IP   string
	Memo string
}

func readWhiteList(listpath string) ([]WhitelistEntry, error) {
	var whitelist []WhitelistEntry
	f, err := os.Open(listpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ";")
		if len(fields) != 2 {
			continue // Skip lines with incorrect format
		}
		w := WhitelistEntry{
			IP:   fields[0],
			Memo: fields[1],
		}
		whitelist = append(whitelist, w)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return whitelist, nil
}

func writeWhitelistFile(whitelist []WhitelistEntry, listpath string) error {
	f, err := os.Create(listpath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, w := range whitelist {
		line := fmt.Sprintf("%s;%s\n", w.IP, w.Memo)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
func wideOpen() error {
	updateErr := executeBashScript(config.UpdateUFWPath, "any")
	if updateErr != nil {
		return fmt.Errorf("error execute script: %v", updateErr)
	}
	return nil
}
func closeWideOpen() error {
	updateErr := executeBashScript(config.DeleteUFWPath, "any")
	if updateErr != nil {
		return fmt.Errorf("error execute script: %v", updateErr)
	}
	return nil
}
func updateIPToWhitelist(entry WhitelistEntry, whitelist []WhitelistEntry) error {
	for i, v := range whitelist {
		if v.IP == entry.IP && v.Memo == entry.Memo {
			log.Println("Info:IP already in whitelist")
			return nil
		} else if v.IP == entry.IP && v.Memo != entry.Memo {
			whitelist[i] = WhitelistEntry{IP: entry.IP, Memo: entry.Memo}
			err := writeWhitelistFile(whitelist, config.WhiteListPath)
			if err != nil {
				return err
			}
			return nil
		}
	}
	whitelist = append(whitelist, WhitelistEntry{IP: entry.IP, Memo: entry.Memo})
	err := writeWhitelistFile(whitelist, config.WhiteListPath)
	if err != nil {
		return err
	}
	updateErr := executeBashScript(config.UpdateUFWPath, "")
	if updateErr != nil {
		return fmt.Errorf("error execute script: %v", updateErr)
	}
	return nil
}

func removeIPFromWhitelist(entry WhitelistEntry, whitelist []WhitelistEntry) error {
	found := false
	if entry.IP == "127.0.0.1" {
		return fmt.Errorf("can not remove 127.0.0.1")
	}
	for i, v := range whitelist {
		if v.IP == entry.IP {
			whitelist = append(whitelist[:i], whitelist[i+1:]...)
			found = true
		}
	}
	if found {
		err := writeWhitelistFile(whitelist, config.WhiteListPath)
		if err != nil {
			return err
		}
		updateErr := executeBashScript(config.DeleteUFWPath, entry.IP)
		if updateErr != nil {
			return fmt.Errorf("error execute script: %v", updateErr)
		}
	} else {
		log.Println("Info:IP is not in whitelist")
	}
	return nil
}
