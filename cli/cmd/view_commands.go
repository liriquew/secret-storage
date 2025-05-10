package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type BucketInfo struct {
	Buckets []string `json:"buckets"`
	KVs     []KV     `json:"kvs"`
}

type BucketFullInfo struct {
	Buckets map[string]*BucketFullInfo `json:"buckets"`
	Kvs     []KV                       `json:"kvs"`
}

func showBuckets(bucket *BucketFullInfo, indent string) {
	for _, kv := range bucket.Kvs {
		fmt.Printf("%sKEYVAL:  %s - %s\n", indent, kv.Key, kv.Value)
	}

	for bName, b := range bucket.Buckets {
		fmt.Printf("%sBUCKET:  %s\n", indent, bName)
		showBuckets(b, indent+"  ")
	}
}

var listBucket = &cobra.Command{
	Use:   "list [-p path] [-r]",
	Short: "Возвращает элементы в бакете",
	Run: func(cmd *cobra.Command, args []string) {
		isRecursion, _ := cmd.Flags().GetBool("recursion")
		var url string
		if isRecursion {
			url = "reclist/" + storagePath
		} else {
			url = "list/" + storagePath
		}

		response, err := prepareRequest("GET", url, nil, true)
		if err != nil {
			fmt.Println("Ошибка при выполнении запроса:", err)
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

		if isRecursion {
			var data BucketFullInfo
			json.NewDecoder(response.Body).Decode(&data)

			showBuckets(&data, "")
		} else {
			var data BucketInfo
			json.NewDecoder(response.Body).Decode(&data)

			fmt.Println("Buckets:")
			for _, bucket := range data.Buckets {
				fmt.Printf("\t%s\n", bucket)
			}

			fmt.Println("Key-value")
			for _, kv := range data.KVs {
				fmt.Printf("\t%s - %s\n", kv.Key, kv.Value)
			}
		}
	},
}

func init() {
	listBucket.Flags().StringVarP(&storagePath, "path", "p", "", "Путь до значения в хранилище")
	listBucket.Flags().BoolP("recursion", "r", false, "Рекурсивное отображение")

	rootCmd.AddCommand(listBucket)
}
