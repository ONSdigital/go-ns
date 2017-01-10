package reverseProxy

import (
	"net/url"
	"net/http"
	"net/http/httputil"
	"net"
	"time"
)

func Create(proxyURL *url.URL, directorFunc func(*http.Request)) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	director := proxy.Director
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = proxyURL.Host
		if directorFunc != nil {
			directorFunc(req)
		}
	}
	return proxy
}
