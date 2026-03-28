package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ccontrerasi/whatchanged/provider"
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
	if listSnapshots {
		return listSavedSnapshots()
	}

	// List providers
	names := provider.Names()
	if len(names) == 0 {
		fmt.Println("No providers registered.")
		return nil
	}

	fmt.Printf("%-25s %s\n", "PROVIDER", "TYPE")
	fmt.Printf("%-25s %s\n", strings.Repeat("-", 25), strings.Repeat("-", 20))
	for _, name := range names {
		fmt.Printf("%-25s %s\n", name, name)
	}

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

	fmt.Printf("%-10s %-20s %-10s %s\n", "ID", "DATE", "RESOURCES", "PROVIDERS")
	fmt.Printf("%-10s %-20s %-10s %s\n", strings.Repeat("-", 10), strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 30))
	for _, m := range metas {
		fmt.Printf("%-10s %-20s %-10d %s\n",
			m.ID[:8],
			m.CreatedAt.Format("2006-01-02 15:04:05"),
			m.ResourceCount,
			strings.Join(m.Providers, ", "),
		)
	}

	return nil
}
