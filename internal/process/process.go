package process

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Process(client *http.Client, url string) (string, error) {
	startTime := time.Now()
	resp, err := client.Get(url)

	if err != nil {
		return "", fmt.Errorf("client: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("client: url %s unavilble: status code: %d", url, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("client: cannot read body: %w", err)
	}

	return fmt.Sprintf("url: %s size: %d bytes time: %v", url, len(b), time.Since(startTime)), nil
}
