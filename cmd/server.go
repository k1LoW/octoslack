/*
Copyright © 2023 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/k1LoW/octoslack/config"
	"github.com/k1LoW/octoslack/server"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

var (
	configPath string
	port       uint64
	verbose    bool
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Long:  `start server.`,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if e := os.Getenv("OCTOSLACK_CONFIG"); e != "" {
			configPath = e
		}
		if e := os.Getenv("OCTOSLACK_PORT"); e != "" {
			port, err = strconv.ParseUint(e, 10, 64)
		}
		if os.Getenv("OCTOSLACK_VERBOSE") != "" || os.Getenv("DEBUG") != "" {
			verbose = true
		}
		setLogger(verbose)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Load config", slog.String("path", configPath))
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, os.Kill)
		eg, cctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return server.Start(cctx, cfg, port)
		})
		if err := eg.Wait(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&configPath, "config", "c", "config.yml", "config path")
	serverCmd.Flags().BoolVarP(&verbose, "verbose", "", false, "show verbose log")
	serverCmd.Flags().Uint64VarP(&port, "port", "p", 8080, "listen port")
}

func setLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.HandlerOptions{
		AddSource: verbose,
		Level:     level,
	}.NewJSONHandler(os.Stdout))
	slog.SetDefault(logger)
}
