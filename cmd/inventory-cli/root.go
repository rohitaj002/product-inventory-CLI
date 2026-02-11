package main

import (
	"context"
	"fmt"
	"inventory-cli/internal/store"
	"os"

	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	storeType string
	filePath  string
	logLevel  string
	appStore  store.ProductStore
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "inventory-cli",
	Short: "A product inventory management system",
	Long: `Inventory CLI is a tool for managing product inventory.
It supports CRUD operations, bulk import/export, and multiple storage backends.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeApp(cmd.Context())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.inventory-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&storeType, "store", "memory", "storage type (memory|json)")
	rootCmd.PersistentFlags().StringVar(&filePath, "db-file", "products.json", "file path for json store")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug|info|warn|error)")

	viper.BindPFlag("store", rootCmd.PersistentFlags().Lookup("store"))
	viper.BindPFlag("db-file", rootCmd.PersistentFlags().Lookup("db-file"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".inventory-cli")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// keeping silent on config loaded
	}
}

func initializeApp(ctx context.Context) error {
	// Setup Logger
	lvl := slog.LevelInfo
	switch viper.GetString("log-level") {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)

	st := viper.GetString("store")
	fp := viper.GetString("db-file")

	var err error
	appStore, err = store.NewStoreFactory(store.StoreType(st), fp)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	slog.Info("Application initialized", "store", st, "file", fp)
	return nil
}
