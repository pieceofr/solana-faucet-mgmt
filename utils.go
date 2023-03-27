package main

import (
	"fmt"
	"strings"
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
