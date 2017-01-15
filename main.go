package main

// little CLI tool for query english(chinese) word/phrase, :)

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/chzyer/readline"
	"github.com/ngaut/log"
)

const (
	BaseUrl = "https://fanyi.youdao.com/openapi.do?keyfrom=c-dict&key=1416039712&type=data&doctype=json&version=1.1&q="
)

var (
	verbose = flag.Bool("v", false, "verbose")
)

func query(word string) (string, error) {
	url := BaseUrl + url.QueryEscape(word)
	if *verbose {
		log.Info(word)
		log.Info(url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var v map[string]interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return "", err
	}

	if *verbose {
		output, _ := json.MarshalIndent(v, "", "  ")
		log.Info(string(output))
	}

	basic, ok := v["basic"]
	if !ok {
		return "not found", nil
	}

	explains, ok := basic.(map[string]interface{})["explains"]
	if !ok {
		return "not found", nil
	}

	var ret string
	for _, e := range explains.([]interface{}) {
		ret += e.(string) + "\n"
	}

	return strings.TrimSpace(ret), nil
}

func interpreter() error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          ">> ",
		InterruptPrompt: "^C",
	})
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		if ret, err := query(line); err == nil {
			fmt.Println(ret)
		} else {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	word := strings.Join(flag.Args(), " ")
	if len(word) > 0 {
		explain, err := query(word)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(explain)
	} else {
		if err := interpreter(); err != nil {
			log.Fatal(err)
		}
	}
}
