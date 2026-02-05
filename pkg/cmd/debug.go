package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug and inspect flows in real-time",
	Long: `Debug command provides real-time inspection of flows and events.
Connect to the Sapliy event stream to monitor automation flows as they execute.`,
}

var debugListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen to real-time event stream via HTTP polling",
	Long: `Connect to Sapliy API and poll for events in real-time.
This is useful for debugging flows and watching events as they happen.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set. Use 'sapliy auth login'.")
			os.Exit(1)
		}

		zone := viper.GetString("current_zone")
		if zone == "" {
			zone, _ = cmd.Flags().GetString("zone")
		}

		apiURL := viper.GetString("api_url")
		if apiURL == "" {
			apiURL = "https://api.sapliy.io"
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		filterType, _ := cmd.Flags().GetString("filter")

		fmt.Printf("🔌 Connecting to %s...\n", apiURL)
		fmt.Println("✅ Connected! Polling for events... (Ctrl+C to stop)")
		fmt.Println(strings.Repeat("─", 60))

		// Handle graceful shutdown
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		client := &http.Client{Timeout: 30 * time.Second}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastEventID string

		for {
			select {
			case <-interrupt:
				fmt.Println("\n👋 Disconnecting...")
				return
			case <-ticker.C:
				events := pollEvents(client, apiURL, apiKey, zone, lastEventID)
				for _, event := range events {
					eventType, _ := event["type"].(string)
					eventID, _ := event["id"].(string)

					// Apply filter if specified
					if filterType != "" && !strings.Contains(eventType, filterType) {
						continue
					}

					lastEventID = eventID
					timestamp := time.Now().Format("15:04:05")

					if verbose {
						prettyJSON, _ := json.MarshalIndent(event, "", "  ")
						fmt.Printf("[%s] %s\n%s\n\n", timestamp, eventType, string(prettyJSON))
					} else {
						fmt.Printf("[%s] %-30s  %s\n", timestamp, eventType, eventID)
					}
				}
			}
		}
	},
}

// pollEvents fetches events from the API
func pollEvents(client *http.Client, baseURL, apiKey, zone, afterID string) []map[string]interface{} {
	url := fmt.Sprintf("%s/v1/events?limit=10", baseURL)
	if zone != "" {
		url += "&zone=" + zone
	}
	if afterID != "" {
		url += "&after=" + afterID
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data
}

var debugInspectCmd = &cobra.Command{
	Use:   "inspect [flow_id]",
	Short: "Inspect a specific flow execution",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set.")
			os.Exit(1)
		}

		flowID := args[0]
		fmt.Printf("🔍 Inspecting flow: %s\n", flowID)
		fmt.Println(strings.Repeat("─", 60))

		// TODO: Implement API call to get flow details
		fmt.Println("Flow inspection coming soon...")
	},
}

var debugReplCmd = &cobra.Command{
	Use:   "repl",
	Short: "Interactive REPL for testing events",
	Long: `Start an interactive REPL to test events and flows.
Type event types and JSON data to trigger events interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set.")
			os.Exit(1)
		}

		zone := viper.GetString("current_zone")

		fmt.Println("🎮 Sapliy Debug REPL")
		fmt.Println("Type 'help' for commands, 'exit' to quit")
		fmt.Printf("Current zone: %s\n", zone)
		fmt.Println(strings.Repeat("─", 60))

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("sapliy> ")
			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			switch input {
			case "exit", "quit":
				fmt.Println("👋 Goodbye!")
				return
			case "help":
				fmt.Println(`Commands:
  emit <type> [json]  - Emit an event (e.g., emit payment.created {"amount":100})
  zone <id>           - Switch to a different zone
  status              - Show current configuration
  exit                - Exit the REPL`)
			case "status":
				fmt.Printf("API Key: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])
				fmt.Printf("Zone: %s\n", zone)
				fmt.Printf("API URL: %s\n", viper.GetString("api_url"))
			default:
				if strings.HasPrefix(input, "emit ") {
					parts := strings.SplitN(input[5:], " ", 2)
					eventType := parts[0]
					data := "{}"
					if len(parts) > 1 {
						data = parts[1]
					}
					fmt.Printf("➡️  Emitting %s: %s\n", eventType, data)
					// TODO: Actually emit the event via SDK
				} else if strings.HasPrefix(input, "zone ") {
					zone = strings.TrimSpace(input[5:])
					viper.Set("current_zone", zone)
					fmt.Printf("✅ Switched to zone: %s\n", zone)
				} else {
					fmt.Printf("Unknown command: %s\n", input)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.AddCommand(debugListenCmd)
	debugCmd.AddCommand(debugInspectCmd)
	debugCmd.AddCommand(debugReplCmd)

	debugListenCmd.Flags().StringP("zone", "z", "", "Zone ID to filter events")
	debugListenCmd.Flags().BoolP("verbose", "v", false, "Show full event payloads")
	debugListenCmd.Flags().StringP("filter", "f", "", "Filter events by type (substring match)")
}
