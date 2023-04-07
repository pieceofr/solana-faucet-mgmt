package main

import (
	"bufio"
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

type WhitelistEntry struct {
	IP   string
	Memo string
}

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
	ip = strings.TrimSpace(ip)
	memo = strings.ReplaceAll(memo, ";", "")
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format")
	}
	whitelist, err := readWhiteList(config.WhiteListPath)
	if err != nil {
		return fmt.Errorf("error parse whitelist file: %v", err)
	}
	for i, v := range whitelist {
		if v.IP == ip && v.Memo == memo {
			return fmt.Errorf("IP address already in whitelist")
		} else if v.IP == ip && v.Memo != memo {
			whitelist[i] = WhitelistEntry{IP: ip, Memo: memo}
			err = writeWhitelistFile(whitelist, config.WhiteListPath)
			if err != nil {
				return err
			}
			return nil
		}
	}
	whitelist = append(whitelist, WhitelistEntry{IP: ip, Memo: memo})
	err = writeWhitelistFile(whitelist, config.WhiteListPath)
	if err != nil {
		return err
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
