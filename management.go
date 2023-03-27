package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func handleFaucetManagement(c *gin.Context) {
	// Check token
	gToken := c.Query("token")
	fmt.Println("gtoken:", gToken)
	fmt.Println("email:", c.Query("email"))
	wlist, err := ReadWhiteList()
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/error?msg=error reading whitelist data")
	}
	email := c.Query("email")
	c.HTML(http.StatusOK, "faucet_management.tmpl", gin.H{"email": email, "IPMemoList": wlist, "token": gToken})

}
func handleAddToWhiteList(c *gin.Context) {
	var wl WhiteListRecord
	wl.IP = c.PostForm("ip")
	wl.Memo = c.PostForm("memo")
	err := appendIPToFile(wl, config.WhiteListPath)
	wlist, err := ReadWhiteList()
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/error?msg=error reading whitelist data")
	}
	email := c.Query("email")
	if err != nil {
		c.HTML(500, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(200, "faucet_management.tmpl", gin.H{"message": "IP added successfully", "email": email, "IPMemoList": wlist, "token": c.Query("token")})
}
func appendIPToFile(wl WhiteListRecord, filePath string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(wl.IP + ";" + wl.Memo + "\n")
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Printf("IP %s with memo '%s' added to file\n", wl.IP, wl.Memo)
	return nil
}

func addIPToWhitelist(ip string, whitelistPath string) error {
	// Check if the IP address is a valid IPv4 or IPv6 address
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format")
	}

	// Read the current whitelist file
	whitelistData, err := os.ReadFile(whitelistPath)
	if err != nil {
		return fmt.Errorf("error reading whitelist file: %v", err)
	}

	// Check if the IP address is already in the whitelist
	whitelistLines := strings.Split(string(whitelistData), "\n")
	for _, line := range whitelistLines {
		if line == ip {
			fmt.Println("IP address already exists in the whitelist")
			return nil
		}
	}

	// Append the IP address to the whitelist file
	f, err := os.OpenFile(whitelistPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening whitelist file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(ip + "\n"); err != nil {
		return fmt.Errorf("error writing IP address to whitelist file: %v", err)
	}

	fmt.Println("IP address added to whitelist")
	return nil
}
