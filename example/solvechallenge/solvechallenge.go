package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/iain17/go-cfscrape"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: solvechallenge url page")
		os.Exit(1)
	}

	url, err := url.Parse(os.Args[1])
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}

	answer, err := cfscrape.SolveChallenge(url.Host, body, cfscrape.NodeExecutor)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", answer)
}
