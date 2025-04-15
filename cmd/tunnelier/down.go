package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down <vpn-name>",
	Short: "Bring down a WireGuard VPN connection",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return ensureWireGuardInstalled()
	},
	Run: func(cmd *cobra.Command, args []string) {
		vpnName := args[0]
		tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.conf", vpnName))

		if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
			fmt.Printf("❌ Temp config file not found: %s\n", tmpPath)
			os.Exit(1)
		}

		fmt.Printf("❌ Bringing down VPN: %s\n", vpnName)

		downCmd := exec.Command("wg-quick", "down", tmpPath)
		downCmd.Stdout = os.Stdout
		downCmd.Stderr = os.Stderr

		if err := downCmd.Run(); err != nil {
			fmt.Printf("❌ wg-quick down failed for %s: %v\n", vpnName, err)
			os.Exit(1)
		}

		if err := os.Remove(tmpPath); err == nil {
			fmt.Printf("🧹 Removed temp file: %s\n", tmpPath)
		}

		fmt.Printf("✅ VPN %s brought down successfully.\n", vpnName)
	},
}
