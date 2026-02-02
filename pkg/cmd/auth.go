package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with your API key",
	Run: func(cmd *cobra.Command, args []string) {
		var apiKey string
		fmt.Print("Enter API Key: ")
		fmt.Scanln(&apiKey)

		viper.Set("api_key", apiKey)
		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}

		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Println("Successfully authenticated!")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
}
