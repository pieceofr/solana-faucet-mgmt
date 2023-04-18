package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
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

	wlist, err := readWhiteList(config.WhiteListPath)
	if err != nil {
		log.Println("Error:invalid whitelist:", err)
		c.Redirect(http.StatusTemporaryRedirect, "/error?msg=error reading whitelist data")
	}

	c.HTML(http.StatusOK, "faucet_management.tmpl", gin.H{"email": c.Query("email"), "IPMemoList": wlist})

}

func handleUpdateWhitelist(c *gin.Context) {
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
	wlist, err := readWhiteList(config.WhiteListPath)
	if err != nil {
		log.Println("Error:invalid whitelist:", err)
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "error+reading+whitelist+data"})
		return
	}
	fmt.Println("add:", c.PostForm("add"), " remove:", c.PostForm("remove"), "openany", c.PostForm("openany"))
	// Determine a remove or an add action
	if c.PostForm("remove") == "remove" && len(c.PostForm("add")) == 0 { // Remove Action
		ip := strings.TrimSpace(c.PostForm("ip"))
		log.Println(email, "try to remmove an ip:", ip, " ----")
		if net.ParseIP(ip) == nil {
			c.HTML(http.StatusBadRequest, "faucet_management.tmpl", gin.H{"email": session.Get("email").(string),
				"IPMemoList": wlist, "ErrorMessage": "invalid IP", "SuccessMessage": ""})
			return
		}
		newEntry := WhitelistEntry{IP: ip, Memo: ""}
		rmErr := removeIPFromWhitelist(newEntry, wlist)
		if rmErr != nil {
			log.Println("Error:addIPToWhitelist:", err)
			c.HTML(http.StatusInternalServerError, "faucet_management.tmpl",
				gin.H{"email": session.Get("email").(string), "IPMemoList": wlist, "ErrorMessage": rmErr, "SuccessMessage": ""})
			return
		}
	} else if c.PostForm("add") == "add" && len(c.PostForm("remove")) == 0 { // Add Action
		ip := strings.TrimSpace(c.PostForm("ip"))
		memo := strings.ReplaceAll(c.PostForm("memo"), ";", "")
		log.Println(email, "try to add an ip:", ip, " ----")
		if net.ParseIP(ip) == nil {
			c.HTML(http.StatusBadRequest, "faucet_management.tmpl", gin.H{"email": session.Get("email").(string),
				"IPMemoList": wlist, "ErrorMessage": "invalid IP", "SuccessMessage": ""})
			return
		}
		newEntry := WhitelistEntry{IP: ip, Memo: memo}
		addErr := updateIPToWhitelist(newEntry, wlist)
		if addErr != nil {
			log.Println("Error:addIPToWhitelist:", err)
			c.HTML(http.StatusInternalServerError, "faucet_management.tmpl", gin.H{"email": session.Get("email").(string),
				"IPMemoList": wlist, "ErrorMessage": addErr, "SuccessMessage": ""})
			return
		}
	} else if c.PostForm("wideopen") == "enable-wide-open" {
		log.Println(email, "try to enable wide open ----")
		addErr := wideOpen()
		if addErr != nil {
			log.Println("Error:enable wide open:", err)
			c.HTML(http.StatusInternalServerError, "faucet_management.tmpl", gin.H{"email": session.Get("email").(string),
				"IPMemoList": wlist, "ErrorMessage": addErr, "SuccessMessage": ""})
			return
		}
	} else if c.PostForm("closewideopen") == "disable-wide-open" {
		log.Println(email, "try to diable wide open ----")
		addErr := closeWideOpen()
		if addErr != nil {
			log.Println("Error:close wide open:", err)
			c.HTML(http.StatusInternalServerError, "faucet_management.tmpl", gin.H{"email": session.Get("email").(string),
				"IPMemoList": wlist, "ErrorMessage": addErr, "SuccessMessage": ""})
			return
		}
	} else {
		log.Println("Error:invalid form submition:submit name mismatch")
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "error+mismatch+form+submition"})
		return
	}

	wlist, err = readWhiteList(config.WhiteListPath)
	if err != nil {
		log.Println("Error:invalid whitelist after add:", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "error+reading+whitelist+data"})
		return
	}
	log.Println("Info:", email, "update whitlist with IP:", c.PostForm("ip"), " successfully")
	c.HTML(200, "faucet_management.tmpl", gin.H{"message": "update whitelist  successfully",
		"email": email, "IPMemoList": wlist, "SuccessMessage": "whitelist is updated", "ErrorMessage": ""})
}
