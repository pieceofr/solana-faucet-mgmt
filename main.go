package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	config               Config
	googleOauthConfig    *oauth2.Config
	oauthStateString     = "random_string_1"
	allowedClientDomains []string
)

type Config struct {
	ClientID      string   `json:"client_id"`
	ClientSecret  string   `json:"client_secret"`
	ClientDomains []string `json:"client_domains"`
	WhiteListPath string   `json:"whitelist_path"`
	ServerPort    string   `json:"server_port"`
}

func init() {
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read config file: %v", err))
		os.Exit(1)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to parse config file: %v", err))
		os.Exit(1)
	}
	googleOauthConfig = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  "http://127.0.0.1:8080/auth/google/callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: google.Endpoint,
	}

	allowedClientDomains = config.ClientDomains
	if config.WhiteListPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Errorf("error getting current working directory: %v", err)
			os.Exit(1)
		}
		config.WhiteListPath = wd + "/whitelist.txt"
		fmt.Println("no whitelist path specified, user default path: " + config.WhiteListPath)
	}
	if config.ServerPort == "" {
		config.ServerPort = "8080"
		fmt.Println("no ServerPort path specified, default port: " + "8080")
	}

}

func handleMain(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func handleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func handleOauthCallback(c *gin.Context) {
	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/error?message=Failed+to+exchange")
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/error?message=Failed+to+get+user+info")
		return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read user info:", err)
		c.Redirect(http.StatusTemporaryRedirect, "/error?message=Failed+to+read+user+info")
		return
	}

	var userInfo map[string]interface{}
	json.Unmarshal(contents, &userInfo)

	email := userInfo["email"].(string)
	// err = checkEmailDomain(allowedClientDomains, email)
	// if err != nil {
	// 	c.Redirect(http.StatusTemporaryRedirect, "/error?message=Invalid domain. Please use a solana.com email.")
	// 	return
	// }
	// redirect to islogin page, and add email, name into url's query string.
	redirectURL, err := url.Parse("/faucet_management")
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	query, err := url.ParseQuery(redirectURL.RawQuery)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	query.Add("email", email)
	redirectURL.RawQuery = query.Encode()
	c.Redirect(http.StatusSeeOther, redirectURL.String())
	//c.Redirect(http.StatusTemporaryRedirect, "/faucet_management")
}

func handleError(c *gin.Context) {
	message := c.Query("message")
	c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": message})
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/", handleMain)
	router.GET("/login", handleLogin)
	router.GET("/error", handleError)
	router.GET("/auth/google/callback", handleOauthCallback)
	router.GET("/faucet_management", handleFaucetManagement)
	router.POST("/faucet_management", handleAddToWhiteList)

	router.Run(":" + config.ServerPort)

}
