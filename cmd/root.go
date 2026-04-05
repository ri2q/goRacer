package cmd

import (
	"maps"
	"os"
	"strings"

	"github.com/ri2q/goRacer/internal/engine"
	"github.com/ri2q/goRacer/internal/input"
	preparer "github.com/ri2q/goRacer/internal/preparer/payload"
	"github.com/spf13/cobra"
)

type rootConfig struct {
	Files      []string
	SleepTime  int64
	Iter       uint32
	Filter     int
	EndStream  bool
	KeyLogPath string
	Proxy      string

	Target      string
	Data        string
	Method      string
	JSON        string
	Cookies     map[string]string
	Headers     map[string]string
	SensHeaders map[string]string
}

var cfg = rootConfig{
	SleepTime: 100,
	Iter:      25,
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "goRacer",
	RunE: func(cmd *cobra.Command, args []string) error {
		isFile := cmd.Flags().Changed("request")
		isUrl := cmd.Flags().Changed("url")

		mu := input.NewMutual()
		mu.EndStream = cfg.EndStream
		mu.Filter = cfg.Filter
		mu.Iter = cfg.Iter
		mu.KeyLogPath = cfg.KeyLogPath
		mu.Proxy = cfg.Proxy
		mu.SleepTime = cfg.SleepTime

		var payloadQueue []*preparer.Payload
		if isFile {
			filesPayloads, err := buildPayloadsFromFiles(&cfg, mu)
			if err != nil {
				cmd.PrintErrln(err)
				return nil
			}
			payloadQueue = append(payloadQueue, filesPayloads...)
		}

		if isUrl {
			urlPayload, err := buildPayloadFromURL(&cfg, mu)
			if err != nil {
				cmd.PrintErrln(err)
				return nil
			}
			payloadQueue = append(payloadQueue, urlPayload)
		}

		if err := engine.Run(payloadQueue); err != nil {
			cmd.PrintErrln(err)
		}

		return nil
	},
}

func buildPayloadsFromFiles(cfg *rootConfig, mu *input.Mutual) ([]*preparer.Payload, error) {
	if len(cfg.Files) == 0 {
		return nil, nil
	}

	payloads := make([]*preparer.Payload, 0, len(cfg.Files))
	for _, file := range cfg.Files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		newFile := input.NewFileHandler()
		newFile.Raw = f
		payload, err := newFile.Parse()
		f.Close()
		if err != nil {
			return nil, err
		}

		mu.Parse(payload)
		payloads = append(payloads, payload)
	}

	return payloads, nil
}

func buildPayloadFromURL(cfg *rootConfig, mu *input.Mutual) (*preparer.Payload, error) {
	if cfg.Target == "" {
		return nil, nil
	}

	u := input.NewUrlHandler()
	u.Target = cfg.Target
	u.SleepTime = cfg.SleepTime
	u.Method = strings.ToUpper(cfg.Method)

	if cfg.Data != "" {
		if u.Method == "" {
			u.Method = "POST"
		}
		u.Data = cfg.Data
	} else if cfg.JSON != "" {
		if u.Method == "" {
			u.Method = "POST"
		}
		u.Json = cfg.JSON
	}

	if cfg.Headers != nil {
		maps.Copy(u.NonSensHeaders, cfg.Headers)
	}
	if cfg.Cookies != nil {
		maps.Copy(u.CookieJar, cfg.Cookies)
	}
	if cfg.SensHeaders != nil {
		maps.Copy(u.SensHeaders, cfg.SensHeaders)
	}

	payload, err := u.Parse()
	if err != nil {
		return nil, err
	}

	mu.Parse(payload)
	return payload, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.Target, "url", "u", "", "Request URL")
	rootCmd.Flags().StringSliceVarP(&cfg.Files, "request", "r", nil, "Path(s) to one or more HTTP request files")
	rootCmd.Flags().Uint32VarP(&cfg.Iter, "iteration", "i", 25, "Number of concurrent racers per request")
	rootCmd.Flags().StringVarP(&cfg.Data, "data", "d", "", "Request body data (form-encoded)")
	rootCmd.Flags().StringVarP(&cfg.JSON, "json", "j", "", "Request body JSON")
	rootCmd.Flags().StringVarP(&cfg.Method, "method", "m", "", "HTTP method (GET, POST, etc)")
	rootCmd.Flags().StringToStringVarP(&cfg.Cookies, "cookie", "c", nil, "Cookie jar values key=value")
	rootCmd.Flags().StringToStringVarP(&cfg.Headers, "header", "x", nil, "Non-sensitive headers key=value")
	rootCmd.Flags().StringToStringVarP(&cfg.SensHeaders, "sensitive-header", "s", nil, "Sensitive headers key=value")
	rootCmd.Flags().IntVarP(&cfg.Filter, "filter", "f", 200, "HTTP status code filter")

	rootCmd.Flags().BoolVarP(&cfg.EndStream, "end-stream", "e", false, "Send requests with END_STREAM flag")
	rootCmd.Flags().Int64VarP(&cfg.SleepTime, "delay", "H", 100, "Delay between iterations in milliseconds")
	rootCmd.Flags().StringVarP(&cfg.KeyLogPath, "tls-key-log", "l", "", "TLS key log file path for debugging")
	rootCmd.Flags().StringVarP(&cfg.Proxy, "proxy", "p", "", "Proxy address type://host:port (e.g. http://192.168.1.1:9090) or just host:port (e.g. 127.0.0.1:9090)")

	rootCmd.MarkFlagsMutuallyExclusive("request", "url")
	rootCmd.MarkFlagsOneRequired("request", "url")
	rootCmd.MarkFlagsMutuallyExclusive("data", "json")
}
