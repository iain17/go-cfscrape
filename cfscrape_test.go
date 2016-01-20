package cfscrape

import (
	"io/ioutil"
	"testing"
)

func TestSolveJS(t *testing.T) {
	tests := []struct {
		File     string
		Domain   string
		Expected string
	}{
		{"./testdata/kissmanga1.html", "kissmanga.com", "-1431"},
		{"./testdata/kissmanga2.html", "kissmanga.com", "94"},
	}

	executor := JSExecutorFunc(NodeExecute)

	for _, test := range tests {
		body, err := ioutil.ReadFile(test.File)
		if err != nil {
			t.Error(err)
		}
		js, err := extractJsFromPage(body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("JS for %s: %s\n", test.File, js)

		answer, err := solveJsAnswer(executor, test.Domain, js)
		if err != nil {
			t.Error(err)
		}

		t.Logf("JS answer: %s\n", answer)
		t.Logf("Expected: %s\n", test.Expected)

		if test.Expected != answer {
			t.Fail()
		}
	}
}
