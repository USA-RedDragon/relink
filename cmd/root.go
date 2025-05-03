package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/USA-RedDragon/configulator"
	"github.com/USA-RedDragon/relink/internal/config"
	"github.com/USA-RedDragon/relink/internal/relink"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

func NewCommand(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relink",
		Version: fmt.Sprintf("%s - %s", version, commit),
		Annotations: map[string]string{
			"version": version,
			"commit":  commit,
		},
		RunE:              runRoot,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(NewBufferBenchCommand(version, commit))
	return cmd
}

func runRoot(cmd *cobra.Command, _ []string) error {
	fmt.Printf("relink - %s (%s)\n", cmd.Annotations["version"], cmd.Annotations["commit"])

	c, err := configulator.FromContext[config.Config](cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get config from context")
	}

	cfg, err := c.Load()
	if err != nil {
		return err
	}

	var logger *slog.Logger
	switch cfg.LogLevel {
	case config.LogLevelDebug:
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))
	case config.LogLevelInfo:
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelInfo}))
	case config.LogLevelWarn:
		logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelWarn}))
	case config.LogLevelError:
		logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	}
	slog.SetDefault(logger)

	return relink.Run(cfg)
}
