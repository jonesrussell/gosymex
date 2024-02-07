/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

type MetaResponse struct {
	Actions []string `json:"actions"`
}

type IPRange struct {
	Range string
}

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

		var meta MetaResponse
		if err := json.Unmarshal(body, &meta); err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		// Save IP ranges to SQLite
		saveIPRangesToSQLite(meta.Actions)
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

func saveIPRangesToSQLite(ipRanges []string) {
	db, err := sql.Open("sqlite3", "./data/ips.db")
	if err != nil {
		fmt.Println("Error opening SQLite database:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ip_ranges (range TEXT)`)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO ip_ranges(range) VALUES(?)")
	if err != nil {
		fmt.Println("Error preparing statement:", err)
		return
	}

	for _, ipRange := range ipRanges {
		if _, err := stmt.Exec(ipRange); err != nil {
			fmt.Println("Error executing statement:", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("Error committing transaction:", err)
		return
	}

	fmt.Println("IP ranges saved successfully.")
}
