// Package firestore is the Cloud Firestore implementation of the
// service/report.Repository interface. Documents live in the report_votes
// collection, one per /report invocation, keyed by
// {chatID}_{targetUserID}_{unixNano} for human-readable history.
package firestore

import (
	"context"
	"fmt"

	fsdk "cloud.google.com/go/firestore"
)

// Client aliases the SDK type so package consumers don't need to import the
// SDK directly when they only need lifecycle (NewClient / Close).
type Client = fsdk.Client

// NewClient builds a Firestore client for the given GCP project.
// In production the project comes from PROJECT_ID (set by Cloud Run).
// For local dev and tests set FIRESTORE_EMULATOR_HOST to point at the emulator.
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	c, err := fsdk.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore.NewClient: %w", err)
	}

	return c, nil
}
