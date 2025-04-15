package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vulnebify/tunnelier/internal"
)

var (
	mongoURL        string
	mongoDB         string
	mongoCollection string
)

func getMongoURL() string {
	if mongoURL != "" {
		return mongoURL
	}
	return os.Getenv("TUNNELIER_MONGO_URL")
}

func ensureWireGuardInstalled() error {
	_, err := exec.LookPath("wg-quick")
	if err != nil {
		return fmt.Errorf("❌ WireGuard is not installed (missing 'wg-quick')\n💡 Please install it: https://www.wireguard.com/install/")
	}
	return nil
}

const asciiArt = `
 _____  _   _  _   _  _   _  _____  _      ___  _____  ____  
|_   _|| | | || \ | || \ | || ____|| |    |_ _|| ____||  _ \ 
  | |  | | | ||  \| ||  \| ||  _|  | |     | | |  _|  | |_) |
  | |  | |_| || |\  || |\  || |___ | |___  | | | |___ |  _ < 
  |_|   \___/ |_| \_||_| \_||_____||_____||___||_____||_| \_\
                                                             	
`

var rootCmd = &cobra.Command{
	Use:     "tunnelier",
	Short:   "Tunnelier is a VPN manager for WireGuard configs stored in MongoDB.",
	Version: app.Version,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(asciiArt)
		_ = cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}
