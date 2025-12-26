/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"aquascore/api/internal/crawler"
	"aquascore/api/internal/crawler/persistence"
	"aquascore/api/internal/db/mongo"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// crawlerCmd represents the crawler command
var crawlerCmd = &cobra.Command{
	Use:   "crawler",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		year, err := cmd.Flags().GetString("year")
		if err != nil {
			return fmt.Errorf("get target url fail: %w", err)
		}
		var targetURL string
		if year == "" {
			targetURL = "https://ctsa.utk.com.tw/CTSA/public/race/game_data.aspx"
		} else {
			targetURL = fmt.Sprintf("https://ctsa.utk.com.tw/CTSA_%s/public/race/game_data.aspx", year)
		}

		// db connection
		dbCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		closeFunc, err := mongo.IniMongodb(
			dbCtx,
			viper.GetString("database.uri"),
			viper.GetString("database.db"),
		)
		if err != nil {
			cancel()
			return fmt.Errorf("init mongodb fail: %w", err)
		}
		defer func() {
			closeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			if err := closeFunc(closeCtx); err != nil {
				fmt.Printf("close mongodb fail: %v\n", err)
			}
		}()
		cancel()

		var crawlerPersistence crawler.Persistence
		mongo.InjectStore(func(s *mongo.Stores) {
			crawlerPersistence = persistence.NewMongoPersistence(s.RaceStore, s.CrawlLogStore)
		})

		crawler, err := crawler.NewCtsaCrawler(
			crawler.WithBaseURL(targetURL),
			crawler.WithPersistence(crawlerPersistence))
		if err != nil {
			return fmt.Errorf("init ctsa crawler fail: %w", err)
		}

		crawlCtx, crawlCancel := context.WithCancel(context.Background())
		defer crawlCancel()
		err = crawler.Crawl(crawlCtx)
		if err != nil {
			return fmt.Errorf("crawler fail: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(crawlerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// crawlerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// crawlerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	crawlerCmd.Flags().String("year", "", "year to crawl")
}
