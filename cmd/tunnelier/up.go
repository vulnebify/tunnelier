package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulnebify/tunnelier/internal/mongo"
	"github.com/vulnebify/tunnelier/internal/vpn"
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
			return fmt.Errorf("mongo URL must be provided via --mongo-url or TUNNELIER_MONGO_URL")
		}

		store, err := mongo.NewStore(ctx, url, mongoDB, mongoCollection)
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

		for _, vpnConfig := range confs {
			if vpn.TryWGConnection(cmd.OutOrStdout(), cmd.ErrOrStderr(), vpnConfig.Name, vpnConfig.Config) {
				return nil
			}
		}
		return fmt.Errorf("all VPN configs failed or were not externally reachable")
	},
}

func init() {
	upCmd.Flags().IntVar(&retries, "retries", 3, "Number of random VPN configs to try")
	upCmd.Flags().StringVar(&mongoURL, "mongo-url", "", "MongoDB connection URL (overrides TUNNELIER_MONGO_URL)")
	upCmd.Flags().StringVar(&mongoDB, "mongo-db", "tunnelier", "MongoDB database name")
	upCmd.Flags().StringVar(&mongoCollection, "mongo-collection", "configs", "MongoDB collection name")
}
