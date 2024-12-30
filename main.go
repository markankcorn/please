package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Function to get the operating system
func getOS() string {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		return "Windows (WSL/Ubuntu)" // Assume WSL if on Windows
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return "Unknown"
	}
}

// Function to read Zsh history (updated to handle both Linux and macOS)
func readZshHistory(n int) ([]string, error) {
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		// Default Zsh history file locations
		if runtime.GOOS == "darwin" {
			historyFile = os.Getenv("HOME") + "/.zsh_history"
		} else { // Assume Linux for other OSes
			historyFile = os.Getenv("HOME") + "/.zsh_history"
		}
	}

	file, err := os.Open(historyFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Extract only the command portion from the history lines
	var commands []string
	for _, line := range lines {
		parts := strings.SplitN(line, ";", 2)
		if len(parts) == 2 {
			commands = append(commands, parts[1])
		}
	}

	// Get the last n commands
	startIndex := len(commands) - n
	if startIndex < 0 {
		startIndex = 0
	}

	return commands[startIndex:], nil
}

// build the prompt and send it to Gemini
func getGeminiResponse(ctx context.Context, client *genai.Client, osName string, history []string, prompt string) ([]string, error) {
	// Build the prompt with OS, history, and user request
	var fullPrompt strings.Builder
	fullPrompt.WriteString(fmt.Sprintf("You are an expert at bash command line for %s. ", osName))
	fullPrompt.WriteString("Here is my zsh command history for context:\n")
	for _, h := range history {
		fullPrompt.WriteString(h)
		fullPrompt.WriteString("\n")
	}
	fullPrompt.WriteString("Based on this history and the following request, generate ONLY three bash commands in your response, each on a new line, and nothing else, no other text, that accomplish what is described: ")
	fullPrompt.WriteString(prompt)

	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt.String()))
	if err != nil {
		return nil, err
	}

	// Parse the response to extract the three commands
	var commands []string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		responseContent := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
		lines := strings.Split(responseContent, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				commands = append(commands, trimmedLine)
			}
		}
	} else {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	return commands, nil
}

// execute the command
func executeCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// show some suggestions
func printSuggestions(suggestions []string, selectedIndex int) {
	for i, suggestion := range suggestions {
		if i == selectedIndex {
			fmt.Printf("-> %s\n", suggestion)
		} else {
			fmt.Printf("   %s\n", suggestion)
		}
	}
}

// 
func handleKeyPress(suggestions []string, selectedIndex int) (int, bool) {
	_, key, err := keyboard.GetKey()
	if err != nil {
		log.Fatal(err)
	}

	switch key {
	case keyboard.KeyTab:
		return (selectedIndex + 1) % len(suggestions), false
	case keyboard.KeyEnter:
		executeCommand(suggestions[selectedIndex])
		return selectedIndex, true
	case keyboard.KeyEsc:
		fmt.Println("\nSuggestions rejected.")
		return selectedIndex, true
	default:
		fmt.Println("\nExiting.")
		return selectedIndex, true
	}
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

	osName := getOS()

	suggestions, err := getGeminiResponse(ctx, client, osName, history, userRequest)
	if err != nil {
		log.Fatal("Error getting response from Gemini:", err)
	}

	if len(suggestions) == 0 {
		fmt.Println("No suggestions available.")
		return
	}

	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Println("Suggestions (use Tab to cycle, Enter to accept, Esc to reject):")
	printSuggestions(suggestions, 0)

	selectedIndex := 0
	for {
		selectedIndex, shouldExit := handleKeyPress(suggestions, selectedIndex)
		if shouldExit {
			break
		}

		fmt.Print("\033[H\033[2J") // Clear the screen
		fmt.Println("Suggestions (use Tab to cycle, Enter to accept, Esc to reject):")
		printSuggestions(suggestions, selectedIndex)
	}
}
