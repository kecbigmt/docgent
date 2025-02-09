package main

import "net/http"

type route interface {
	http.Handler

	Pattern() string
}
