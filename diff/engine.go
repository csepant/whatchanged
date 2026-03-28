package diff

import (
	"time"

	"github.com/ccontrerasi/whatchanged/provider"
	"github.com/ccontrerasi/whatchanged/storage"
)

// DiffResult contains the full diff between two snapshots.
type DiffResult struct {
	OldSnapshot  string         `json:"old_snapshot"`
	NewSnapshot  string         `json:"new_snapshot"`
	OldTimestamp time.Time      `json:"old_timestamp"`
	NewTimestamp time.Time      `json:"new_timestamp"`
	Profile      string         `json:"profile"`
	Region       string         `json:"region"`
	Added        []ResourceDiff `json:"added,omitempty"`
	Removed      []ResourceDiff `json:"removed,omitempty"`
	Modified     []ResourceDiff `json:"modified,omitempty"`
	Unchanged    int            `json:"unchanged"`
}

// ResourceDiff represents a single resource change.
type ResourceDiff struct {
	Resource   provider.Resource `json:"resource"`
	ChangeType string            `json:"change_type"`
	Changes    []PropertyChange  `json:"changes,omitempty"`
}

// PropertyChange is a single property that changed on a resource.
type PropertyChange struct {
	Property string `json:"property"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// HasChanges returns true if there are any differences.
func (d *DiffResult) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Modified) > 0
}

// Compare takes two snapshots and produces a DiffResult.
func Compare(old, new *storage.Snapshot) *DiffResult {
	result := &DiffResult{
		OldSnapshot:  old.ID,
		NewSnapshot:  new.ID,
		OldTimestamp: old.CreatedAt,
		NewTimestamp: new.CreatedAt,
		Profile:      new.Profile,
		Region:       new.Region,
	}

	oldMap := buildResourceMap(old.Resources)
	newMap := buildResourceMap(new.Resources)

	// Find added and modified resources
	for key, newRes := range newMap {
		oldRes, exists := oldMap[key]
		if !exists {
			result.Added = append(result.Added, ResourceDiff{
				Resource:   newRes,
				ChangeType: "added",
			})
			continue
		}

		changes := compareResources(oldRes, newRes)
		if len(changes) > 0 {
			result.Modified = append(result.Modified, ResourceDiff{
				Resource:   newRes,
				ChangeType: "modified",
				Changes:    changes,
			})
		} else {
			result.Unchanged++
		}
	}

	// Find removed resources
	for key, oldRes := range oldMap {
		if _, exists := newMap[key]; !exists {
			result.Removed = append(result.Removed, ResourceDiff{
				Resource:   oldRes,
				ChangeType: "removed",
			})
		}
	}

	return result
}

func resourceKey(r provider.Resource) string {
	return r.Type + ":" + r.ID
}

func buildResourceMap(resources []provider.Resource) map[string]provider.Resource {
	m := make(map[string]provider.Resource, len(resources))
	for _, r := range resources {
		m[resourceKey(r)] = r
	}
	return m
}

func compareResources(old, new provider.Resource) []PropertyChange {
	var changes []PropertyChange

	// Compare properties
	allProps := make(map[string]bool)
	for k := range old.Properties {
		allProps[k] = true
	}
	for k := range new.Properties {
		allProps[k] = true
	}

	for prop := range allProps {
		oldVal := old.Properties[prop]
		newVal := new.Properties[prop]
		if oldVal != newVal {
			changes = append(changes, PropertyChange{
				Property: prop,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}

	// Compare tags
	allTags := make(map[string]bool)
	for k := range old.Tags {
		allTags[k] = true
	}
	for k := range new.Tags {
		allTags[k] = true
	}

	for tag := range allTags {
		oldVal := old.Tags[tag]
		newVal := new.Tags[tag]
		if oldVal != newVal {
			changes = append(changes, PropertyChange{
				Property: "tag:" + tag,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}

	return changes
}
