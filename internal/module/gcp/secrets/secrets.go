package secrets

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/api/secretmanager/v1"
)

const BotTokenSecretID = "BOT_TOKEN"

type Client struct {
	*secretmanager.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	client, err := secretmanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret client: %w", err)
	}

	return &Client{client}, nil
}

// GetSecretValue returns the value of the secret with the given secret version name.
// Note: please, use the BuildSecretName function to build a correct secret version name.
func (c Client) GetSecretValue(ctx context.Context, name string) (string, error) {
	resp, err := c.Projects.Secrets.Versions.Access(name).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("secret access request: %w", err)
	}
	if resp.HTTPStatusCode != 200 { //nolint:gomnd,mnd
		return "", fmt.Errorf("secret access request: code=%d, data=%v", resp.HTTPStatusCode, resp.Payload.Data)
	}

	decoded, err := base64.StdEncoding.DecodeString(resp.Payload.Data)
	if err != nil {
		return "", fmt.Errorf("decode base64 string: %w", err)
	}

	return string(decoded), nil
}

func BuildSecretName(projectID, secretID, version string) string {
	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, version)
}
