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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"github.com/k1LoW/octoslack/config"
	"github.com/k1LoW/octoslack/transformer"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	endpoint string
	method   string
	headers  []string
)

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "transform payload",
	Long:  `transform payload.`,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if e := os.Getenv("OCTOSLACK_CONFIG"); e != "" && configPath == "" {
			configPath = e
		}
		return nil
	},
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		tr := transformer.New(cfg)
		var b []byte
		if len(args) == 0 {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return errors.New("need payload file or STDIN")
			}
			b, err = io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		} else {
			b, err = os.ReadFile(args[0])
			if err != nil {
				return err
			}
		}
		req, err := http.NewRequest(method, endpoint, bytes.NewReader(b))
		if err != nil {
			return err
		}
		if len(headers) > 0 {
			tp := textproto.NewReader(bufio.NewReader(strings.NewReader(strings.Join(headers, "\r\n") + "\r\n\r\n")))
			mimeHeader, err := tp.ReadMIMEHeader()
			if err != nil {
				return err
			}
			req.Header = http.Header(mimeHeader)
		}
		out, err := tr.Transform(req)
		if err != nil {
			return err
		}
		defer out.Body.Close()
		ob, err := io.ReadAll(out.Body)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := json.Indent(&buf, ob, "", "  "); err != nil {
			return err
		}
		fmt.Println(buf.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().StringVarP(&configPath, "config", "c", "config.yml", "config path")
	catCmd.Flags().StringVarP(&endpoint, "endpoint", "", "https://octoslack.example.com/services/XXX/YYY", "request endpoint")
	catCmd.Flags().StringVarP(&method, "method", "", http.MethodPost, "request method")
	catCmd.Flags().StringSliceVarP(&headers, "header", "H", []string{}, "request header")
	catCmd.Flags().BoolVarP(&verbose, "verbose", "", false, "show verbose log")
}
