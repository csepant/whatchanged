package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/ccontrerasi/whatchanged/provider"
	"github.com/ccontrerasi/whatchanged/storage"
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

	// Fetch resources from each provider
	var allResources []provider.Resource
	var providerNames []string
	counts := make(map[string]int)

	for _, name := range providers {
		p, err := provider.Get(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}

		fmt.Fprintf(os.Stderr, "Fetching %s...\n", name)
		resources, err := p.List(ctx, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", name, err)
			continue
		}

		allResources = append(allResources, resources...)
		providerNames = append(providerNames, name)
		counts[name] = len(resources)
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
