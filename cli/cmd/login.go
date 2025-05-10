package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/liriquew/secret_storage/cli/config"
	"github.com/spf13/cobra"
)

type userData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenJWT struct {
	Token string `json:"token,omitempty"`
}

var signIn = &cobra.Command{
	Use:   "signin",
	Short: "Авторизовывает пользователя",
	Run: func(cmd *cobra.Command, args []string) {
		buf, err := json.Marshal(userData{username, password})
		if err != nil {
			fmt.Println(err)
			return
		}

		response, err := http.Post("http://localhost:8080/api/signin", "application/json", bytes.NewBuffer(buf))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			fmt.Printf("Status: %v\n", response.StatusCode)
			if response.Body != nil {
				fmt.Println(response.Body)
			}
			return
		}

		var token tokenJWT
		json.NewDecoder(response.Body).Decode(&token)
		config.SetToken(token.Token)
		fmt.Println("Token:", token.Token)
	},
}

var signUp = &cobra.Command{
	Use:   "signup",
	Short: "Регистрирует пользователя",
	Run: func(cmd *cobra.Command, args []string) {
		buf, err := json.Marshal(userData{username, password})
		if err != nil {
			fmt.Println(err)
			return
		}

		response, err := http.Post("http://localhost:8080/api/signup", "application/json", bytes.NewBuffer(buf))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			fmt.Printf("Status: %v\n", response.StatusCode)
			if response.Body != nil {
				fmt.Println(response.Body)
			}
			return
		}

		var token tokenJWT
		json.NewDecoder(response.Body).Decode(&token)
		config.SetToken(token.Token)
		fmt.Println("Token:", token.Token)
	},
}

var (
	username string
	password string
)

func init() {
	signIn.Flags().StringVarP(&username, "username", "u", "", "Имя пользователя")
	signIn.Flags().StringVarP(&password, "password", "p", "", "Пароль")

	signUp.Flags().StringVarP(&username, "userame", "u", "", "Имя пользователя")
	signUp.Flags().StringVarP(&password, "password", "p", "", "Пароль")

	rootCmd.AddCommand(signIn)
	rootCmd.AddCommand(signUp)
}
