package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func getSession(r *http.Request) (*sessions.Session, error) {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return nil, err
	}
	return session, nil
}
