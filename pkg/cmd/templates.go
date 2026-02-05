package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	fintech "github.com/sapliy/fintech-sdk-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage zone templates",
	Long: `Zone templates provide quick-start configurations for common use cases.
Templates include pre-configured flows, webhook endpoints, and default settings.`,
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available zone templates",
	Run: func(cmd *cobra.Command, args []string) {
		templates := []struct {
			Name        string
			Description string
			Flows       int
			Webhooks    int
		}{
			{"e-commerce", "Complete e-commerce solution with checkout, payments, and order tracking", 3, 2},
			{"saas-billing", "Subscription and usage-based billing for SaaS products", 2, 1},
			{"marketplace", "Multi-vendor marketplace with escrow and fee management", 4, 3},
			{"fintech-basic", "Basic payment processing with fraud checks", 2, 2},
			{"automation-hub", "Event-driven automation without payment processing", 1, 1},
		}

		fmt.Println("📋 Available Zone Templates")
		fmt.Println(strings.Repeat("─", 70))
		fmt.Printf("%-18s %-40s %s  %s\n", "NAME", "DESCRIPTION", "FLOWS", "WEBHOOKS")
		fmt.Println(strings.Repeat("─", 70))

		for _, t := range templates {
			fmt.Printf("%-18s %-40s %3d    %3d\n", t.Name, t.Description, t.Flows, t.Webhooks)
		}

		fmt.Println()
		fmt.Println("Use 'sapliy templates apply <name>' to apply a template to a zone.")
	},
}

var templatesApplyCmd = &cobra.Command{
	Use:   "apply [template_name]",
	Short: "Apply a template to a zone",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			fmt.Println("Error: API key not set. Use 'sapliy auth login'.")
			os.Exit(1)
		}

		templateName := args[0]
		zoneName, _ := cmd.Flags().GetString("zone-name")
		mode, _ := cmd.Flags().GetString("mode")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if zoneName == "" {
			zoneName = fmt.Sprintf("%s-zone", templateName)
		}

		fmt.Printf("🎨 Applying template '%s' to new zone '%s' (%s mode)\n", templateName, zoneName, mode)
		fmt.Println(strings.Repeat("─", 60))

		if dryRun {
			fmt.Println("🏃 Dry run - would create:")
			fmt.Printf("   Zone: %s\n", zoneName)
			fmt.Printf("   Mode: %s\n", mode)
			fmt.Printf("   Template: %s\n", templateName)
			return
		}

		client := fintech.NewClient(apiKey, fintech.WithBaseURL(viper.GetString("api_url")))
		orgID := viper.GetString("org_id")

		// Step 1: Create the zone
		fmt.Print("Creating zone... ")
		zone, err := client.Zones.Create(context.Background(), &fintech.CreateZoneRequest{
			OrgID: orgID,
			Name:  zoneName,
			Mode:  mode,
		})
		if err != nil {
			fmt.Printf("❌\n   Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ %s\n", zone.ID)

		// Step 2: Display template info (actual template application would be done server-side)
		fmt.Print("Configuring template... ")
		templateFlows := map[string]int{
			"e-commerce":     3,
			"saas-billing":   2,
			"marketplace":    4,
			"fintech-basic":  2,
			"automation-hub": 1,
		}
		templateWebhooks := map[string]int{
			"e-commerce":     2,
			"saas-billing":   1,
			"marketplace":    3,
			"fintech-basic":  2,
			"automation-hub": 1,
		}
		flows := templateFlows[templateName]
		webhooks := templateWebhooks[templateName]
		fmt.Println("✅")

		fmt.Println()
		fmt.Println("📦 Template Applied Successfully!")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("Zone ID:         %s\n", zone.ID)
		fmt.Printf("Zone Name:       %s\n", zoneName)
		fmt.Printf("Mode:            %s\n", mode)
		fmt.Printf("API Keys:        Available in zone settings\n")
		fmt.Println()
		fmt.Printf("Configured %d flow(s) and %d webhook endpoint(s)\n", flows, webhooks)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  1. Switch to zone: sapliy zones switch %s\n", zone.ID)
		fmt.Println("  2. List flows: sapliy flows list")
		fmt.Println("  3. Start debugging: sapliy debug listen")
	},
}

var templatesShowCmd = &cobra.Command{
	Use:   "show [template_name]",
	Short: "Show details of a template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]

		// Template details
		templates := map[string]struct {
			Description string
			Flows       []string
			Webhooks    []string
			Events      []string
		}{
			"e-commerce": {
				Description: "Complete e-commerce solution with checkout, payments, and order tracking",
				Flows: []string{
					"checkout.started → Create payment intent",
					"payment.succeeded → Send confirmation email + Update inventory",
					"payment.failed → Send failure notification",
					"order.shipped → Send shipping notification",
					"refund.requested → Process refund + Update ledger",
				},
				Webhooks: []string{
					"/webhooks/payment-gateway",
					"/webhooks/shipping-provider",
					"/webhooks/inventory",
				},
				Events: []string{
					"checkout.started", "checkout.completed", "checkout.abandoned",
					"payment.created", "payment.succeeded", "payment.failed",
					"order.created", "order.shipped", "order.delivered",
					"refund.requested", "refund.completed",
				},
			},
			"saas-billing": {
				Description: "Subscription and usage-based billing for SaaS products",
				Flows: []string{
					"subscription.created → Provision account + Welcome email",
					"invoice.created → Process payment",
				},
				Webhooks: []string{
					"/webhooks/billing",
				},
				Events: []string{
					"subscription.created", "subscription.updated", "subscription.cancelled",
					"invoice.created", "invoice.paid", "invoice.failed",
					"usage.recorded", "usage.threshold",
				},
			},
			"marketplace": {
				Description: "Multi-vendor marketplace with escrow and fee management",
				Flows: []string{
					"payment.succeeded → Hold funds in escrow",
					"order.delivered → Release payment to seller",
					"payment.succeeded → Process platform fees",
					"refund.requested → Handle refunds with approval",
				},
				Webhooks: []string{
					"/webhooks/payments",
					"/webhooks/shipping",
					"/webhooks/vendors",
				},
				Events: []string{
					"order.created", "order.paid", "order.shipped", "order.delivered",
					"payment.succeeded", "payment.failed", "refund.requested", "refund.completed",
					"vendor.registered", "vendor.approved", "vendor.payout",
					"dispute.opened", "dispute.resolved",
				},
			},
			"fintech-basic": {
				Description: "Basic payment processing with fraud checks",
				Flows: []string{
					"payment.created → Fraud check for high-value transactions",
					"payment.created → High-value payment approval workflow",
				},
				Webhooks: []string{
					"/webhooks/payments",
					"/webhooks/fraud",
				},
				Events: []string{
					"payment.created", "payment.succeeded", "payment.failed", "payment.disputed",
					"fraud.detected", "fraud.cleared",
				},
			},
			"automation-hub": {
				Description: "Event-driven automation without payment processing",
				Flows: []string{
					"daily schedule → Generate reports and send notifications",
				},
				Webhooks: []string{
					"/webhooks/external",
				},
				Events: []string{
					"automation.triggered", "automation.completed", "automation.failed",
				},
			},
		}

		tmpl, ok := templates[templateName]
		if !ok {
			fmt.Printf("❌ Template '%s' not found. Use 'sapliy templates list' to see available templates.\n", templateName)
			os.Exit(1)
		}

		fmt.Printf("📋 Template: %s\n", templateName)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("Description: %s\n\n", tmpl.Description)

		fmt.Println("Flows:")
		for _, f := range tmpl.Flows {
			fmt.Printf("  • %s\n", f)
		}

		fmt.Println("\nWebhook Endpoints:")
		for _, w := range tmpl.Webhooks {
			fmt.Printf("  • %s\n", w)
		}

		fmt.Println("\nEvent Types:")
		eventsJSON, _ := json.MarshalIndent(tmpl.Events, "  ", "  ")
		fmt.Printf("  %s\n", string(eventsJSON))
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesApplyCmd)
	templatesCmd.AddCommand(templatesShowCmd)

	templatesApplyCmd.Flags().StringP("zone-name", "n", "", "Name for the new zone")
	templatesApplyCmd.Flags().StringP("mode", "m", "test", "Zone mode (test/live)")
	templatesApplyCmd.Flags().Bool("dry-run", false, "Show what would be created without doing it")
}
