package cmd

import (
	"context"
	"fmt"
	"os"

	fintech "github.com/sapliy/fintech-sdk-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Manage zones",
}

var listZonesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones in an organization",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		orgID := viper.GetString("org_id")
		if apiKey == "" || orgID == "" {
			fmt.Println("Error: Not authenticated or org_id not set. Use 'sapliy auth login'.")
			os.Exit(1)
		}

		client := fintech.NewClient(apiKey)
		zones, err := client.Zones.List(context.Background(), orgID)
		if err != nil {
			fmt.Printf("Error listing zones: %v\n", err)
			return
		}

		fmt.Printf("%-20s %-20s %-10s\n", "ID", "NAME", "MODE")
		for _, z := range zones {
			fmt.Printf("%-20s %-20s %-10s\n", z.ID, z.Name, z.Mode)
		}
	},
}

var createZoneCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new zone",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api_key")
		orgID := viper.GetString("org_id")
		if apiKey == "" || orgID == "" {
			fmt.Println("Error: Not authenticated or org_id not set.")
			os.Exit(1)
		}

		name, _ := cmd.Flags().GetString("name")
		mode, _ := cmd.Flags().GetString("mode")

		client := fintech.NewClient(apiKey)
		z, err := client.Zones.Create(context.Background(), &fintech.CreateZoneRequest{
			OrgID: orgID,
			Name:  name,
			Mode:  mode,
		})
		if err != nil {
			fmt.Printf("Error creating zone: %v\n", err)
			return
		}

		fmt.Printf("Zone created successfully! ID: %s, Mode: %s\n", z.ID, z.Mode)
	},
}

var switchZoneCmd = &cobra.Command{
	Use:   "switch [id]",
	Short: "Switch current zone",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("current_zone", args[0])
		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}
		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}
		fmt.Printf("Switched to zone: %s\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(zonesCmd)
	zonesCmd.AddCommand(listZonesCmd)
	zonesCmd.AddCommand(createZoneCmd)
	zonesCmd.AddCommand(switchZoneCmd)

	createZoneCmd.Flags().StringP("name", "n", "", "Name of the zone")
	createZoneCmd.Flags().StringP("mode", "m", "test", "Mode (test/live)")
	createZoneCmd.MarkFlagRequired("name")
}
