package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"importarr/internal/api"
	"importarr/internal/config"
	"importarr/internal/logger"
	"importarr/internal/models"
	"importarr/internal/tui"
)

var (
	interactive bool
	fallback    bool
	instance    string
	all         bool
	interval    time.Duration
)

var mauveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))

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

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously scan and import stuck items at interval",
	RunE:  runWatch,
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

	watchCmd.Flags().DurationVarP(&interval, "interval", "t", 10*time.Minute, "Scan interval (e.g., 5m, 1h)")
	watchCmd.Flags().BoolVarP(&fallback, "fallback", "f", false, "Remove and search on import failure")
	watchCmd.Flags().StringVarP(&instance, "instance", "n", "", "Target specific instance")
	watchCmd.Flags().BoolVarP(&all, "all", "a", false, "Target all instances")

	configCmd.AddCommand(configListCmd)

	rootCmd.AddCommand(scanCmd, importCmd, watchCmd, configCmd)
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

	for _, inst := range instances {
		il := logger.ForInstance(inst.Type)

		client, err := api.NewClient(inst)
		if err != nil {
			il.Error("failed to create client", "error", err)
			continue
		}

		records, err := client.GetQueue()
		if err != nil {
			il.Error("failed to fetch queue", "error", err)
			continue
		}

		il.Info(fmt.Sprintf("[%s] Found %d stuck item(s)", inst.Name, len(records)))
		for _, r := range records {
			msg := extractMessage(r.StatusMessages)
			il.Info(fmt.Sprintf("  #%d %s [%s]", r.ID, mauveStyle.Render(r.Title), msg))
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

	l := logger.New()

	totalOk, totalErr := 0, 0

	for _, inst := range instances {
		il := logger.ForInstance(inst.Type)

		client, err := api.NewClient(inst)
		if err != nil {
			il.Error("failed to create client", "error", err)
			continue
		}

		records, err := client.GetQueue()
		if err != nil {
			il.Error("failed to fetch queue", "error", err)
			continue
		}

		if len(records) == 0 {
			il.Info(fmt.Sprintf("[%s] No stuck items found", inst.Name))
			continue
		}

		il.Info(fmt.Sprintf("[%s] Processing %d stuck item(s)...", inst.Name, len(records)))

		for _, record := range records {
			il.Info(fmt.Sprintf("  Processing: %s", mauveStyle.Render(record.Title)))

			files, err := client.GetManualImport(record)
			if err != nil {
				il.Error("  manualimport GET failed", "error", err)
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						il.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						il.Error("  search trigger failed", "error", sErr)
					}
				}
				totalErr++
				continue
			}

			if len(files) == 0 {
				il.Warn("  No importable files found")
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						il.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						il.Error("  search trigger failed", "error", sErr)
					}
				}
				totalErr++
				continue
			}

			results, err := client.PostManualImport(files)
			if err != nil {
				il.Error("  import failed", "error", err)
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						il.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						il.Error("  search trigger failed", "error", sErr)
					}
				}
				totalErr++
				continue
			}

			for _, r := range results {
				switch r.Status {
				case "imported":
					il.Info(fmt.Sprintf("    [OK] %s", shortLogPath(r.Path)))
					totalOk++
				case "skipped":
					il.Warn(fmt.Sprintf("    [SKIP] %s - %s", shortLogPath(r.Path), r.Message))
				case "rejected":
					il.Error(fmt.Sprintf("    [REJECT] %s - %s", shortLogPath(r.Path), r.Message))
					totalErr++
				}
			}
		}

		il.Info(fmt.Sprintf("[%s] Done", inst.Name))
	}

	l.Info(fmt.Sprintf("Total - Imported: %d, Failed: %d", totalOk, totalErr))
	return nil
}

func runWatch(cmd *cobra.Command, args []string) error {
	instances, err := config.Load()
	if err != nil {
		return err
	}
	instances = config.FilterInstances(instances, instance, all)

	l := logger.New()
	l.Info(fmt.Sprintf("Watch mode started (interval: %s)", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		runImportLoop(instances, fallback, l)
		l.Info(fmt.Sprintf("Next scan in %s...", interval))
		<-ticker.C
	}
}

func runImportLoop(instances []models.Instance, fallback bool, l *log.Logger) {
	for _, inst := range instances {
		il := logger.ForInstance(inst.Type)

		client, err := api.NewClient(inst)
		if err != nil {
			il.Error("failed to create client", "error", err)
			continue
		}

		records, err := client.GetQueue()
		if err != nil {
			il.Error("failed to fetch queue", "error", err)
			continue
		}

		if len(records) == 0 {
			il.Info(fmt.Sprintf("[%s] No stuck items", inst.Name))
			continue
		}

		il.Info(fmt.Sprintf("[%s] Processing %d stuck item(s)...", inst.Name, len(records)))

		for _, record := range records {
			il.Info(fmt.Sprintf("  Processing: %s", mauveStyle.Render(record.Title)))

			files, err := client.GetManualImport(record)
			if err != nil {
				il.Error("  manualimport GET failed", "error", err)
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						il.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						il.Error("  search trigger failed", "error", sErr)
					}
				}
				continue
			}

			if len(files) == 0 {
				il.Warn("  No importable files found")
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						il.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						il.Error("  search trigger failed", "error", sErr)
					}
				}
				continue
			}

			results, err := client.PostManualImport(files)
			if err != nil {
				il.Error("  import failed", "error", err)
				if fallback {
					if rmErr := client.RemoveFromQueue(record.ID); rmErr != nil {
						il.Error("  remove failed", "error", rmErr)
					}
					if sErr := client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber); sErr != nil {
						il.Error("  search trigger failed", "error", sErr)
					}
				}
				continue
			}

			for _, r := range results {
				switch r.Status {
				case "imported":
					il.Info(fmt.Sprintf("    [OK] %s", shortLogPath(r.Path)))
				case "skipped":
					il.Warn(fmt.Sprintf("    [SKIP] %s - %s", shortLogPath(r.Path), r.Message))
				case "rejected":
					il.Error(fmt.Sprintf("    [REJECT] %s - %s", shortLogPath(r.Path), r.Message))
				}
			}
		}
	}
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
