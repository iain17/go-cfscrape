package cfscrape

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	r1 = regexp.MustCompile(`name="jschl_vc" value="(\w+)"`)
	r2 = regexp.MustCompile(`name="pass" value="(.+?)"`)

	r3 = regexp.MustCompile(`setTimeout\(function\(\){\s+(var s,t,o,p,b,r,e,a,k,i,n,g,f.+?\r?\n[\s\S]+?a\.value =.+?)\r?\n`)
	r4 = regexp.MustCompile(`a\.value =(.+?) \+ .+?;`)
	r5 = regexp.MustCompile(`\s{3,}[a-z](?: = |\.).+`)

	r6 = regexp.MustCompile(`[\n\\']`)
)

func anyNoMatch(matches ...[][]byte) bool {
	for _, match := range matches {
		if match == nil {
			return true
		}
	}
	return false
}

func extractJsFromPage(body []byte) (string, error) {
	r3m := r3.FindSubmatch(body)

	if anyNoMatch(r3m) {
		return "", ErrCouldNotSolve
	}

	tmp := string(r3m[1])
	tmp = r4.ReplaceAllString(tmp, "$1")
	tmp = r5.ReplaceAllString(tmp, "")

	return r6.ReplaceAllString(tmp, ""), nil
}

func extractPageTokens(body []byte) (p challengeTokens, err error) {
	r1m := r1.FindSubmatch(body)
	r2m := r2.FindSubmatch(body)

	if anyNoMatch(r1m, r2m) {
		err = ErrCouldNotSolve
		return
	}

	p.Challenge = string(r1m[1])
	p.ChallengePass = string(r2m[1])
	return
}

type challengeTokens struct {
	Challenge     string
	ChallengePass string
}

func solveJsAnswer(executor JSExecutor, domain, javascript string) (string, error) {
	if executor == nil {
		return "", errors.New("Invalid executor")
	}

	result, err := executor.ExecuteJS(javascript)
	if err != nil {
		return "", err
	}

	result = strings.TrimSpace(result)

	resultI, err := strconv.Atoi(result)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", resultI+len(domain)), nil
}

func copyStrSlice(in []string) []string {
	r := make([]string, 0, len(in))
	r = append(r, in...)
	return r
}

func copyHeader(header http.Header) http.Header {
	m := make(map[string][]string)

	for k, v := range header {
		m[k] = copyStrSlice(v)
	}

	return m
}

type ChallengeAnswer struct {
	Challenge     string
	ChallengePass string
	JsAnswer      string

	ExecutedJavascript string
}

// Given a hostname, the cloudflare anti bot page, and a javascript executor,
// this returns the "answer" to cloudflare's challenge.
func SolveChallenge(hostname string, body []byte, executor JSExecutor) (*ChallengeAnswer, error) {
	challengeTokens, err := extractPageTokens(body)
	if err != nil {
		return nil, err
	}

	js, err := extractJsFromPage(body)
	if err != nil {
		return nil, err
	}

	jsAnswer, err := solveJsAnswer(executor, hostname, js)
	if err != nil {
		return nil, err
	}

	answer := &ChallengeAnswer{
		Challenge:          challengeTokens.Challenge,
		ChallengePass:      challengeTokens.ChallengePass,
		JsAnswer:           jsAnswer,
		ExecutedJavascript: js,
	}

	return answer, nil
}
