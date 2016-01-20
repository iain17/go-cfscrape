// A Golang port(?) of https://github.com/Anorov/cloudflare-scrape
package cfscrape

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	// This error is returned if we could not solve cloudflare's challenge.
	// Most likely, this means Cloudflare changed their "challenge"
	ErrCouldNotSolve = errors.New("Could not solve cloudflare challenge")
)

type cfRoundTripper struct {
	original      http.RoundTripper
	executor      JSExecutor
	cloudflarejar *FilteringJar
}

func (rt cfRoundTripper) originalTrip(req *http.Request) (*http.Response, error) {
	// Modify the request.

	// Yes, golang docs explicitly forbid this behavior for RoundTrippers, but
	// there is no other way to transparently handle *all* requests sent out
	// by an http client to inject cloudflare cookie data.
	if req.UserAgent() == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.111 Safari/537.36")
	}

	for _, cookie := range rt.cloudflarejar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}

	req.Write(os.Stdout)
	resp, err := rt.original.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Record cookie data
	if rc := resp.Cookies(); len(rc) > 0 {
		rt.cloudflarejar.SetCookies(resp.Request.URL, resp.Cookies())
	}

	return resp, err
}

// Fulfills the http.RoundTripper interface
func (rt cfRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.originalTrip(req)
	if err != nil {
		return resp, err
	}

	// Examine response data
	if IsCloudflareChallenge(resp) {
		defer resp.Body.Close()
		req, err := getChallengeSolution(resp, rt.executor)
		if err != nil {
			return nil, err
		}

		time.Sleep(4200 * time.Millisecond)
		return rt.originalTrip(req)
	}
	return resp, err
}

type Options struct {
	// If nil, this uses the default executor, which uses node.js
	Executor JSExecutor

	// If nil, this uses the default http Transport
	RoundTripper http.RoundTripper
}

// Creates a RoundTripper that solves Cloudflare "I Am Under Attack" / "anti-bot" page.
//
// To use this: create an http.Client with a Transport field set to the RoundTripper that this function creates.
//
// A nil value for options will set reasonable defaults for some definition of reasonable.
//
// But in the spirit of Go, where it is *actually* encouraged for each package to define their own set of random global variables,
// you may find an existing http.Client defined already as DefaultClient.
func NewRoundTripper(options *Options) http.RoundTripper {
	jar := NewFilteringJar(CookiesWithName("__cfduid", "cf_clearance"))

	var executor JSExecutor
	var original http.RoundTripper

	if options != nil && options.Executor != nil {
		executor = options.Executor
	} else {
		executor = NodeExecutor
	}

	if options != nil && options.RoundTripper != nil {
		original = options.RoundTripper
	} else {
		original = http.DefaultTransport
	}

	return &cfRoundTripper{
		original:      original,
		executor:      executor,
		cloudflarejar: jar,
	}
}

func makeSolutionRequest(cfResp *http.Response, answer *ChallengeAnswer) (*http.Request, error) {
	url, err := url.Parse(fmt.Sprintf("%s://%s/cdn-cgi/l/chk_jschl", cfResp.Request.URL.Scheme, cfResp.Request.URL.Host))
	if err != nil {
		return nil, err
	}

	v := url.Query()
	v.Set("jschl_vc", answer.Challenge)
	v.Set("jschl_answer", answer.JsAnswer)
	v.Set("pass", answer.ChallengePass)
	url.RawQuery = v.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header = copyHeader(cfResp.Request.Header)
	req.Header.Set("Referer", cfResp.Request.URL.String())

	return req, nil
}

func getChallengeSolution(resp *http.Response, executor JSExecutor) (*http.Request, error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	answer, err := SolveChallenge(resp.Request.URL.Host, body, executor)
	if err != nil {
		return nil, err
	}

	return makeSolutionRequest(resp, answer)
}

func IsCloudflareChallenge(resp *http.Response) bool {
	refresh := resp.Header.Get("Refresh")
	server := resp.Header.Get("Server")

	return resp.StatusCode == http.StatusServiceUnavailable &&
		strings.Contains(refresh, "URL=/cdn-cgi/") &&
		server == "cloudflare-nginx"
}

var DefaultClient = &http.Client{Transport: NewRoundTripper(nil)}

func Get(url string) (resp *http.Response, err error) {
	return DefaultClient.Get(url)
}
