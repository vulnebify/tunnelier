package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulnebify/tunnelier/internal/mongo"
)

var (
	retries int
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Connect to a working WireGuard VPN from MongoDB",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return ensureWireGuardInstalled()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		url := getMongoURL()
		if url == "" {
			return fmt.Errorf("mongo URL must be provided via --mongo-url or TUNNELIER_MONGO_URI")
		}

		store, err := mongo.NewClient(ctx, url, mongoDB, mongoCollection)
		if err != nil {
			return fmt.Errorf("mongo connection failed: %w", err)
		}

		confs, err := store.FetchWireguardSample(ctx, retries)
		if err != nil {
			return fmt.Errorf("failed to fetch VPN configs: %w", err)
		}

		if len(confs) == 0 {
			return fmt.Errorf("no VPN configs found")
		}

		for _, vpn := range confs {
			tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.conf", vpn.Name))
			if err := os.WriteFile(tmpPath, []byte(vpn.Config), 0600); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "❌ Failed to write config: %s — %v\n", vpn.Name, err)
				continue
			}

			_ = exec.Command("wg-quick", "down", tmpPath).Run()

			fmt.Fprintf(cmd.OutOrStdout(), "🚀 Trying VPN: %s\n", vpn.Name)

			wgCmd := exec.Command("wg-quick", "up", tmpPath)
			wgCmd.Stdout = cmd.OutOrStdout()
			wgCmd.Stderr = cmd.ErrOrStderr()

			if err := wgCmd.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "❌ Failed: %s — %v\n", vpn.Name, err)
				continue
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✅ Connected to VPN: %s\n", vpn.Name)
			return nil
		}

		return fmt.Errorf("all VPN configs failed")
	},
}

func init() {
	upCmd.Flags().IntVar(&retries, "retries", 3, "Number of random VPN configs to try")
	upCmd.Flags().StringVar(&mongoURL, "mongo-url", "", "MongoDB connection URL (overrides TUNNELIER_MONGO_URI)")
	upCmd.Flags().StringVar(&mongoDB, "mongo-db", "tunnelier", "MongoDB database name")
	upCmd.Flags().StringVar(&mongoCollection, "mongo-collection", "configs", "MongoDB collection name")
}
