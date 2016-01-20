package cfscrape

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type CookieFilter func([]*http.Cookie) []*http.Cookie

type FilteringJar struct {
	Filter CookieFilter
	*cookiejar.Jar
}

func NewFilteringJar(filter CookieFilter) *FilteringJar {
	cj, _ := cookiejar.New(nil)

	return &FilteringJar{
		Filter: filter,
		Jar:    cj,
	}
}

func (j *FilteringJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	filtered := j.Filter(cookies)
	j.Jar.SetCookies(u, filtered)
}

func CookiesWithName(names ...string) CookieFilter {
	keep := make(map[string]struct{})
	for _, n := range names {
		keep[n] = struct{}{}
	}

	return func(cookies []*http.Cookie) []*http.Cookie {
		res := make([]*http.Cookie, 0, len(cookies))
		for _, c := range cookies {
			if _, exists := keep[c.Name]; exists {
				res = append(res, c)
				continue
			}
		}
		return res
	}
}
