package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var get = &cobra.Command{
	Use:   "get [-k key] [-p path]",
	Short: "Возвращает значение связанное с ключом",
	Run: func(cmd *cobra.Command, args []string) {
		response, err := prepareRequest("GET", "secrets/"+storagePath+key, nil)
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

		var data KV
		json.NewDecoder(response.Body).Decode(&data)
		fmt.Printf("Key:\t%s\nValue:\t%s\n", data.Key, data.Value)
	},
}

var set = &cobra.Command{
	Use:   "set [-k key] [-v value] [-p path]",
	Short: "Добавляет в хранилище пару ключ-значение по пути path",
	Run: func(cmd *cobra.Command, args []string) {
		data := KV{key, value}

		response, err := prepareRequest("POST", "secrets/"+storagePath, &data)
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

		json.NewDecoder(response.Body).Decode(&data)
		fmt.Printf("Key:\t%s\nValue:\t%s\n", data.Key, data.Value)
	},
}

var delete = &cobra.Command{
	Use:   "del [-k key] [-p path]",
	Short: "Удаляет значение связанное с ключом",
	Run: func(cmd *cobra.Command, args []string) {
		response, err := prepareRequest("DELETE", "secrets/"+storagePath+key, nil)
		if err != nil {
			fmt.Println(err)
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

	},
}

var (
	key         string
	value       string
	storagePath string
)

func init() {
	get.Flags().StringVarP(&key, "key", "k", "", "Ключ, по которому нужно получить/установить значение")
	get.Flags().StringVarP(&storagePath, "path", "p", "", "Путь до значения в хранилище")

	delete.Flags().StringVarP(&key, "key", "k", "", "Ключ, по которому нужно получить/установить значение")
	delete.Flags().StringVarP(&storagePath, "path", "p", "", "Путь до значения в хранилище")

	set.Flags().StringVarP(&key, "key", "k", "", "Ключ, по которому нужно получить/установить значение")
	set.Flags().StringVarP(&value, "value", "v", "", "Значение, которое надо установить по ключу")
	set.Flags().StringVarP(&storagePath, "path", "p", "", "Путь до значения в хранилище")

	rootCmd.AddCommand(get)
	rootCmd.AddCommand(set)
	rootCmd.AddCommand(delete)
}
