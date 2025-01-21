package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/loopwork-ai/emcee/mcp"
)

var rootCmd = &cobra.Command{
	Use:   "emcee [openapi-spec-url]",
	Short: "An MCP server for a given OpenAPI specification",
	Long: `emcee is a CLI tool that provides an MCP stdio transport for a given OpenAPI specification.
It takes an OpenAPI specification URL as input and processes JSON-RPC requests
from stdin, making corresponding API calls and returning JSON-RPC responses to stdout.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			opts := []mcp.ServerOption{
				mcp.WithSpecURL(args[0]),
			}

			if auth != "" {
				opts = append(opts, mcp.WithAuth(auth))
			}

			if verbose {
				opts = append(opts, mcp.WithVerbose(os.Stderr))
			}

			server, err := mcp.NewServer(opts...)
			if err != nil {
				return fmt.Errorf("error creating server: %v", err)
			}

			transport := mcp.NewStdioTransport(server, os.Stdin, os.Stdout, os.Stderr)
			return transport.Run(ctx)
		})

		return g.Wait()
	},
}

var (
	auth    string
	verbose bool
)

func init() {
	rootCmd.Flags().StringVar(&auth, "auth", "", "Authorization header value (e.g. 'Bearer token123' or 'Basic dXNlcjpwYXNz')")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging to stderr")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
