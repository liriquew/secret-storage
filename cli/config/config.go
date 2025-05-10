package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config") // Название файла без расширения
	viper.SetConfigType("yaml")   // Тип файла конфигурации
	viper.AddConfigPath("./cli")  // Путь к файлу конфигурации
	err := viper.ReadInConfig()   // Чтение конфигурации
	if err != nil {
		fmt.Printf("Fatal error config file: %s", err)
	}

}

func GetToken() string {
	return viper.GetString("token")
}

func SetToken(token string) {
	viper.Set("token", token)
	viper.WriteConfig()
}
