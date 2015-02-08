package main

// little CLI tool for query english(chinese) word/phrase, :)

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	BaseUrl = "https://fanyi.youdao.com/openapi.do?keyfrom=c-dict&key=1416039712&type=data&doctype=json&version=1.1&q="
)

var (
	verbose = flag.Bool("v", false, "verbose")
)

func main() {
	flag.Parse()
	word := strings.Join(flag.Args(), " ")
	url := BaseUrl + url.QueryEscape(word)
	if *verbose {
		log.Println(word)
		log.Println(url)
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var v map[string]interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		log.Fatal(err)
	}

	if *verbose {
		output, _ := json.MarshalIndent(v, "", "  ")
		log.Println(string(output))
	}

	basic, ok := v["basic"]
	if !ok {
		os.Exit(0)
	}

	explains, ok := basic.(map[string]interface{})["explains"]
	if !ok {
		os.Exit(0)
	}

	for _, e := range explains.([]interface{}) {
		fmt.Println(e.(string))
	}
}
