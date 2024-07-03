package cmd

import (
	"os"
	"secret-storage/cli/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "storage",
	Short: "CLI для доступа в хранилище",
	Long:  "CLI реализующий интерфейс для достпа в хранилище",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigFile("config")
	config.InitConfig()
}
