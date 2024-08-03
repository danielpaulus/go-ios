package auth_test

import (
	"log"
	"testing"

	"github.com/danielpaulus/go-ios/agent/auth"
)

func TestAuth0(t *testing.T) {
	account, err := auth.AuthorizeUser()
	if err != nil {
		t.Errorf("error authorizing user: %v", err)
		return
	}
	log.Printf("%+v", account)
}
