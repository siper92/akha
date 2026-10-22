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

	stream := client.Messages.NewStreaming(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.Model("claude-opus-5-5"),
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})

	fmt.Println("Prompt:", prompt)
	fmt.Print("Answer: ")

	msg := anthropic.Message{} // rebuilt from the events, for usage at the end
	for stream.Next() {
		event := stream.Current()
		if err := msg.Accumulate(event); err != nil {
			log.Fatal(err)
		}

		// Print each text chunk as soon as it arrives.
		if ev, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			if d, ok := ev.Delta.AsAny().(anthropic.TextDelta); ok {
				fmt.Print(d.Text)
			}
		}
	}
	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}

	// Final token counts for cost calculation.
	fmt.Printf("\n\n[input tokens: %d, output tokens: %d, stop: %s]\n",
		msg.Usage.InputTokens, msg.Usage.OutputTokens, msg.StopReason)
}
