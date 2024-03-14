package cloudrun

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
)

const serviceInfoURLTmpl = "https://%s-run.googleapis.com/apis/serving.knative.dev/v1/namespaces/%s/services/%s"

type (
	serviceInfo struct {
		Status status `json:"status"`
	}

	status struct {
		URL string `json:"url"`
	}
)

func GetServiceURL(ctx context.Context) (string, error) {
	client, err := google.DefaultClient(ctx)
	if err != nil {
		return "", fmt.Errorf("create google client: %w", err)
	}

	url := fmt.Sprintf(
		serviceInfoURLTmpl,
		viper.GetString("project-region"),
		viper.GetString("project-id"),
		viper.GetString("k-service"),
	)
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("call cloud run API %s with error: %w", url, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read cloud run API call body: %w", err)
	}

	var si serviceInfo
	if err := json.Unmarshal(body, &si); err != nil {
		return "", fmt.Errorf("unmarshal cloud run API call body: %w", err)
	}

	return si.Status.URL, nil
}
