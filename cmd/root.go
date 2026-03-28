package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	profile string
	region  string
	storageBackend string
	dataDir string
	noColor bool
)

var rootCmd = &cobra.Command{
	Use:   "whatchanged",
	Short: "Snapshot AWS account state and diff it over time",
	Long:  "A CLI tool that snapshots AWS account state and diffs it over time. Run 'whatchanged snap' to capture current resources, run 'whatchanged diff' to see what changed since last snapshot.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile (default: AWS_PROFILE env or 'default')")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region (default: AWS_DEFAULT_REGION env or 'us-east-1')")
	rootCmd.PersistentFlags().StringVar(&storageBackend, "storage", "json", "Storage backend: json | sqlite")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", defaultDataDir(), "Data directory")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".whatchanged"
	}
	return filepath.Join(home, ".whatchanged")
}

func getProfile() string {
	if profile != "" {
		return profile
	}
	if p := os.Getenv("AWS_PROFILE"); p != "" {
		return p
	}
	return "default"
}

func getRegion() string {
	if region != "" {
		return region
	}
	if r := os.Getenv("AWS_DEFAULT_REGION"); r != "" {
		return r
	}
	return "us-east-1"
}

func getDataDir() string {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data directory: %v\n", err)
		os.Exit(2)
	}
	return dataDir
}
