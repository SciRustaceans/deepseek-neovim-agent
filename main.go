package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"deepseek-nvim-agent/commands"
	"deepseek-nvim-agent/config"
	"deepseek-nvim-agent/deepseek"
	"deepseek-nvim-agent/nvim"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	if cfg.DeepSeekAPIKey == "" {
		log.Fatal("DeepSeek API key not found. Set DEEPSEEK_API_KEY environment variable or create ~/.config/deepseek-nvim/config")
	}

	if cfg.NeovimSocket == "" {
		log.Fatal("NVIM_LISTEN_ADDRESS environment variable not set. Neovim must be running with RPC server enabled")
	}

	// Initialize clients
	deepseekClient := deepseek.NewClient(cfg)
	nvimClient, err := nvim.NewRPCClient(cfg.NeovimSocket)
	if err != nil {
		log.Fatalf("Failed to connect to Neovim: %v", err)
	}
	defer nvimClient.Close()

	commandHandler := commands.NewCommandHandler(deepseekClient, nvimClient)

	fmt.Println("DeepSeek Neovim Agent started!")
	fmt.Println("Available commands:")
	fmt.Println("  edit <prompt>    - Edit current code with AI")
	fmt.Println("  explain          - Explain current code")
	fmt.Println("  quit             - Exit the agent")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.SplitN(input, " ", 2)
		command := parts[0]

		switch command {
		case "edit":
			if len(parts) < 2 {
				fmt.Println("Usage: edit <prompt>")
				continue
			}
			prompt := parts[1]
			if err := commandHandler.HandleEditCommand(prompt); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Code edited successfully!")
			}

		case "explain":
			if err := commandHandler.HandleExplainCommand(); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Explanation generated in new buffer!")
			}

		case "quit", "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Available: edit, explain, quit")
		}
	}
}
