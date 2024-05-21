package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	claude "github.com/potproject/claude-sdk-go"
)

func main() {
	DEFAULT_CLAUDE_MODEL := "claude-3-haiku-20240307"
	// Define and parse the command-line flags
	inputFile := flag.String("input-file", "", "Path to the input text file")
	inputText := flag.String("input-text", "", "Input text to summarize")
	model := flag.String("model", DEFAULT_CLAUDE_MODEL, "Model to use for the API")
	maxTokens := flag.Int("max-tokens", 200, "Maximum number of tokens in the summary")
	flag.Parse()

	// Check for the ANTHROPIC_API_KEY environment variable
	apiKey, exists := os.LookupEnv("ANTHROPIC_API_KEY")
	if !exists {
		log.Fatal("ANTHROPIC_API_KEY environment variable is not set")
	}

	// Create a Claude client
	client := claude.NewClient(apiKey)

	ctx := context.Background()

	// Define the system prompt
	systemPrompt := `You are a text summarization assistant. 
	Generate a concise summary of the given input text while preserving the key information and main points. 
	Provide the summary in three bullet points, totalling 100 words or less.`

	var userMessage string
	if *inputFile != "" {
		// Read input from file
		content, err := os.ReadFile(*inputFile)
		if err != nil {
			log.Fatalf("Error reading input file: %v\n", err)
		}
		userMessage = string(content)
	} else if *inputText != "" {
		// Use input text from command-line argument
		userMessage = *inputText
	} else {
		log.Fatal("Either input-file or input-text must be provided")
	}

	start := time.Now()

	req := claude.RequestBodyMessages{
		Model:     *model,
		MaxTokens: *maxTokens,
		System:    systemPrompt,
		Messages: []claude.RequestBodyMessagesMessages{
			{
				Role:    claude.MessagesRoleUser,
				Content: userMessage,
			},
		},
		Stream: true, // Enable streaming
	}

	stream, err := client.CreateMessagesStream(ctx, req)
	if err != nil {
		log.Fatalf("ChatCompletion error: %v\n", err)
	}
	defer stream.Close()

	// Stream the summary output to the terminal
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Stream error: %v\n", err)
		}
		fmt.Print(resp.Content[0].Text)
	}

	elapsed := time.Since(start)
	// Print token usage, tokens per second, and total execution time
	fmt.Printf("\n\nTokens generated: %d\n", stream.ResponseBodyMessagesStream.Usage.OutputTokens)
	fmt.Printf("\n\nInput token count: %d\n", stream.ResponseBodyMessagesStream.Usage.InputTokens)

	fmt.Printf("Output tokens per Second: %.2f\n", float64(stream.ResponseBodyMessagesStream.Usage.OutputTokens)/elapsed.Seconds())
	fmt.Printf("Total Execution Time: %s\n", elapsed)
}
