package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"

	"github.com/ccontrerasi/whatchanged/diff"
)

func init() {
	// Respect NO_COLOR env var
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}

// SetNoColor disables colored output.
func SetNoColor(disabled bool) {
	color.NoColor = disabled
}

// PrintDiff prints a DiffResult to the writer.
func PrintDiff(w io.Writer, result *diff.DiffResult) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	bold := color.New(color.Bold)

	// Header
	bold.Fprintf(w, "Comparing snapshots\n")
	fmt.Fprintf(w, "  Old: %s (%s)\n", result.OldSnapshot[:8], result.OldTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "  New: %s (%s)\n", result.NewSnapshot[:8], result.NewTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(w)

	// Summary
	bold.Fprintf(w, "Summary: ")
	parts := []string{}
	if len(result.Added) > 0 {
		parts = append(parts, green.Sprintf("%d added", len(result.Added)))
	}
	if len(result.Removed) > 0 {
		parts = append(parts, red.Sprintf("%d removed", len(result.Removed)))
	}
	if len(result.Modified) > 0 {
		parts = append(parts, yellow.Sprintf("%d modified", len(result.Modified)))
	}
	if result.Unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", result.Unchanged))
	}

	if !result.HasChanges() {
		green.Fprintln(w, "No changes detected")
		return
	}

	fmt.Fprintln(w, strings.Join(parts, ", "))
	fmt.Fprintln(w)

	// Group by resource type
	added := groupByType(result.Added)
	removed := groupByType(result.Removed)
	modified := groupByType(result.Modified)

	types := allTypes(added, removed, modified)
	sort.Strings(types)

	for _, resType := range types {
		bold.Fprintf(w, "── %s ──\n", resType)

		for _, rd := range added[resType] {
			name := resourceName(rd)
			green.Fprintf(w, "  + %s %s", rd.Resource.ID, name)
			printKeyProps(w, rd)
			fmt.Fprintln(w)
		}

		for _, rd := range removed[resType] {
			name := resourceName(rd)
			red.Fprintf(w, "  - %s %s\n", rd.Resource.ID, name)
		}

		for _, rd := range modified[resType] {
			name := resourceName(rd)
			yellow.Fprintf(w, "  ~ %s %s\n", rd.Resource.ID, name)
			for _, change := range rd.Changes {
				fmt.Fprintf(w, "      %s: %s → %s\n", change.Property, change.OldValue, change.NewValue)
			}
		}

		fmt.Fprintln(w)
	}
}

// PrintDiffJSON outputs the DiffResult as JSON.
func PrintDiffJSON(w io.Writer, result *diff.DiffResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func resourceName(rd diff.ResourceDiff) string {
	if name, ok := rd.Resource.Tags["Name"]; ok {
		return fmt.Sprintf("(%s)", name)
	}
	if name, ok := rd.Resource.Properties["group_name"]; ok {
		return fmt.Sprintf("(%s)", name)
	}
	return ""
}

func printKeyProps(w io.Writer, rd diff.ResourceDiff) {
	props := rd.Resource.Properties
	var parts []string
	for _, key := range []string{"instance_type", "state"} {
		if v, ok := props[key]; ok && v != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", key, v))
		}
	}
	if len(parts) > 0 {
		fmt.Fprintf(w, " [%s]", strings.Join(parts, ", "))
	}
}

func groupByType(diffs []diff.ResourceDiff) map[string][]diff.ResourceDiff {
	m := make(map[string][]diff.ResourceDiff)
	for _, d := range diffs {
		m[d.Resource.Type] = append(m[d.Resource.Type], d)
	}
	return m
}

func allTypes(maps ...map[string][]diff.ResourceDiff) []string {
	seen := make(map[string]bool)
	for _, m := range maps {
		for t := range m {
			seen[t] = true
		}
	}
	types := make([]string, 0, len(seen))
	for t := range seen {
		types = append(types, t)
	}
	return types
}
