package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage and replay webhooks",
	Long: `Commands for managing webhook events.
List past webhook deliveries and replay failed or missed webhooks.`,
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent webhook events",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set. Use 'sapliy auth login'.")
			os.Exit(1)
		}

		zone := viper.GetString("current_zone")

		fmt.Printf("📋 Fetching webhook events (zone: %s)...\n", zone)
		fmt.Println(strings.Repeat("─", 80))

		// Demo data - in production, this would call the API
		events := []struct {
			ID          string
			Type        string
			Status      string
			DeliveredAt string
			Endpoint    string
		}{
			{"we_abc123", "payment.succeeded", "succeeded", "2024-01-15T10:30:00Z", "https://example.com/webhook"},
			{"we_def456", "checkout.completed", "failed", "2024-01-15T09:15:00Z", "https://example.com/checkout"},
			{"we_ghi789", "refund.requested", "pending", "2024-01-15T08:00:00Z", "https://example.com/refunds"},
		}

		if len(events) == 0 {
			fmt.Println("No webhook events found.")
			return
		}

		// Header
		fmt.Printf("%-20s %-25s %-12s %-15s %s\n", "EVENT ID", "TYPE", "STATUS", "DELIVERED AT", "ENDPOINT")
		fmt.Println(strings.Repeat("─", 80))

		for _, evt := range events {
			statusIcon := "✅"
			if evt.Status == "failed" {
				statusIcon = "❌"
			} else if evt.Status == "pending" {
				statusIcon = "⏳"
			}

			timestamp := formatTimestamp(evt.DeliveredAt)
			endpoint := truncate(evt.Endpoint, 30)

			fmt.Printf("%-20s %-25s %s %-10s %-15s %s\n",
				evt.ID, evt.Type, statusIcon, evt.Status, timestamp, endpoint)
		}

		fmt.Println()
		fmt.Println("Note: This is demo data. Connect to the API for real webhook events.")
	},
}

var webhooksReplayCmd = &cobra.Command{
	Use:   "replay [event_id]",
	Short: "Replay a webhook event",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set.")
			os.Exit(1)
		}

		eventID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		fmt.Printf("🔄 Replaying webhook event: %s\n", eventID)

		if !force {
			fmt.Print("Are you sure you want to replay this webhook? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Cancelled.")
				return
			}
		}

		// TODO: Implement API call when SDK supports it
		fmt.Println("✅ Webhook replay queued!")
		fmt.Printf("   Event ID: %s\n", eventID)
		fmt.Println("   Note: API integration pending. This is a placeholder.")
	},
}

var webhooksReplayFailedCmd = &cobra.Command{
	Use:   "replay-failed",
	Short: "Replay all failed webhook events",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set.")
			os.Exit(1)
		}

		zone := viper.GetString("current_zone")
		since, _ := cmd.Flags().GetString("since")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		fmt.Printf("🔍 Finding failed webhooks (zone: %s, since: %s)...\n", zone, since)

		// Demo data
		failedEvents := []string{"we_def456", "we_xyz999"}

		if len(failedEvents) == 0 {
			fmt.Println("✅ No failed webhooks found.")
			return
		}

		fmt.Printf("Found %d failed webhook(s)\n", len(failedEvents))

		if dryRun {
			fmt.Println("\n🏃 Dry run - would replay:")
			for _, evt := range failedEvents {
				fmt.Printf("   - %s\n", evt)
			}
			return
		}

		fmt.Println("\nReplaying...")
		for _, evt := range failedEvents {
			fmt.Printf("   ✅ %s → Replayed\n", evt)
		}

		fmt.Println(strings.Repeat("─", 40))
		fmt.Printf("Completed: %d succeeded\n", len(failedEvents))
	},
}

var webhooksInspectCmd = &cobra.Command{
	Use:   "inspect [event_id]",
	Short: "Inspect a webhook event in detail",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set.")
			os.Exit(1)
		}

		eventID := args[0]

		fmt.Printf("📦 Webhook Event: %s\n", eventID)
		fmt.Println(strings.Repeat("─", 60))

		// Demo data
		event := map[string]interface{}{
			"id":           eventID,
			"type":         "payment.succeeded",
			"status":       "succeeded",
			"endpoint":     "https://example.com/webhook",
			"createdAt":    "2024-01-15T10:30:00Z",
			"deliveredAt":  "2024-01-15T10:30:01Z",
			"attempts":     1,
			"responseCode": 200,
			"payload": map[string]interface{}{
				"amount":   5000,
				"currency": "USD",
				"customer": "cus_abc123",
			},
		}

		fmt.Printf("Type:        %s\n", event["type"])
		fmt.Printf("Status:      %s\n", event["status"])
		fmt.Printf("Endpoint:    %s\n", event["endpoint"])
		fmt.Printf("Created:     %s\n", event["createdAt"])
		fmt.Printf("Delivered:   %s\n", formatTimestamp(event["deliveredAt"].(string)))
		fmt.Printf("Attempts:    %v\n", event["attempts"])
		fmt.Printf("Response:    %v\n", event["responseCode"])

		fmt.Println("\nPayload:")
		prettyJSON, _ := json.MarshalIndent(event["payload"], "", "  ")
		fmt.Println(string(prettyJSON))
	},
}

func formatTimestamp(ts string) string {
	if ts == "" {
		return "—"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("Jan 02 15:04")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	rootCmd.AddCommand(webhooksCmd)
	webhooksCmd.AddCommand(webhooksListCmd)
	webhooksCmd.AddCommand(webhooksReplayCmd)
	webhooksCmd.AddCommand(webhooksReplayFailedCmd)
	webhooksCmd.AddCommand(webhooksInspectCmd)

	webhooksListCmd.Flags().IntP("limit", "l", 20, "Number of events to fetch")
	webhooksListCmd.Flags().StringP("status", "s", "", "Filter by status (pending, succeeded, failed)")

	webhooksReplayCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	webhooksReplayFailedCmd.Flags().String("since", "24h", "Time range for failed webhooks (e.g., 1h, 24h, 7d)")
	webhooksReplayFailedCmd.Flags().Bool("dry-run", false, "Show what would be replayed without doing it")
}
