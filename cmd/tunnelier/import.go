package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulnebify/tunnelier/internal/mongo"
)

var (
	folder string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import WireGuard configs into MongoDB",
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := filepath.Glob(filepath.Join(folder, "*.conf"))
		if err != nil || len(files) == 0 {
			return fmt.Errorf("no .conf files found in: %s", folder)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mongoUrl := getMongoURL()
		store, err := mongo.NewClient(ctx, mongoUrl, mongoDB, mongoCollection)
		if err != nil {
			return fmt.Errorf("mongo connection error: %w", err)
		}

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "⚠️ Failed to read: %s: %v\n", file, err)
				continue
			}

			if len(bytes.TrimSpace(data)) == 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "⚠️ Skipping empty config: %s\n", file)
				continue
			}

			name := strings.TrimSuffix(filepath.Base(file), ".conf")

			cfg := mongo.VPNConfig{
				Name:   name,
				Type:   "wireguard",
				Config: string(data),
			}

			if err := store.StoreWireguardConfig(ctx, cfg); err != nil {
				fmt.Println("⚠️ Failed to insert:", name, "-", err)
				continue
			}

			fmt.Println("✅ Imported:", name)

		}

		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&folder, "folder", "f", ".", "Folder containing .conf files")
	importCmd.Flags().StringVar(&mongoURL, "mongo-url", "", "MongoDB connection URL (overrides TUNNELIER_MONGO_URI)")
	importCmd.Flags().StringVar(&mongoDB, "mongo-db", "tunnelier", "MongoDB database name")
	importCmd.Flags().StringVar(&mongoCollection, "mongo-collection", "configs", "MongoDB collection name")
}
