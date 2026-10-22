package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient() // reads ANTHROPIC_API_KEY from the environment

	prompt := "Explain in two sentences what a goroutine is."

	msg, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.Model("claude-opus-5-5"),
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// The answer comes back as a list of content blocks; print the text ones.
	fmt.Println("Prompt:", prompt)
	fmt.Print("Answer: ")
	for _, block := range msg.Content {
		if block.Type == "text" {
			fmt.Println(block.Text)
		}
	}

	// Token counts for cost calculation.
	fmt.Printf("\n[input tokens: %d, output tokens: %d, stop: %s]\n",
		msg.Usage.InputTokens, msg.Usage.OutputTokens, msg.StopReason)
}
