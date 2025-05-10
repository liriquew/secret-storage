package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

type SecretInfo struct {
	Parts     int `json:"parts"`
	Threshold int `json:"threshold"`
}

var seal = &cobra.Command{
	Use:   "seal [-p partsNum] [-t threshold]",
	Short: "Возвращает части мастер ключа",
	Run: func(cmd *cobra.Command, args []string) {
		parts, err := cmd.Flags().GetInt("parts")
		if err != nil {
			fmt.Println(err)
			return
		}

		threshold, err := cmd.Flags().GetInt("threshold")
		if err != nil {
			fmt.Println(err)
			return
		}

		if parts < 2 || threshold < 2 {
			fmt.Println("Неверно указаны аргументы: 2 <= parts, threshold  <= 256")
			return
		}

		data := &SecretInfo{parts, threshold}
		buf, _ := json.Marshal(data)

		client := &http.Client{}

		req, err := http.NewRequest("GET", baseURL+"master", bytes.NewBuffer(buf))
		if err != nil {
			fmt.Println(err)
			return
		}

		response, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer response.Body.Close()

		masterParts := make([]string, 0)
		json.NewDecoder(response.Body).Decode(&masterParts)

		fmt.Println("Части мастер ключа")
		for _, p := range masterParts {
			fmt.Println(p)
		}
	},
}

var unseal = &cobra.Command{
	Use:   "unseal",
	Short: "Расшифровывает хранилище по частям мастер ключа",
	Run: func(cmd *cobra.Command, args []string) {

		parts := make([]string, 0)

		for {
			part := ""
			fmt.Scanln(&part)
			if part == "" {
				break
			}
			parts = append(parts, part)
		}

		buf, _ := json.Marshal(parts)

		response, err := http.Post(baseURL+"unseal", "application/json", bytes.NewBuffer(buf))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			fmt.Printf("Status: %v\n", response.StatusCode)
			buf, _ := io.ReadAll(response.Body)
			if len(buf) != 0 {
				fmt.Printf("%s\n", buf)
			}
			return
		}

		fmt.Println("ОК")
	},
}

func init() {
	seal.Flags().IntP("parts", "p", -1, "Общее число генерируемых ключей")
	seal.Flags().IntP("threshold", "t", -1, "Число ключей, необходимое для разблокировки хранилища")

	rootCmd.AddCommand(seal)
	rootCmd.AddCommand(unseal)
}
