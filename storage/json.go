package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// JSONStorage stores snapshots as individual JSON files.
type JSONStorage struct {
	dir string
}

// NewJSONStorage creates a JSON storage backend in the given directory.
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	dir := filepath.Join(dataDir, "snapshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}
	return &JSONStorage{dir: dir}, nil
}

func (s *JSONStorage) Save(snap *Snapshot) error {
	filename := fmt.Sprintf("%s_%s_%s_%s.json",
		snap.Profile,
		snap.Region,
		snap.CreatedAt.Format("20060102T150405Z"),
		snap.ID[:8],
	)
	path := filepath.Join(s.dir, filename)

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

func (s *JSONStorage) Latest(profile, region string) (*Snapshot, error) {
	metas, err := s.List(profile, region, 1)
	if err != nil {
		return nil, err
	}
	if len(metas) == 0 {
		return nil, nil
	}
	return s.Get(metas[0].ID)
}

func (s *JSONStorage) Get(id string) (*Snapshot, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if strings.Contains(entry.Name(), id[:8]) {
			return s.loadFile(filepath.Join(s.dir, entry.Name()))
		}
	}

	return nil, fmt.Errorf("snapshot not found: %s", id)
}

func (s *JSONStorage) List(profile, region string, limit int) ([]SnapshotMeta, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var metas []SnapshotMeta
	prefix := fmt.Sprintf("%s_%s_", profile, region)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}

		snap, err := s.loadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}

		metas = append(metas, SnapshotMeta{
			ID:            snap.ID,
			Profile:       snap.Profile,
			Region:        snap.Region,
			Providers:     snap.Providers,
			ResourceCount: len(snap.Resources),
			CreatedAt:     snap.CreatedAt,
		})
	}

	sort.Slice(metas, func(i, j int) bool {
		return metas[i].CreatedAt.After(metas[j].CreatedAt)
	})

	if limit > 0 && len(metas) > limit {
		metas = metas[:limit]
	}

	return metas, nil
}

func (s *JSONStorage) Before(profile, region string, t time.Time) (*Snapshot, error) {
	metas, err := s.List(profile, region, 0)
	if err != nil {
		return nil, err
	}

	for _, meta := range metas {
		if meta.CreatedAt.Before(t) {
			return s.Get(meta.ID)
		}
	}

	return nil, nil
}

func (s *JSONStorage) Close() error {
	return nil
}

func (s *JSONStorage) loadFile(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}

	return &snap, nil
}

// DeleteOldSnapshots removes snapshots beyond the given limit for a profile/region.
func (s *JSONStorage) DeleteOldSnapshots(profile, region string, maxSnapshots int) (int, error) {
	if maxSnapshots <= 0 {
		return 0, nil
	}

	metas, err := s.List(profile, region, 0)
	if err != nil {
		return 0, err
	}

	if len(metas) <= maxSnapshots {
		return 0, nil
	}

	toDelete := metas[maxSnapshots:]
	deleted := 0

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, err
	}

	for _, meta := range toDelete {
		for _, entry := range entries {
			if strings.Contains(entry.Name(), meta.ID[:8]) {
				if err := os.Remove(filepath.Join(s.dir, entry.Name())); err == nil {
					deleted++
				}
				break
			}
		}
	}

	return deleted, nil
}
