package main

import (
	"net/http"
)

func checkRestUser(r *http.Request) bool {
	user, pass, _ := r.BasicAuth()

	userCheck := "admin"
	passCheck := "password"

	if userCheck != user || passCheck != pass {
		return false
	}

	return true
}

func checkHaUser(r *http.Request) bool {
	user, pass, _ := r.BasicAuth()

	userCheck := "admin"
	passCheck := "password"

	if userCheck != user || passCheck != pass {
		return false
	}

	return true
}

func checkBothUsers(r *http.Request) bool {
	if checkHaUser(r) || checkRestUser(r) {
		return true
	}

	return false
}
