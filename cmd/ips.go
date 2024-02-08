/*
Copyright © 2024 Russell Jones
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/jonesrussell/gosymex/metaresponse"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

const fileName = "data/actions.db"

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

		var meta metaresponse.Response
		if err := json.Unmarshal(body, &meta); err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		validRanges := GetValidIPv4Ranges(meta.Actions)
		fmt.Println("Valid IP ranges:", validRanges)

		db, err := sql.Open("sqlite3", fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		ipRepository := metaresponse.NewSQLiteRepository(db)

		if err := ipRepository.Migrate(); err != nil {
			log.Fatal(err)
		}

		InsertValidIPv4Ranges(ipRepository, validRanges)
	},
}

func init() {
	rootCmd.AddCommand(ipsCmd)
}

func GetValidIPv4Ranges(ipRanges []string) []string {
	// Compile a regex pattern for IP ranges
	ipRangePattern := `^(?:\d{1,3}\.){3}\d{1,3}(?:/\d+)?$`
	re := regexp.MustCompile(ipRangePattern)

	// Collect valid IP ranges
	var validRanges []string
	for _, ipRange := range ipRanges {
		if re.MatchString(ipRange) {
			validRanges = append(validRanges, ipRange)
		}
	}

	return validRanges
}

func InsertValidIPv4Ranges(db *metaresponse.SQLiteRepository, validRanges []string) {
	// Loop through the valid IP ranges and insert them into the database
	for _, ipRange := range validRanges {
		_, err := db.Create(metaresponse.Ipv4cidr{CIDR: ipRange})
		if err != nil {
			log.Fatal(err)
		}
	}
}
