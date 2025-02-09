package main

import "net/http"

type Route interface {
	http.Handler

	Pattern() string
}
