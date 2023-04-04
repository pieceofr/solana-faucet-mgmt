package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oauthStateStringLength = 16

var (
	config               Config
	googleOauthConfig    *oauth2.Config
	oauthStateString     = "random_string_1"
	allowedClientDomains []string
	mongoClient          *mongo.Client
)

type Config struct {
	LogPath              string   `json:"log_path"`
	ClientID             string   `json:"client_id"`
	ClientSecret         string   `json:"client_secret"`
	ClientDomains        []string `json:"client_domains"`
	WhiteListPath        string   `json:"whitelist_path"`
	UpdateUFWPath        string   `json:"update_ufw_script_path"`
	ServerPort           string   `json:"server_port"`
	SessionMaxAge        int      `json:"session_max_age"`
	MongoAddr            string   `json:"mongo_address"`
	MongoUsername        string   `json:"mongo_username"`
	MongoPassword        string   `json:"mongo_password"`
	MongoDB              string   `json:"mongodb"`
	MongoLoginCollection string   `json:"mongo_login_Col"`
	MongoLoginExpireSec  int      `json:"mongo_login_expire_sec"`
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
		RedirectURL:  "http://vipfaucet.solana.com:8080/auth/google/callback",
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
	mongo_init()
	//gin.DisableConsoleColor()
	// // Logging to a file.
	// f, _ := os.Create(config.LogPath)
	// gin.DefaultWriter = io.MultiWriter(f)
	return
}

func handleMain(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func handleLogin(c *gin.Context) {
	newStateString, err := GenerateRandomString(oauthStateStringLength)
	if err != nil {
		c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": "Failed+to+generate+random+state"})
		return
	}
	oauthStateString = newStateString
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func handleOauthCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		fmt.Println("Invalid oauth state")
		c.Redirect(http.StatusTemporaryRedirect, "/error?message=Failed+to+match+oauth+state")
		return
	}
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

	// Save to mongodb for checking
	createTime := time.Now().UTC()
	expired := TimePlusSeconds(createTime, config.MongoLoginExpireSec)
	u := User{
		User:         email,
		Token:        token.AccessToken,
		CreateOn:     createTime,
		LastVerified: createTime,
		ExpiredTime:  expired,
	}

	err = mongoUpdateUser(mongoClient, config.MongoDB, config.MongoLoginCollection, u)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/error?message=Failed+to+add+user")
		return
	}
	redirectURL, err := url.Parse("/faucet_management")
	if err != nil {
		fmt.Println("Parse Failed")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	query, err := url.ParseQuery(redirectURL.RawQuery)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	query.Add("email", email)
	query.Add("token", token.AccessToken)
	redirectURL.RawQuery = query.Encode()
	c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
}

func handleError(c *gin.Context) {
	fmt.Println("***handleError")
	message := c.Query("message")
	c.HTML(http.StatusOK, "error.tmpl", gin.H{"message": message})
}

func main() {
	router := gin.Default()
	store := cookie.NewStore([]byte("auth"))
	router.Use(sessions.Sessions("sol", store))
	router.LoadHTMLGlob("templates/*")
	router.GET("/", handleMain)
	router.GET("/login", handleLogin)
	// router.GET("/error", handleError)
	router.GET("/auth/google/callback", handleOauthCallback)
	router.GET("/faucet_management", handleFaucetManagement)
	router.POST("/faucet_management", handleAddToWhiteList)
	defer func() {
		if mongoClient != nil {
			mongoClient.Disconnect(context.Background())
		}
	}()
	router.Run(":" + config.ServerPort)

}
