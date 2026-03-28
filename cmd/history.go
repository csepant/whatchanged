package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ccontrerasi/whatchanged/output"
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
	if noColor {
		output.SetNoColor(true)
	}

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

	t := &output.Table{
		Headers: []string{"ID", "DATE", "PROVIDERS", "RESOURCES", "PROFILE/REGION"},
	}

	for _, m := range metas {
		t.Rows = append(t.Rows, []string{
			m.ID[:8],
			m.CreatedAt.Format("2006-01-02 15:04:05"),
			strconv.Itoa(len(m.Providers)),
			strconv.Itoa(m.ResourceCount),
			m.Profile + "/" + m.Region,
		})
	}

	t.Print(os.Stdout)
	return nil
}
