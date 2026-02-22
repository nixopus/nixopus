package listcmd

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications in the family",
	Long:  `List all applications that belong to the current family.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Validate family_id exists
		if cfg.FamilyID == "" {
			return fmt.Errorf("family_id not found in config. Run 'nixopus live' to initialize first")
		}

		// List applications from config
		if len(cfg.Applications) == 0 {
			// Fallback to ProjectID if Applications map is empty
			if cfg.ProjectID != "" {
				fmt.Println("Applications in family:")
				fmt.Printf("  root (default) - application_id: %s\n", cfg.ProjectID)
				return nil
			}
			fmt.Println("No applications found in family.")
			return nil
		}

		fmt.Println("Applications in family:")
		for name, appID := range cfg.Applications {
			defaultMarker := ""
			if name == "default" {
				defaultMarker = " (default)"
			}
			fmt.Printf("  %s%s - application_id: %s\n", name, defaultMarker, appID)
		}

		return nil
	},
}

func init() {
	// No flags needed
}
