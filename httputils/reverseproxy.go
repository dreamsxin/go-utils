package httputils

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// AsURL expands :port and hostname to http://localhost:port & http://hostname respectively
func AsURL(s string) string {
	if isPort(s) {
		s = "localhost" + s
	}
	// google.com will be parsed as URL{Path: google.com} without an explicit protocol
	// hence the hack
	if !strings.Contains(s, "://") {
		s = "http://" + s
	}
	return s
}

func isPort(s string) bool {
	match, _ := regexp.MatchString(`^:\d{1,5}$`, s)
	return match
}

// ReverseProxy
//
//   - upstream: http://user:pass@example.com
//     behavior: use upstream credential by default
//   - upstream: http://example.com
//     behavior: passthrough client credential if any
//   - upstream: http://-@example.com
//     behavior: don't pass any credential to upstream
func ReverseProxy(addr string) http.Handler {
	addr = AsURL(addr)
	upstream, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	rewrite := func(r *httputil.ProxyRequest) {
		r.SetURL(upstream)
		r.SetXForwarded()

		// passthrough Host from client
		if os.Getenv("PASS_HOST") != "" {
			r.Out.Host = r.In.Host
		}

		// ignore credential when default upstream user set to -
		if upstream.User.String() == "-" {
			r.Out.Header.Del("Authorization")
			return
		}

		// use client credentials if available
		if r.In.Header.Get("Authorization") != "" {
			return
		}

		// otherwise use credentials from upstream
		if upstream.User != nil {
			user := upstream.User.Username()
			pass, _ := upstream.User.Password()
			r.Out.SetBasicAuth(user, pass)
		}
	}
	modify := func(r *http.Response) error {
		r.Header.Del("Content-Security-Policy")
		return nil
	}
	rp := &httputil.ReverseProxy{
		Rewrite:        rewrite,
		ModifyResponse: modify,
		ErrorLog:       ReverseProxyLogger(),
	}

	return rp
}

// TransparentProxy is a reverse proxy that preserves the original Host header
func TransparentProxy(addr string) http.Handler {
	addr = AsURL(addr)
	upstream, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	rewrite := func(r *httputil.ProxyRequest) {
		r.SetURL(upstream)
		r.SetXForwarded()

		// passthrough Host from client
		r.Out.Host = r.In.Host
	}
	rp := &httputil.ReverseProxy{
		Rewrite:  rewrite,
		ErrorLog: ReverseProxyLogger(),
	}

	return rp
}

func ReverseProxyLogger() *log.Logger {
	if os.Getenv("REVERSEPROXY_LOG") == "" {
		return log.New(io.Discard, "", 0) // discard logger
	}
	return nil // default logger
}

func LoggedReverseProxy(rt http.RoundTripper) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport: rt,
		ErrorLog:  ReverseProxyLogger(),
	}
}
