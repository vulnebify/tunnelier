package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vulnebify/tunnelier/internal/mongo"
)

func setupTestEnv(t *testing.T) (string, error) {
	tmpDir := t.TempDir()

	confPath := filepath.Join(tmpDir, "test.conf")
	data := `[Interface]
PrivateKey = testkey
Address = 10.0.0.1/32

[Peer]
PublicKey = peerkey
Endpoint = 1.2.3.4:51820`

	if err := os.WriteFile(confPath, []byte(data), 0600); err != nil {
		return "", err
	}

	return tmpDir, nil
}

func TestImportCommand_Integration(t *testing.T) {
	tmpDir, err := setupTestEnv(t)
	if err != nil {
		t.Fatalf("failed to setup temp conf dir: %v", err)
	}

	mongoURL := getenv("MONGO_URL", "mongodb://admin:adminpassword@localhost:27017")
	mongoDB := getenv("MONGO_DB", "tunnelier")
	mongoCollection := getenv("MONGO_COLLECTION", "configs")

	// Clear test document before test
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := mongo.NewStore(ctx, mongoURL, mongoDB, mongoCollection)
	if err != nil {
		t.Fatalf("Mongo connection failed: %v", err)
	}

	_ = store.Client.Database(mongoDB).Collection(mongoCollection).Drop(ctx)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"import",
		"--folder", tmpDir,
		"--mongo-url", mongoURL,
		"--mongo-db", mongoDB,
		"--mongo-collection", mongoCollection,
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("import command failed: %v", err)
	}

	// Verify insert
	results, err := store.FetchWireguardSample(ctx, 1)
	if err != nil {
		t.Fatalf("failed to fetch configs from db: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 config in DB, got %d", len(results))
	}

	if results[0].Name != "test" {
		t.Errorf("expected config name 'test', got %s", results[0].Name)
	}

	t.Log("✅ Import integration test passed.")
}

func TestImportCommand_FolderDoesNotExist(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{
		"import",
		"--folder", "./nonexistent-folder",
		"--mongo-url", getenv("MONGO_URL", "mongodb://admin:adminpassword@localhost:27017"),
		"--mongo-db", "tunnelier",
		"--mongo-collection", "configs",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := "no .conf files found"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error to contain %q, got %v", expected, err)
	}
}

func TestImportCommand_EmptyFolder(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd.SetOut(new(bytes.Buffer))
	rootCmd.SetErr(new(bytes.Buffer))

	rootCmd.SetArgs([]string{
		"import",
		"--folder", tmpDir,
		"--mongo-url", getenv("MONGO_URL", "mongodb://admin:adminpassword@localhost:27017"),
		"--mongo-db", "tunnelier",
		"--mongo-collection", "configs",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty folder, got nil")
	}

	if !strings.Contains(err.Error(), "no .conf files found") {
		t.Errorf("expected 'no .conf files found' error, got: %v", err)
	}
}
func TestImportCommand_InvalidConfFile(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "invalid.conf")

	if err := os.WriteFile(badFile, []byte("   \n\n"), 0600); err != nil {
		t.Fatalf("failed to write empty conf file: %v", err)
	}

	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetErr(errOut)

	rootCmd.SetArgs([]string{
		"import",
		"--folder", tmpDir,
		"--mongo-url", getenv("MONGO_URL", "mongodb://admin:adminpassword@localhost:27017"),
		"--mongo-db", "tunnelier",
		"--mongo-collection", "configs",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String() + errOut.String()

	if !strings.Contains(output, "Skipping empty config") {
		t.Errorf("expected warning for empty config, got: %q", output)
	}
}
