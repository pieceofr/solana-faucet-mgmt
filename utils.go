package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func checkEmailDomain(domains []string, email string) error {
	// Check if the email address ends with one of the specified domains
	validDomain := false
	for _, domain := range domains {
		if strings.HasSuffix(email, "@"+domain) {
			validDomain = true
			break
		}
	}
	if !validDomain {
		return fmt.Errorf("email address does not end with a valid domain")
	}
	return nil
}

func IsExpired(t time.Time) bool {
	return time.Now().After(t)
}

func TimePlusSeconds(t time.Time, seconds int) time.Time {
	timeInSeconds := time.Duration(seconds) * time.Second
	futureTime := t.Add(timeInSeconds)
	return futureTime
}

func GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(randomBytes)[:length], nil
}

func getURLOfEmailToken(path string, email string, token string) (*url.URL, error) {
	redirectURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	query, err := url.ParseQuery(redirectURL.RawQuery)
	if err != nil {
		return nil, err
	}
	query.Add("email", email)
	query.Add("gtoken", token)
	redirectURL.RawQuery = query.Encode()
	return redirectURL, nil
}
