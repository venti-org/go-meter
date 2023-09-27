package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gmeter "github.com/venti-org/go-meter"
)

var concurrency int
var count int
var api string
var path string
var extraPath string
var skipError bool

var rootCmd = &cobra.Command{
	Use: "go-meter",
	Run: func(cmd *cobra.Command, args []string) {
		config := &gmeter.DriverConfig{
			Concurrency:   concurrency,
			Count:         count,
			Api:           api,
			Path:          path,
			ExtraJsonPath: extraPath,
            SkipError:     skipError,
		}
		driver := gmeter.NewDriver(config)
		driver.Run()
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", 1, "")
	rootCmd.PersistentFlags().IntVarP(&count, "count", "n", 1, "")
	rootCmd.PersistentFlags().StringVar(&api, "api", "http://localhost:8080/api/v1/parser", "")
	rootCmd.PersistentFlags().StringVarP(&path, "path", "p", "", "")
	rootCmd.PersistentFlags().StringVar(&extraPath, "extra-path", "", "")
	rootCmd.PersistentFlags().BoolVar(&skipError, "skip-error", false, "")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
