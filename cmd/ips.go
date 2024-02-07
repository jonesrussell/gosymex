/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jonesrussell/gosymex/metaresponse"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

const fileName = "sqlite.db"

// ipsCmd represents the ips command
var ipsCmd = &cobra.Command{
	Use:   "ips",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get("https://api.github.com/meta")
		if err != nil {
			fmt.Println("Error fetching meta info:", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		var meta metaresponse.Ipv4cidr
		if err := json.Unmarshal(body, &meta); err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		os.Remove(fileName)

		db, err := sql.Open("sqlite3", fileName)
		if err != nil {
			log.Fatal(err)
		}

		websiteRepository := metaresponse.NewSQLiteRepository(db)

		if err := websiteRepository.Migrate(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(ipsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ipsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ipsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
