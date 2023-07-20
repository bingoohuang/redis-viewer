package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bingoohuang/redis-viewer/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "redis-viewer",
	Short: "view redis data in terminal.",
	Long:  `Redis Viewer is a tool to view redis data in terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.Get()

		model, err := internal.New(config)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			model.Close()
		}()

		p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			log.Fatal("start failed: ", err)
		}
	},
}

// main adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().
		StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.redis-viewer.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".redis-viewer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".redis-viewer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	viper.SetDefault("mode", "client")
	viper.SetDefault("count", internal.DefaultCount)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
