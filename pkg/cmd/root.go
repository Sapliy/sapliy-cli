package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "sapliy",
	Version: "1.0.0",
	Short:   "Sapliy Fintech Ecosystem CLI",
	Long: `Sapliy CLI is the official command line interface for the Sapliy Fintech Ecosystem.
It allows you to manage automation zones, flows, and interact with the event bus.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sapliy.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".sapliy")
	}

	viper.SetEnvPrefix("SAPLIY")
	viper.AutomaticEnv()

	// Explicitly bind environment variables
	viper.BindEnv("api_key", "SAPLIY_API_KEY")
	viper.BindEnv("api_url", "SAPLIY_API_URL")
	viper.BindEnv("org_id", "SAPLIY_ORG_ID")
	viper.BindEnv("current_zone", "SAPLIY_ZONE")

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	if viper.GetBool("verbose") {
		fmt.Printf("DEBUG: api_key='%s', api_url='%s', org_id='%s'\n",
			viper.GetString("api_key"),
			viper.GetString("api_url"),
			viper.GetString("org_id"))
	}
}
