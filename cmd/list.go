package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/csepant/whatchanged/output"
	"github.com/csepant/whatchanged/provider"
)

var listSnapshots bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List providers or snapshots",
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&listSnapshots, "snapshots", false, "List saved snapshots instead of providers")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	if noColor {
		output.SetNoColor(true)
	}

	if listSnapshots {
		return listSavedSnapshots()
	}

	names := provider.Names()
	if len(names) == 0 {
		fmt.Println("No providers registered.")
		return nil
	}

	t := &output.Table{
		Headers: []string{"PROVIDER", "TYPE"},
	}
	for _, name := range names {
		t.Rows = append(t.Rows, []string{name, name})
	}
	t.Print(os.Stdout)
	return nil
}

func listSavedSnapshots() error {
	store, err := newStorage()
	if err != nil {
		return err
	}
	defer store.Close()

	metas, err := store.List(getProfile(), getRegion(), 0)
	if err != nil {
		return err
	}

	if len(metas) == 0 {
		fmt.Println("No snapshots found. Run 'whatchanged snap' first.")
		return nil
	}

	t := &output.Table{
		Headers: []string{"ID", "DATE", "RESOURCES", "PROVIDERS"},
	}
	for _, m := range metas {
		t.Rows = append(t.Rows, []string{
			m.ID[:8],
			m.CreatedAt.Format("2006-01-02 15:04:05"),
			strconv.Itoa(m.ResourceCount),
			strings.Join(m.Providers, ", "),
		})
	}
	t.Print(os.Stdout)
	return nil
}
