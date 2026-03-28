package cmd

import (
	"fmt"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/csepant/whatchanged/output"
	"github.com/csepant/whatchanged/provider"
	"github.com/csepant/whatchanged/storage"
)

var snapProviders []string

var snapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Take a snapshot of current AWS resources",
	RunE:  runSnap,
}

func init() {
	snapCmd.Flags().StringSliceVar(&snapProviders, "provider", nil, "Providers to run (default: all)")
	rootCmd.AddCommand(snapCmd)
}

func runSnap(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if noColor {
		output.SetNoColor(true)
	}

	// Load AWS config
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(getRegion()),
	}
	if p := getProfile(); p != "default" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(p))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load AWS credentials. Run 'aws configure' or set AWS_PROFILE.\n  Error: %w", err)
	}

	// Determine which providers to run
	providers := getProviders()

	var allResources []provider.Resource
	var providerNames []string
	counts := make(map[string]int)
	var fetchErrors []string

	err = output.RunWithSpinner("Fetching AWS resources...", func(updateMsg func(string)) error {
		for _, name := range providers {
			p, err := provider.Get(name)
			if err != nil {
				fetchErrors = append(fetchErrors, fmt.Sprintf("Warning: %v", err))
				continue
			}

			updateMsg(fmt.Sprintf("Fetching %s...", name))
			resources, err := p.List(ctx, cfg)
			if err != nil {
				fetchErrors = append(fetchErrors, fmt.Sprintf("Error fetching %s: %v", name, err))
				continue
			}

			allResources = append(allResources, resources...)
			providerNames = append(providerNames, name)
			counts[name] = len(resources)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Print any warnings/errors that occurred
	for _, e := range fetchErrors {
		fmt.Println(e)
	}

	// Create snapshot
	snap := &storage.Snapshot{
		ID:        uuid.New().String(),
		Profile:   getProfile(),
		Region:    getRegion(),
		Providers: providerNames,
		Resources: allResources,
		CreatedAt: time.Now().UTC(),
	}

	// Save
	store, err := newStorage()
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Save(snap); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	// Print summary
	var parts []string
	for _, name := range providerNames {
		parts = append(parts, fmt.Sprintf("%d %s", counts[name], name))
	}
	fmt.Printf("Snapshot %s saved: %d resources (%s)\n", snap.ID[:8], len(allResources), strings.Join(parts, ", "))

	return nil
}

func getProviders() []string {
	if len(snapProviders) > 0 {
		return snapProviders
	}
	return provider.Names()
}

func newStorage() (storage.Storage, error) {
	dir := getDataDir()
	switch storageBackend {
	case "json":
		return storage.NewJSONStorage(dir)
	default:
		return storage.NewJSONStorage(dir)
	}
}
