package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"importarr/internal/api"
	"importarr/internal/config"
	"importarr/internal/models"
	"importarr/internal/tui"
)

var (
	interactive bool
	fallback    bool
	instance    string
	all         bool
)

var rootCmd = &cobra.Command{
	Use:   "importarr",
	Short: "Force import stuck queue items in Sonarr/Radarr",
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for stuck queue items",
	RunE:  runScan,
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Force import stuck queue items",
	RunE:  runImport,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage instance configurations",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured instances",
	RunE:  runConfigList,
}

func init() {
	scanCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive TUI mode")
	scanCmd.Flags().StringVarP(&instance, "instance", "n", "", "Target specific instance")
	scanCmd.Flags().BoolVarP(&all, "all", "a", false, "Target all instances")

	importCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive TUI mode")
	importCmd.Flags().BoolVarP(&fallback, "fallback", "f", false, "Remove and search on import failure")
	importCmd.Flags().StringVarP(&instance, "instance", "n", "", "Target specific instance")
	importCmd.Flags().BoolVarP(&all, "all", "a", false, "Target all instances")

	configCmd.AddCommand(configListCmd)

	rootCmd.AddCommand(scanCmd, importCmd, configCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	instances, err := config.Load()
	if err != nil {
		return err
	}
	instances = config.FilterInstances(instances, instance, all)

	if interactive {
		tui.Run(instances, false)
		return nil
	}

	l := log.New(os.Stdout)

	for _, inst := range instances {
		client, err := api.NewClient(inst)
		if err != nil {
			l.Error("failed to create client", "instance", inst.Name, "error", err)
			continue
		}

		records, err := client.GetQueue()
		if err != nil {
			l.Error("failed to fetch queue", "instance", inst.Name, "error", err)
			continue
		}

		l.Info(fmt.Sprintf("[%s] Found %d stuck item(s)", inst.Name, len(records)))
		for _, r := range records {
			msg := extractMessage(r.StatusMessages)
			l.Info(fmt.Sprintf("  #%d %s [%s]", r.ID, r.Title, msg))
		}
	}

	return nil
}

func runImport(cmd *cobra.Command, args []string) error {
	instances, err := config.Load()
	if err != nil {
		return err
	}
	instances = config.FilterInstances(instances, instance, all)

	if interactive {
		tui.Run(instances, fallback)
		return nil
	}

	l := log.New(os.Stdout)

	for _, inst := range instances {
		client, err := api.NewClient(inst)
		if err != nil {
			l.Error("failed to create client", "instance", inst.Name, "error", err)
			continue
		}

		records, err := client.GetQueue()
		if err != nil {
			l.Error("failed to fetch queue", "instance", inst.Name, "error", err)
			continue
		}

		if len(records) == 0 {
			l.Info(fmt.Sprintf("[%s] No stuck items found", inst.Name))
			continue
		}

		l.Info(fmt.Sprintf("[%s] Processing %d stuck item(s)...", inst.Name, len(records)))

		totalOk, totalErr := 0, 0

		for _, record := range records {
			l.Info(fmt.Sprintf("  Processing: %s", record.Title))

			files, err := client.GetManualImport(record)
			if err != nil {
				l.Error("  manualimport GET failed", "error", err)
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						l.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						l.Error("  search trigger failed", "error", sErr)
					}
				}
				totalErr++
				continue
			}

			if len(files) == 0 {
				l.Warn("  No importable files found")
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						l.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						l.Error("  search trigger failed", "error", sErr)
					}
				}
				totalErr++
				continue
			}

			results, err := client.PostManualImport(files)
			if err != nil {
				l.Error("  import failed", "error", err)
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						l.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						l.Error("  search trigger failed", "error", sErr)
					}
				}
				totalErr++
				continue
			}

			for _, r := range results {
				switch r.Status {
				case "imported":
					l.Info(fmt.Sprintf("    [OK] %s", shortLogPath(r.Path)))
					totalOk++
				case "skipped":
					l.Warn(fmt.Sprintf("    [SKIP] %s - %s", shortLogPath(r.Path), r.Message))
				case "rejected":
					l.Error(fmt.Sprintf("    [REJECT] %s - %s", shortLogPath(r.Path), r.Message))
					totalErr++
				}
			}
		}

		l.Info(fmt.Sprintf("[%s] Done - Imported: %d, Failed: %d", inst.Name, totalOk, totalErr))
	}

	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	instances, err := config.Load()
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		fmt.Println("No instances configured.")
		fmt.Println("Set SONARR_URL/SONARR_API_KEY in .env or create config.yaml.")
		return nil
	}

	fmt.Println("Configured instances:")
	for _, inst := range instances {
		fmt.Printf("  %s (%s) - %s\n", inst.Name, inst.Type, inst.URL)
	}
	return nil
}

func extractMessage(messages []models.StatusMessage) string {
	for _, sm := range messages {
		if len(sm.Messages) > 0 {
			return sm.Messages[0]
		}
	}
	return ""
}

func shortLogPath(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) > 3 {
		return "..." + strings.Join(parts[len(parts)-3:], "/")
	}
	return p
}
