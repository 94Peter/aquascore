/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"aquascore/internal/db/mongo"
	"aquascore/internal/server"
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the AquaScore API server",
	Long: `Starts the AquaScore API server which provides RESTful endpoints
and communicates with the Python analysis service via gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tp, err := initTracer()
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()

		ctx := cmd.Context()

		mongoURI := viper.GetString("database.uri")
		dbName := viper.GetString("database.db")

		closeDB, err := mongo.IniMongodb(ctx, mongoURI, dbName)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer closeDB(ctx)

		var store *mongo.Stores
		mongo.InjectStore(func(s *mongo.Stores) {
			store = s
		})

		port := viper.GetString("http.port")
		addr := fmt.Sprintf(":%s", port)
		analysisServiceAddr := viper.GetString("grpc.analysis.addr")
		server, err := server.NewHttpServer(store, analysisServiceAddr)
		if err != nil {
			return fmt.Errorf("failed to create http server: %w", err)
		}
		fmt.Printf("Starting AquaScore API server on %s...\n", addr)
		return server.Start(addr)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringP("port", "p", "8080", "Port to run the server on")
	viper.BindPFlag("server.port", serverCmd.PersistentFlags().Lookup("port"))
}
