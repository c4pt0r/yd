package main

// little CLI tool for query english(chinese) word/phrase, :)

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/ngaut/log"

	openai "github.com/sashabaranov/go-openai"
)

var (
	client *openai.Client

	//langDetectPrompt = `What language is the following text? "%s", output should be in {"english", "chinese"}`
	defaultPrompt = `你是一个智能的翻译，如果下面引号内部的内容是英文，那么将下面的英文翻译成中文, 输出中文，并给出音标，英文例句和常用用法; 如果是中文，那么将内容翻译成英文, 不需要音标例句等内容；其他语种，直接翻译成中文即可，也不需要音标例句。要求尽可能详细: "%s"`
	customPrompt  = flag.String("prompt", "", "custom prompt")
)

type Streamer interface {
	// stop steaming by returning io.EOF, or other error
	Recv() (string, error)
	Close()
}

type OpenAIStreamer struct {
	s *openai.ChatCompletionStream
}

func NewOpenAIStreamer(s *openai.ChatCompletionStream) Streamer {
	return &OpenAIStreamer{
		s: s,
	}
}
func (s *OpenAIStreamer) Recv() (string, error) {
	r, err := s.s.Recv()
	if err != nil {
		return "", err
	}
	return r.Choices[0].Delta.Content, nil
}

func (s *OpenAIStreamer) Close() {
	s.s.Close()
}

func init() {
	// read openai token from environment variable
	openaiToken := os.Getenv("OPENAI_API_KEY")
	if openaiToken == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}
	client = openai.NewClient(openaiToken)
}

func query(word string) (Streamer, error) {
	if client == nil {
		panic("client is nil")
	}
	prompt := defaultPrompt
	if *customPrompt != "" {
		prompt = *customPrompt
	}
	fullPrompt := fmt.Sprintf(prompt, word)

	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		Model:     "gpt-4o",
		MaxTokens: 3000,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fullPrompt,
			},
		},
		Stream: true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return nil, err
	}
	return NewOpenAIStreamer(stream), nil
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
		if len(line) == 0 {
			continue
		}
		if stream, err := query(line); err == nil {
			for {
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return err
				}
				fmt.Printf("%s", response)
			}
			fmt.Println()
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
		stream, err := query(word)
		if err != nil {
			log.Fatal(err)
		}
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s", response)
		}
		fmt.Println()
	} else {
		if err := interpreter(); err != nil {
			log.Fatal(err)
		}
	}
}
