package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

type Pricing struct {
	Input, Output, CacheWrite, CacheRead float64
}

var prices = Pricing{
	Input:      0, // e.g. 5.00
	Output:     0, // e.g. 25.00
	CacheWrite: 0, // usually 1.25x input (5-min cache)
	CacheRead:  0, // usually 0.1x input
}

func callCost(u anthropic.Usage, p Pricing) float64 {
	return (float64(u.InputTokens)*p.Input +
		float64(u.OutputTokens)*p.Output +
		float64(u.CacheCreationInputTokens)*p.CacheWrite +
		float64(u.CacheReadInputTokens)*p.CacheRead) / 1_000_000
}

// Your actual tool implementation.
func runTool(name string, input json.RawMessage) string {
	switch name {
	case "get_weather":
		var args struct {
			Location string `json:"location"`
		}
		_ = json.Unmarshal(input, &args)
		return fmt.Sprintf(`{"location":%q,"temp_c":18,"conditions":"cloudy"}`, args.Location)
	}
	return `{"error":"unknown tool"}`
}

func main() {
	ctx := context.Background()
	client := anthropic.NewClient() // reads ANTHROPIC_API_KEY

	tools := []anthropic.ToolUnionParam{{
		OfTool: &anthropic.ToolParam{
			Name:        "get_weather",
			Description: anthropic.String("Get current weather for a city"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"location": map[string]any{"type": "string"},
				},
				Required: []string{"location"},
			},
		},
	}}

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock("What's the weather in Sofia?")),
	}

	var total anthropic.Usage
	var totalCost float64

	// Agentic loop: each iteration is one API call with its own usage.
	for turn := 1; ; turn++ {
		stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model("claude-opus-5-5"),
			MaxTokens: 1024,
			Tools:     tools,
			Messages:  messages,
		})

		// Accumulate builds the full Message (content blocks + final usage)
		// from the events, same shape as the non-streaming response.
		msg := anthropic.Message{}
		for stream.Next() {
			event := stream.Current()
			if err := msg.Accumulate(event); err != nil {
				log.Fatal(err)
			}

			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockStartEvent:
				if ev.ContentBlock.Type == "tool_use" {
					fmt.Printf("\n[calling tool %s] ", ev.ContentBlock.Name)
				}
			case anthropic.ContentBlockDeltaEvent:
				switch d := ev.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					fmt.Print(d.Text) // live text
				case anthropic.InputJSONDelta:
					fmt.Print(d.PartialJSON) // live tool args (partial JSON — don't parse yet)
				}
			}
		}
		if err := stream.Err(); err != nil {
			log.Fatal(err)
		}

		// Cost for this call. msg.Usage is final here (message_delta counts are
		// cumulative; Accumulate keeps the latest, so no manual summing).
		c := callCost(msg.Usage, prices)
		totalCost += c
		total.InputTokens += msg.Usage.InputTokens
		total.OutputTokens += msg.Usage.OutputTokens
		total.CacheCreationInputTokens += msg.Usage.CacheCreationInputTokens
		total.CacheReadInputTokens += msg.Usage.CacheReadInputTokens
		fmt.Printf("\n[turn %d: in=%d out=%d cache_w=%d cache_r=%d cost=$%.6f]\n",
			turn, msg.Usage.InputTokens, msg.Usage.OutputTokens,
			msg.Usage.CacheCreationInputTokens, msg.Usage.CacheReadInputTokens, c)

		messages = append(messages, msg.ToParam())

		if msg.StopReason != anthropic.StopReasonToolUse {
			break
		}

		// Run every tool_use block and send the results back in one user message.
		var results []anthropic.ContentBlockParamUnion
		for _, block := range msg.Content {
			if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				out := runTool(tu.Name, tu.Input)
				results = append(results, anthropic.NewToolResultBlock(tu.ID, out, false))
			}
		}
		messages = append(messages, anthropic.NewUserMessage(results...))
	}

	fmt.Printf("\nTOTAL: in=%d out=%d cache_w=%d cache_r=%d cost=$%.6f\n",
		total.InputTokens, total.OutputTokens,
		total.CacheCreationInputTokens, total.CacheReadInputTokens, totalCost)
}
