// Package config parsing passed config values
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// OptionsClient configuration parameters for the GophKeeper client.
//
// order of precedence: environment vars, command-line flags, config file, defaults.
type OptionsClient struct {
	ServerAddress OptionalString `env:"SERVER_ADDRESS" json:"server_address"`
	CACertPath    OptionalString `env:"CA_CERT" json:"ca_cert"`
	TokenFile     OptionalString `env:"TOKEN_FILE" json:"token_file"`
	Config        OptionalString `env:"CONFIG" json:"-"`
	logger        *zap.Logger
}

func logSetFlagsClient(options *OptionsClient) {
	if options == nil {
		return
	}
	fields := make([]zap.Field, 0)
	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("-a", options.ServerAddress.Value))
	}
	if options.CACertPath.BeenSet {
		fields = append(fields, zap.String("-ca", options.CACertPath.Value))
	}
	if options.TokenFile.BeenSet {
		fields = append(fields, zap.String("-t", options.TokenFile.Value))
	}
	if options.Config.BeenSet {
		fields = append(fields, zap.String("-config", options.Config.Value))
	}
	if len(fields) == 0 {
		options.logger.Info("no command-line flags were set")
		return
	}
	options.logger.Info("command line options", fields...)
}

func logSetEnvClient(options *OptionsClient) {
	if options == nil {
		return
	}
	fields := make([]zap.Field, 0)
	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("SERVER_ADDRESS", options.ServerAddress.Value))
	}
	if options.CACertPath.BeenSet {
		fields = append(fields, zap.String("CA_CERT", options.CACertPath.Value))
	}
	if options.TokenFile.BeenSet {
		fields = append(fields, zap.String("TOKEN_FILE", options.TokenFile.Value))
	}
	if options.Config.BeenSet {
		fields = append(fields, zap.String("CONFIG", options.Config.Value))
	}
	if len(fields) == 0 {
		options.logger.Info("no environment variables were set")
		return
	}
	options.logger.Info("environment variables", fields...)
}

func logConfigOptionsClient(options *OptionsClient) {
	if options == nil {
		return
	}
	fields := make([]zap.Field, 0)
	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("server_address", options.ServerAddress.Value))
	}
	if options.CACertPath.BeenSet {
		fields = append(fields, zap.String("ca_cert", options.CACertPath.Value))
	}
	if options.TokenFile.BeenSet {
		fields = append(fields, zap.String("token_file", options.TokenFile.Value))
	}
	if len(fields) == 0 {
		options.logger.Info("no config file options were set")
		return
	}
	options.logger.Info("config file options", fields...)
}

// ReadFlagsClient reads client configuration from command-line arguments and environment variables.
//
// Precedence: environment variables, command-line flags, config file, defaults.
func ReadFlagsClient(args []string, logger *zap.Logger) (*OptionsClient, []string, error) {
	if logger == nil {
		panic("config client logger is nil")
	}
	cmdOptions, rest, err := getOptionsClient(args, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("read command-line flags: %w", err)
	}
	logSetFlagsClient(cmdOptions)
	envOptions, err := getEnvOptionsClient(logger)
	if err != nil {
		return nil, nil, fmt.Errorf("read environment variables: %w", err)
	}
	logSetEnvClient(envOptions)
	var diskConfigOptions OptionsClient
	if cmdOptions.Config.BeenSet || envOptions.Config.BeenSet {
		var configFilename string
		if envOptions.Config.BeenSet && envOptions.Config.Value != "" {
			configFilename = envOptions.Config.Value
		} else if cmdOptions.Config.BeenSet && cmdOptions.Config.Value != "" {
			configFilename = cmdOptions.Config.Value
		}
		diskConfigOptions, err = getDiskConfigOptionsClient(configFilename, logger)
		if err != nil {
			return nil, nil, fmt.Errorf("read config file: %w", err)
		}
		logConfigOptionsClient(&diskConfigOptions)
	}
	finalOptions := OptionsClient{
		ServerAddress: OptionalString{
			Value:   "https://localhost:8080",
			BeenSet: false,
		},
		CACertPath: OptionalString{
			Value:   "",
			BeenSet: false,
		},
		TokenFile: OptionalString{
			Value:   defaultClientTokenFile(),
			BeenSet: false,
		},
		logger: logger,
	}
	mergeOptionsClient(&finalOptions, diskConfigOptions)
	mergeOptionsClient(&finalOptions, *cmdOptions)
	mergeOptionsClient(&finalOptions, *envOptions)
	return &finalOptions, rest, nil
}

func getDiskConfigOptionsClient(filename string, logger *zap.Logger) (OptionsClient, error) {
	if filename == "" {
		return OptionsClient{logger: logger}, nil
	}
	configBytes, err := os.ReadFile(filename)
	if err != nil {
		return OptionsClient{logger: logger}, err
	}
	options := OptionsClient{
		logger: logger,
	}
	if err = json.Unmarshal(configBytes, &options); err != nil {
		return OptionsClient{logger: logger}, err
	}
	return options, nil
}

func mergeOptionsClient(mergeInto *OptionsClient, newValues OptionsClient) {
	if newValues.ServerAddress.BeenSet {
		mergeInto.ServerAddress = newValues.ServerAddress
		mergeInto.ServerAddress.BeenSet = true
	}
	if newValues.CACertPath.BeenSet {
		mergeInto.CACertPath = newValues.CACertPath
		mergeInto.CACertPath.BeenSet = true
	}
	if newValues.TokenFile.BeenSet {
		mergeInto.TokenFile = newValues.TokenFile
		mergeInto.TokenFile.BeenSet = true
	}
	if newValues.Config.BeenSet {
		mergeInto.Config = newValues.Config
		mergeInto.Config.BeenSet = true
	}
}

func getEnvOptionsClient(logger *zap.Logger) (*OptionsClient, error) {
	opt := OptionsClient{
		logger: logger,
	}
	if err := env.Parse(&opt); err != nil {
		return nil, err
	}
	return &opt, nil
}

func getOptionsClient(args []string, logger *zap.Logger) (*OptionsClient, []string, error) {
	opt := &OptionsClient{
		logger: logger,
	}
	fs := flag.NewFlagSet("gophkeeper-client", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&opt.ServerAddress, "base-url", "базовый URL сервера, например https://localhost:8080")
	fs.Var(&opt.CACertPath, "custom-certificate", "путь до CA-сертификата для проверки сервера (необязательно)")
	fs.Var(&opt.TokenFile, "token", "путь до файла с токеном авторизации")
	fs.Var(&opt.Config, "config", "путь до файла с конфигурацией приложения")
	fs.Var(&opt.Config, "c", "путь до файла с конфигурацией приложения")
	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}
	return opt, fs.Args(), nil
}

func defaultClientTokenFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gophkeeper_token"
	}
	return filepath.Join(home, ".gophkeeper", "token")
}
