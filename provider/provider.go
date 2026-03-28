package provider

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// Resource represents a single AWS resource in a normalized form.
type Resource struct {
	Type      string            `json:"type"`
	ID        string            `json:"id"`
	ARN       string            `json:"arn,omitempty"`
	Region    string            `json:"region"`
	AccountID string            `json:"account_id"`
	Tags      map[string]string `json:"tags,omitempty"`
	Properties map[string]string `json:"properties"`
	FetchedAt time.Time         `json:"fetched_at"`
}

// Provider is the interface that all resource providers must implement.
type Provider interface {
	ResourceType() string
	List(ctx context.Context, cfg aws.Config) ([]Resource, error)
}
