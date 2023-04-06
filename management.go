package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func handleFaucetManagement(c *gin.Context) {
	session := sessions.Default(c)
	session.Set("email", c.Query("email"))
	session.Set("token", c.Query("token"))
	session.Options(sessions.Options{
		MaxAge: config.SessionMaxAge,
	})
	session.Save()
	if session.Get("email") == nil || session.Get("email").(string) == "" ||
		session.Get("token") == nil || session.Get("token").(string) == "" {
		log.Println("Error:invalid session token:")
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "invalid+session"})
		return
	}

	err := IsUserValidate(session.Get("email").(string), session.Get("token").(string))
	if err != nil {
		log.Println("Error:invalid session:", err)
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "invalid+session"})
		return
	}

	wlist, err := ReadWhiteList()
	if err != nil {
		log.Println("Error:invalid whitelist:", err)
		c.Redirect(http.StatusTemporaryRedirect, "/error?msg=error reading whitelist data")
	}

	c.HTML(http.StatusOK, "faucet_management.tmpl", gin.H{"email": c.Query("email"), "IPMemoList": wlist})

}

func handleAddToWhiteList(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") == nil || session.Get("email").(string) == "" ||
		session.Get("token") == nil || session.Get("token").(string) == "" {
		log.Println("Error:invalid session token")
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "invalid+session"})
		return
	}
	email := session.Get("email").(string)
	token := session.Get("token").(string)
	err := IsUserValidate(email, token)
	if err != nil {
		log.Println("Error:invalid session:", err)
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "invalid+session"})
		return
	}

	wlist, err := ReadWhiteList()
	if err != nil {
		log.Println("Error:invalid whitelist:", err)
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "error+reading+whitelist+data"})
		return
	}
	addErr := addIPToWhitelist(c.PostForm("ip"), c.PostForm("memo"), config.WhiteListPath)
	if addErr != nil {
		log.Println("Error:addIPToWhitelist:", err)
		c.HTML(http.StatusNotFound, "faucet_management.tmpl", gin.H{"email": session.Get("email").(string),
			"IPMemoList": wlist, "ErrorMessage": addErr, "SuccessMessage": ""})
		return
	}
	wlist, err = ReadWhiteList()
	if err != nil {
		log.Println("Error:invalid whitelist after add:", err)
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "error+reading+whitelist+data"})
		return
	}
	log.Println("Info:", email, "add IP:", c.PostForm("ip"), " to whitelist successfully")
	c.HTML(200, "faucet_management.tmpl", gin.H{"message": "IP added successfully",
		"email": email, "IPMemoList": wlist, "SuccessMessage": "IP is add", "ErrorMessage": ""})

}

func addIPToWhitelist(ip string, memo string, whitelistPath string) error {
	// Check if the IP address is a valid IPv4 or IPv6 address
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format")
	}
	rmMemo := strings.ReplaceAll(memo, ";", "")
	// Read the current whitelist file
	whitelistData, err := os.ReadFile(whitelistPath)
	if err != nil {
		return fmt.Errorf("error reading whitelist file: %v", err)
	}
	// Check if the IP address is already in the whitelist
	whitelistLines := strings.Split(string(whitelistData), "\n")
	for _, line := range whitelistLines {
		if line == ip {
			log.Println("Warn:IP is already in whitelist")
			return nil
		}
	}

	// Append the IP address to the whitelist file
	f, err := os.OpenFile(whitelistPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening whitelist file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(ip + ";" + rmMemo + "\n"); err != nil {
		return fmt.Errorf("error writing IP address to whitelist file: %v", err)
	}
	updateErr := executeBashScript(config.UpdateUFWPath)
	if updateErr != nil {
		return fmt.Errorf("error execute script: %v", updateErr)
	}
	return nil
}

func executeBashScript(scriptPath string) error {
	// Check if the file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("File does not exist: %s", scriptPath)
	}

	// Execute the script using the "bash" command
	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to execute script: %s", err)
	}

	return nil
}
