package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/vulnebify/tunnelier/internal/mongo"
)

func TestUpCommand_MissingMongoURL(t *testing.T) {
	rootCmd.SetOut(new(bytes.Buffer))
	rootCmd.SetErr(new(bytes.Buffer))

	rootCmd.SetArgs([]string{
		"up",
		"--mongo-url", "",
		"--mongo-db", "tunnelier",
		"--mongo-collection", "configs",
	})

	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "mongo URL must be provided") {
		t.Fatalf("expected mongo URL error, got: %v", err)
	}
}

func TestUpCommand_NoConfigsInDB(t *testing.T) {
	ctx := context.Background()
	mongoURL := getenv("MONGO_URL", "mongodb://admin:adminpassword@localhost:27017")
	db := "tunnelier"
	col := "configs"

	store, err := mongo.NewClient(ctx, mongoURL, db, col)
	if err != nil {
		t.Fatalf("Mongo connection failed: %v", err)
	}

	_ = store.Client.Database(db).Collection(col).Drop(ctx)

	rootCmd.SetOut(new(bytes.Buffer))
	rootCmd.SetErr(new(bytes.Buffer))

	rootCmd.SetArgs([]string{
		"up",
		"--mongo-url", mongoURL,
		"--mongo-db", db,
		"--mongo-collection", col,
		"--retries", "2",
	})

	err = rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "no VPN configs found") {
		t.Fatalf("expected 'no VPN configs found' error, got: %v", err)
	}
}

func TestUpCommand_ConfigFailsToConnect(t *testing.T) {
	ctx := context.Background()
	mongoURL := getenv("MONGO_URL", "mongodb://admin:adminpassword@localhost:27017")
	db := "tunnelier"
	col := "configs"

	store, err := mongo.NewClient(ctx, mongoURL, db, col)
	if err != nil {
		t.Fatalf("Mongo connection failed: %v", err)
	}

	_ = store.Client.Database(db).Collection(col).Drop(ctx)

	err = store.StoreWireguardConfig(ctx, mongo.VPNConfig{
		Name: "failme",
		Type: "wireguard",
		Config: `[Interface]
PrivateKey = somekey
Address = 10.0.0.2/32
[Peer]
PublicKey = pubkey
Endpoint = 1.2.3.4:51820`,
	})
	if err != nil {
		t.Fatalf("failed to insert config: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{
		"up",
		"--mongo-url", mongoURL,
		"--mongo-db", db,
		"--mongo-collection", col,
		"--retries", "1",
	})

	err = rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "all VPN configs failed") {
		t.Fatalf("expected 'all VPN configs failed', got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Trying VPN: failme") {
		t.Errorf("expected output to contain 'Trying VPN: failme', got: %s", output)
	}
}
