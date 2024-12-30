package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ... (readZshHistory, getGeminiResponse functions from above)

func executeCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GOOGLE_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal("Error creating Gemini client:", err)
	}
	defer client.Close()

	if len(os.Args) < 2 {
		fmt.Println("Usage: please <your request>")
		return
	}

	userRequest := strings.Join(os.Args[1:], " ")

	history, err := readZshHistory(10)
	if err != nil {
		log.Fatal("Error reading Zsh history:", err)
	}

	suggestions, err := getGeminiResponse(ctx, client, history, userRequest)
	if err != nil {
		log.Fatal("Error getting response from Gemini:", err)
	}

	if len(suggestions) == 0 {
		fmt.Println("No suggestions available.")
		return
	}

	// Open keyboard in raw mode
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	// Print initial suggestions
	fmt.Println("Suggestions (use Tab to cycle, Enter to accept, Esc to reject):")
	for i, suggestion := range suggestions {
		if i == 0 {
			fmt.Printf("-> %s\n", suggestion) // Highlight the first suggestion
		} else {
			fmt.Printf("   %s\n", suggestion)
		}
	}

	selectedIndex := 0
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}

		// Handle key presses
		switch key {
		case keyboard.KeyTab:
			// Cycle to the next suggestion
			selectedIndex = (selectedIndex + 1) % len(suggestions)
		case keyboard.KeyEnter:
			// Execute the selected command
			executeCommand(suggestions[selectedIndex])
			return
		case keyboard.KeyEsc:
			// Reject all suggestions and exit
			fmt.Println("\nSuggestions rejected.")
			return
		default:
			// Handle other keys if needed (e.g., for editing a suggestion)
			if char != 0 {
				// In this basic example, any other key press will exit
				fmt.Println("\nExiting.")
				return
			}
		}

		// Clear the previous suggestions
		for range suggestions {
			fmt.Print("\033[A\033[2K") // ANSI escape codes to move up and clear the line
		}

		// Reprint suggestions with the current selection highlighted
		fmt.Println("Suggestions (use Tab to cycle, Enter to accept, Esc to reject):")
		for i, suggestion := range suggestions {
			if i == selectedIndex {
				fmt.Printf("-> %s\n", suggestion)
			} else {
				fmt.Printf("   %s\n", suggestion)
			}
		}
	}
}
