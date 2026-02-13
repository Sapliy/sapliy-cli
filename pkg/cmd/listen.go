package cmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listenCmd = &cobra.Command{
	Use:   "listen [event_pattern]",
	Short: "Start local webhook listener for debugging",
	Long: `Start a local HTTP server to intercept and display webhooks in real-time.
This is useful for debugging flows that send webhooks during development.

Examples:
  sapliy listen                    # Listen to all events
  sapliy listen payment.*          # Listen to payment events only
  sapliy listen --port 3001        # Use custom port`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		secret := viper.GetString("webhook_secret")

		eventPattern := "*"
		if len(args) > 0 {
			eventPattern = args[0]
		}

		green := color.New(color.FgGreen, color.Bold)
		yellow := color.New(color.FgYellow)
		red := color.New(color.FgRed)
		cyan := color.New(color.FgCyan)

		green.Printf("\n🎧 Sapliy Webhook Listener\n")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("Listening on: http://localhost:%d\n", port)
		fmt.Printf("Event filter: %s\n", eventPattern)
		if secret != "" {
			fmt.Printf("Signature verification: %s\n", green.Sprint("ENABLED"))
		} else {
			fmt.Printf("Signature verification: %s\n", yellow.Sprint("DISABLED (set SAPLIY_WEBHOOK_SECRET)"))
		}
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Read body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				red.Printf("❌ Error reading body: %v\n", err)
				http.Error(w, "Error reading body", http.StatusBadRequest)
				return
			}

			// Extract headers
			eventID := r.Header.Get("X-Sapliy-Event-ID")
			if eventID == "" {
				eventID = r.Header.Get("X-Webhook-ID") // Fallback to deprecated header
			}

			eventType := r.Header.Get("X-Sapliy-Event-Type")
			if eventType == "" {
				eventType = r.Header.Get("X-Webhook-Event") // Fallback
			}

			timestamp := r.Header.Get("X-Sapliy-Timestamp")
			if timestamp == "" {
				timestamp = r.Header.Get("X-Webhook-Timestamp") // Fallback
			}

			signature := r.Header.Get("X-Sapliy-Signature")
			if signature == "" {
				signature = r.Header.Get("X-Webhook-Signature") // Fallback
			}

			// Filter by event pattern
			if eventPattern != "*" && !matchPattern(eventType, eventPattern) {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Display webhook
			fmt.Println()
			cyan.Printf("📨 Incoming Webhook\n")
			fmt.Println(strings.Repeat("─", 60))
			fmt.Printf("Event ID:   %s\n", eventID)
			fmt.Printf("Event Type: %s\n", eventType)
			fmt.Printf("Timestamp:  %s\n", timestamp)

			// Verify signature
			if secret != "" && signature != "" {
				h := hmac.New(sha256.New, []byte(secret))
				h.Write(body)
				expectedSig := hex.EncodeToString(h.Sum(nil))

				if signature == expectedSig {
					green.Printf("Signature:  ✓ VALID\n")
				} else {
					red.Printf("Signature:  ✗ INVALID\n")
					red.Printf("  Expected: %s\n", expectedSig)
					red.Printf("  Got:      %s\n", signature)
				}
			} else if signature != "" {
				yellow.Printf("Signature:  %s (not verified)\n", signature)
			}

			fmt.Println()
			fmt.Println("Payload:")
			fmt.Println(strings.Repeat("─", 60))

			// Pretty print JSON
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err == nil {
				prettyJSON, _ := json.MarshalIndent(payload, "", "  ")
				fmt.Println(string(prettyJSON))
			} else {
				fmt.Println(string(body))
			}

			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"received"}`))
		})

		addr := fmt.Sprintf(":%d", port)
		green.Printf("✓ Server started successfully\n")
		fmt.Printf("Press Ctrl+C to stop\n\n")

		if err := http.ListenAndServe(addr, nil); err != nil {
			red.Printf("❌ Failed to start server: %v\n", err)
			os.Exit(1)
		}
	},
}

// matchPattern checks if event type matches the pattern (supports * wildcard)
func matchPattern(eventType, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple wildcard matching: "payment.*" matches "payment.completed"
	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		return strings.HasPrefix(eventType, prefix+".")
	}

	return eventType == pattern
}

func init() {
	rootCmd.AddCommand(listenCmd)
	listenCmd.Flags().IntP("port", "p", 3000, "Port to listen on")
	viper.BindEnv("webhook_secret", "SAPLIY_WEBHOOK_SECRET")
}
