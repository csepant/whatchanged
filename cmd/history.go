package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show snapshot history",
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntVar(&historyLimit, "limit", 10, "Maximum number of snapshots to show")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	store, err := newStorage()
	if err != nil {
		return err
	}
	defer store.Close()

	metas, err := store.List(getProfile(), getRegion(), historyLimit)
	if err != nil {
		return err
	}

	if len(metas) == 0 {
		fmt.Println("No snapshots found. Run 'whatchanged snap' first.")
		return nil
	}

	fmt.Printf("%-10s %-20s %-12s %-10s %s\n", "ID", "DATE", "PROVIDERS", "RESOURCES", "PROFILE/REGION")
	fmt.Printf("%-10s %-20s %-12s %-10s %s\n",
		strings.Repeat("-", 10), strings.Repeat("-", 20), strings.Repeat("-", 12),
		strings.Repeat("-", 10), strings.Repeat("-", 20))

	for _, m := range metas {
		fmt.Printf("%-10s %-20s %-12d %-10d %s/%s\n",
			m.ID[:8],
			m.CreatedAt.Format("2006-01-02 15:04:05"),
			len(m.Providers),
			m.ResourceCount,
			m.Profile,
			m.Region,
		)
	}

	return nil
}
