package main

import (
	"net/url"
	"strings"
)

func addrToUri(addr string) (*url.URL, error) {
	if strings.Contains(addr, "://") {
		// Looks like a full uri
		return url.Parse(addr)
	}
	return url.Parse("http://" + addr)
}
