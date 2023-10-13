package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	gmeter "github.com/venti-org/go-meter"
)

var rootCmd = &cobra.Command{
	Use: "gmeter",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	methods := []string{"get", "post", "head", "put", "delete", "patch", "connect", "options", "trace"}

	for _, m := range methods {
		method := m
		var concurrency *int
		var count *int
		var url *string
		var urlsPath *string
		var body *string
		var bodyPath *string
		var bodiesPath *string
		var extraJsonPath *string
		var skipError *bool
		var proxy *string
		var headers *[]string
		var skip *int

		cmd := &cobra.Command{
			Use: method,
			RunE: func(cmd *cobra.Command, args []string) error {
				config := &gmeter.DriverConfig{
					Concurrency: *concurrency,
					Skip:        *skip,
					SkipError:   *skipError,
					ClientConfig: gmeter.ClientConfig{
						Count: *count,
						Proxy: *proxy,
					},
					RequestGeneratorConfig: gmeter.RequestGeneratorConfig{
						Headers:       *headers,
						Method:        strings.ToUpper(method),
						Url:           *url,
						UrlsPath:      *urlsPath,
						Body:          *body,
						BodyPath:      *bodyPath,
						BodiesPath:    *bodiesPath,
						ExtraJsonPath: *extraJsonPath,
					},
				}
				if driver, err := gmeter.NewDriver(config); err != nil {
					return err
				} else {
					var errs []error
					errs = append(errs, driver.Run())
					errs = append(errs, driver.Close())
					return gmeter.GainError(errs)
				}
			},
		}
		rootCmd.AddCommand(cmd)

		concurrency = cmd.PersistentFlags().IntP("concurrency", "c", 1, "")
		count = cmd.PersistentFlags().IntP("client-count", "n", 1, "")
		skip = cmd.PersistentFlags().IntP("skip", "s", 0, "")
		url = cmd.PersistentFlags().StringP("url", "u", "", "")
		urlsPath = cmd.PersistentFlags().String("urls-path", "", "")
		proxy = cmd.PersistentFlags().StringP("proxy", "p", "", "")
		body = cmd.PersistentFlags().StringP("body", "b", "", "")
		bodyPath = cmd.PersistentFlags().String("body-path", "", "")
		bodiesPath = cmd.PersistentFlags().String("bodies-path", "", "")
		extraJsonPath = cmd.PersistentFlags().String("extra-json-path", "", "")
		skipError = cmd.PersistentFlags().Bool("skip-error", false, "")
		headers = cmd.PersistentFlags().StringArrayP("headers", "H", []string{}, "")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		gmeter.ErrPrintln(err.Error())
		os.Exit(1)
	}
}
