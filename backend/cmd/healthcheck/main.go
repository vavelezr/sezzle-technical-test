// Command healthcheck probes the service's health endpoint and exits with a
// non-zero status if it is unhealthy. It exists because the distroless
// runtime image has no shell, curl or wget for a container HEALTHCHECK.
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"sezzle-calculator/backend/internal/config"
	httptransport "sezzle-calculator/backend/internal/transport/http"
)

const timeout = 2 * time.Second

func main() {
	if err := check(); err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck failed:", err)
		os.Exit(1)
	}
}

func check() error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://127.0.0.1:%d%s/health", cfg.Port, httptransport.APIPrefix)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
