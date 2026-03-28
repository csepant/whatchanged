package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/csepant/whatchanged/diff"
)

var noColor bool

func init() {
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
}

// SetNoColor disables colored output.
func SetNoColor(disabled bool) {
	noColor = disabled
}

// Styles
var (
	addedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	removedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	modifiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // cyan
	faintStyle = lipgloss.NewStyle().Faint(true)
	boldStyle = lipgloss.NewStyle().Bold(true)

	hunkStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true)

	sectionStyle = lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("8")).
		MarginTop(1).
		PaddingBottom(0)

	summaryBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1)

	noChangesStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Bold(true)
)

func render(s lipgloss.Style, text string) string {
	if noColor {
		return text
	}
	return s.Render(text)
}

// PrintDiff prints a DiffResult in git-diff style with lipgloss styling.
func PrintDiff(w io.Writer, result *diff.DiffResult) {
	// Header
	fmt.Fprintf(w, "%s\n", render(boldStyle, fmt.Sprintf("diff snapshots %s..%s", result.OldSnapshot[:8], result.NewSnapshot[:8])))
	fmt.Fprintf(w, "%s\n", render(faintStyle, fmt.Sprintf("--- a/ %s (%s)", result.OldSnapshot[:8], result.OldTimestamp.Format("2006-01-02 15:04:05"))))
	fmt.Fprintf(w, "%s\n", render(faintStyle, fmt.Sprintf("+++ b/ %s (%s)", result.NewSnapshot[:8], result.NewTimestamp.Format("2006-01-02 15:04:05"))))
	fmt.Fprintln(w)

	// Summary
	if !result.HasChanges() {
		fmt.Fprintln(w, render(noChangesStyle, "  No changes detected"))
		return
	}

	var parts []string
	if len(result.Added) > 0 {
		parts = append(parts, render(addedStyle, fmt.Sprintf("%d added", len(result.Added))))
	}
	if len(result.Removed) > 0 {
		parts = append(parts, render(removedStyle, fmt.Sprintf("%d removed", len(result.Removed))))
	}
	if len(result.Modified) > 0 {
		parts = append(parts, render(modifiedStyle, fmt.Sprintf("%d modified", len(result.Modified))))
	}
	if result.Unchanged > 0 {
		parts = append(parts, render(faintStyle, fmt.Sprintf("%d unchanged", result.Unchanged)))
	}

	summary := fmt.Sprintf("  %s", strings.Join(parts, "  ·  "))
	if noColor {
		fmt.Fprintln(w, summary)
	} else {
		fmt.Fprintln(w, summaryBoxStyle.Render(strings.Join(parts, "  ·  ")))
	}
	fmt.Fprintln(w)

	// Group by resource type
	added := groupByType(result.Added)
	removed := groupByType(result.Removed)
	modified := groupByType(result.Modified)

	types := allTypes(added, removed, modified)
	sort.Strings(types)

	for _, resType := range types {
		// Section header
		fmt.Fprintln(w, render(sectionStyle, fmt.Sprintf(" %s ", resType)))

		for _, rd := range added[resType] {
			printHunkHeader(w, "added", rd)
			for _, key := range sortedKeys(rd.Resource.Properties) {
				fmt.Fprintf(w, "%s\n", render(addedStyle, fmt.Sprintf("+ %s: %s", key, rd.Resource.Properties[key])))
			}
			for _, key := range sortedKeys(rd.Resource.Tags) {
				fmt.Fprintf(w, "%s\n", render(addedStyle, fmt.Sprintf("+ tag:%s: %s", key, rd.Resource.Tags[key])))
			}
			fmt.Fprintln(w)
		}

		for _, rd := range removed[resType] {
			printHunkHeader(w, "removed", rd)
			for _, key := range sortedKeys(rd.Resource.Properties) {
				fmt.Fprintf(w, "%s\n", render(removedStyle, fmt.Sprintf("- %s: %s", key, rd.Resource.Properties[key])))
			}
			for _, key := range sortedKeys(rd.Resource.Tags) {
				fmt.Fprintf(w, "%s\n", render(removedStyle, fmt.Sprintf("- tag:%s: %s", key, rd.Resource.Tags[key])))
			}
			fmt.Fprintln(w)
		}

		for _, rd := range modified[resType] {
			printHunkHeader(w, "modified", rd)
			sortedChanges := make([]diff.PropertyChange, len(rd.Changes))
			copy(sortedChanges, rd.Changes)
			sort.Slice(sortedChanges, func(i, j int) bool {
				return sortedChanges[i].Property < sortedChanges[j].Property
			})
			for _, change := range sortedChanges {
				fmt.Fprintf(w, "%s\n", render(removedStyle, fmt.Sprintf("- %s: %s", change.Property, change.OldValue)))
				fmt.Fprintf(w, "%s\n", render(addedStyle, fmt.Sprintf("+ %s: %s", change.Property, change.NewValue)))
			}
			fmt.Fprintln(w)
		}
	}
}

// PrintDiffJSON outputs the DiffResult as JSON.
func PrintDiffJSON(w io.Writer, result *diff.DiffResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func printHunkHeader(w io.Writer, changeType string, rd diff.ResourceDiff) {
	name := resourceName(rd)
	var header string
	if name != "" {
		header = fmt.Sprintf("@@ %s %s %s @@", rd.Resource.ID, name, changeType)
	} else {
		header = fmt.Sprintf("@@ %s %s @@", rd.Resource.ID, changeType)
	}
	fmt.Fprintln(w, render(hunkStyle, header))
}

func resourceName(rd diff.ResourceDiff) string {
	if name, ok := rd.Resource.Tags["Name"]; ok {
		return name
	}
	if name, ok := rd.Resource.Properties["group_name"]; ok {
		return name
	}
	return ""
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
