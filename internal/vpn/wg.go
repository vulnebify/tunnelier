package vpn

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var jsonIPServices = []string{
	"https://api.ipify.org/?format=json",
	"https://api.myip.com/",
	"https://get.geojs.io/v1/ip.json",
	"https://api.ip.sb/jsonip",
	"https://l2.io/ip.json",
}

func fetchPublicIP() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	for _, service := range jsonIPServices {
		resp, err := client.Get(service)

		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			continue
		}

		if ip, ok := parsed["ip"].(string); ok && ip != "" {
			return ip, nil
		}
	}

	return "", errors.New("failed to fetch public IP from all services")
}

func TryWGConnection(cmdOut, cmdErr io.Writer, vpnName, config string) bool {
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.conf", vpnName))
	if err := os.WriteFile(tmpPath, []byte(config), 0600); err != nil {
		fmt.Fprintf(cmdErr, "❌ Failed to write config: %s — %v\n", vpnName, err)
		return false
	}

	_ = exec.Command("wg-quick", "down", tmpPath).Run()

	fmt.Fprintf(cmdOut, "🚀 Trying VPN: %s\n", vpnName)

	wgCmd := exec.Command("wg-quick", "up", tmpPath)
	wgCmd.Stdout = cmdOut
	wgCmd.Stderr = cmdErr

	if err := wgCmd.Run(); err != nil {
		fmt.Fprintf(cmdErr, "❌ Failed: %s — %v\n", vpnName, err)
		return false
	}

	for attempt := 1; attempt <= 3; attempt++ {
		ip, e := fetchPublicIP()
		if e == nil {
			fmt.Fprintf(cmdOut, "✅ VPN connection validated with external IP check: %s\n", ip)
			return true
		}
		fmt.Fprintf(cmdErr, "❌ IP check attempt %d failed, retrying...\n", attempt)
		time.Sleep(time.Duration(attempt*attempt) * time.Second)
	}

	// Disconnect after failed validation
	_ = exec.Command("wg-quick", "down", tmpPath).Run()
	fmt.Fprintf(cmdErr, "❌ VPN connection not validated after retries\n")
	return false
}
