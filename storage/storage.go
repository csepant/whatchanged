package storage

import (
	"time"

	"github.com/ccontrerasi/whatchanged/provider"
)

// Snapshot represents a point-in-time capture of AWS resources.
type Snapshot struct {
	ID        string              `json:"id"`
	Profile   string              `json:"profile"`
	Region    string              `json:"region"`
	Providers []string            `json:"providers"`
	Resources []provider.Resource `json:"resources"`
	CreatedAt time.Time           `json:"created_at"`
}

// SnapshotMeta is a lightweight snapshot reference for listing.
type SnapshotMeta struct {
	ID            string    `json:"id"`
	Profile       string    `json:"profile"`
	Region        string    `json:"region"`
	Providers     []string  `json:"providers"`
	ResourceCount int       `json:"resource_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// Storage is the interface for snapshot persistence.
type Storage interface {
	Save(snap *Snapshot) error
	Latest(profile, region string) (*Snapshot, error)
	Get(id string) (*Snapshot, error)
	List(profile, region string, limit int) ([]SnapshotMeta, error)
	Before(profile, region string, t time.Time) (*Snapshot, error)
	Close() error
}
