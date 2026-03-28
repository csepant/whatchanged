package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/ccontrerasi/whatchanged/diff"
	"github.com/ccontrerasi/whatchanged/output"
)

var (
	diffSince    string
	diffSnapshot string
	diffOutput   string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show changes between two snapshots",
	RunE:  runDiff,
}

func init() {
	diffCmd.Flags().StringVar(&diffSince, "since", "", "Compare against snapshot from this long ago (e.g. 2h, 1d, 7d)")
	diffCmd.Flags().StringVar(&diffSnapshot, "snapshot", "", "Compare against a specific snapshot ID")
	diffCmd.Flags().StringVar(&diffOutput, "output", "text", "Output format: text | json")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	if noColor {
		output.SetNoColor(true)
	}

	store, err := newStorage()
	if err != nil {
		return err
	}
	defer store.Close()

	profile := getProfile()
	region := getRegion()

	// Get the new (latest) snapshot
	newSnap, err := store.Latest(profile, region)
	if err != nil {
		return fmt.Errorf("failed to load latest snapshot: %w", err)
	}
	if newSnap == nil {
		return fmt.Errorf("no snapshots found. Run 'whatchanged snap' first")
	}

	// Get the old snapshot
	var oldSnap = newSnap // will be replaced
	switch {
	case diffSnapshot != "":
		oldSnap, err = store.Get(diffSnapshot)
		if err != nil {
			return fmt.Errorf("failed to load snapshot %s: %w", diffSnapshot, err)
		}

	case diffSince != "":
		dur, err := parseDuration(diffSince)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
		oldSnap, err = store.Before(profile, region, time.Now().Add(-dur))
		if err != nil {
			return fmt.Errorf("failed to find snapshot: %w", err)
		}
		if oldSnap == nil {
			return fmt.Errorf("no snapshot found before %s ago", diffSince)
		}

	default:
		// Compare latest vs second-latest
		metas, err := store.List(profile, region, 2)
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}
		if len(metas) < 2 {
			return fmt.Errorf("need at least 2 snapshots to diff. Run 'whatchanged snap' again")
		}
		oldSnap, err = store.Get(metas[1].ID)
		if err != nil {
			return fmt.Errorf("failed to load snapshot: %w", err)
		}
	}

	// Run diff
	result := diff.Compare(oldSnap, newSnap)

	// Output
	switch diffOutput {
	case "json":
		if err := output.PrintDiffJSON(os.Stdout, result); err != nil {
			return err
		}
	default:
		output.PrintDiff(os.Stdout, result)
	}

	// Exit code 1 if changes detected
	if result.HasChanges() {
		os.Exit(1)
	}
	return nil
}

// parseDuration parses duration strings like "2h", "1d", "7d".
func parseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)([hdwm])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		// Try standard Go duration
		return time.ParseDuration(s)
	}

	n, _ := strconv.Atoi(matches[1])
	switch matches[2] {
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "d":
		return time.Duration(n) * 24 * time.Hour, nil
	case "w":
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case "m":
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("unknown unit in %q", s)
}
