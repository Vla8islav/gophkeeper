// Package config parsing passed config values
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// OptionsServer configuration parameters for the GophKeeper server.
//
// Values' order of precedence: environment vars, command-line flags, config file, defaults.
type OptionsServer struct {
	ServerAddress    OptionalString `env:"RUN_ADDRESS" json:"server_address" command_arg:"a"`
	DatabaseURI      OptionalString `env:"DATABASE_URI" json:"database_uri" command_arg:"d"`
	MigrationsFolder OptionalString `env:"MIGRATIONS_FOLDER" json:"migrations_folder" command_arg:"m"`
	AuthTokenSecret  OptionalString `env:"AUTH_TOKEN_SECRET" json:"auth_token_secret" command_arg:"s"`
	PublicCertKey    OptionalString `env:"PUBLIC_CERT_KEY" json:"public_cert_key" command_arg:"public-key"`
	PrivateKey       OptionalString `env:"PRIVATE_KEY" json:"private_key" command_arg:"private-key"`
	AuditLogPath     OptionalString `env:"AUDIT_LOG_PATH" json:"audit_log_path" command_arg:"audit-log"`
	Config           OptionalString `env:"CONFIG" json:"-" command_arg:"config"`
}

// ReadFlagsServer reads server configuration from command-line arguments and environment variables.
//
// Precedence: environment variables, command-line flags, config file, defaults.
func ReadFlagsServer(args []string, logger *zap.Logger) (*OptionsServer, error) {
	if logger == nil {
		panic("config server logger is nil")
	}
	cmdOptions, err := getOptionsServer(args, logger)
	if err != nil {
		return nil, fmt.Errorf("read command-line flags: %w", err)
	}
	logSetFlagsServer(cmdOptions, logger)
	envOptions, err := getEnvOptionsServer(logger)
	if err != nil {
		return nil, fmt.Errorf("read environment variables: %w", err)
	}
	logSetEnvServer(envOptions, logger)
	var diskConfigOptions OptionsServer
	if cmdOptions.Config.BeenSet || envOptions.Config.BeenSet {
		// We need to read the config file before assembling the full consensus.
		var configFilename string
		if envOptions.Config.BeenSet && envOptions.Config.Value != "" {
			configFilename = envOptions.Config.Value
		} else if cmdOptions.Config.BeenSet && cmdOptions.Config.Value != "" {
			configFilename = cmdOptions.Config.Value
		}
		diskConfigOptions, err = getDiskConfigOptionsServer(configFilename, logger)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		logConfigOptionsServer(&diskConfigOptions, logger)
	}
	finalOptions := OptionsServer{
		ServerAddress: OptionalString{
			Value:   "localhost:8080",
			BeenSet: false,
		},
		DatabaseURI: OptionalString{
			Value:   "postgres://default_user:default_password@localhost:5432/gophkeeper_db?sslmode=disable",
			BeenSet: false,
		},
		MigrationsFolder: OptionalString{
			Value:   "./migrations",
			BeenSet: false,
		},
		AuthTokenSecret: OptionalString{
			Value:   "super-duper-secret-dev-change-in-prod",
			BeenSet: false,
		},
		PublicCertKey: OptionalString{
			Value:   "",
			BeenSet: false,
		},
		PrivateKey: OptionalString{
			Value:   "",
			BeenSet: false,
		},
		AuditLogPath: OptionalString{
			Value:   "", // empty = audit disabled
			BeenSet: false,
		},
		Config: OptionalString{
			Value:   "",
			BeenSet: false,
		},
	}
	// Environment options have the highest priority,
	// then command-line options, then disk config options.
	mergeOptionsServer(&finalOptions, diskConfigOptions)
	mergeOptionsServer(&finalOptions, *cmdOptions)
	mergeOptionsServer(&finalOptions, *envOptions)
	return &finalOptions, nil
}

func getDiskConfigOptionsServer(filename string, logger *zap.Logger) (OptionsServer, error) {
	if filename == "" {
		return OptionsServer{}, nil
	}
	configBytes, err := os.ReadFile(filename)
	if err != nil {
		return OptionsServer{}, err
	}
	options := OptionsServer{}
	if err = json.Unmarshal(configBytes, &options); err != nil {
		return OptionsServer{}, err
	}
	return options, nil
}

func mergeOptionsServer(mergeInto *OptionsServer, newValues OptionsServer) {
	if newValues.ServerAddress.BeenSet {
		mergeInto.ServerAddress = newValues.ServerAddress
		mergeInto.ServerAddress.BeenSet = true
	}
	if newValues.DatabaseURI.BeenSet {
		mergeInto.DatabaseURI = newValues.DatabaseURI
		mergeInto.DatabaseURI.BeenSet = true
	}
	if newValues.MigrationsFolder.BeenSet {
		mergeInto.MigrationsFolder = newValues.MigrationsFolder
		mergeInto.MigrationsFolder.BeenSet = true
	}
	if newValues.AuthTokenSecret.BeenSet {
		mergeInto.AuthTokenSecret = newValues.AuthTokenSecret
		mergeInto.AuthTokenSecret.BeenSet = true
	}
	if newValues.PublicCertKey.BeenSet {
		mergeInto.PublicCertKey = newValues.PublicCertKey
		mergeInto.PublicCertKey.BeenSet = true
	}
	if newValues.PrivateKey.BeenSet {
		mergeInto.PrivateKey = newValues.PrivateKey
		mergeInto.PrivateKey.BeenSet = true
	}
	if newValues.AuditLogPath.BeenSet {
		mergeInto.AuditLogPath = newValues.AuditLogPath
		mergeInto.AuditLogPath.BeenSet = true
	}
	if newValues.Config.BeenSet {
		mergeInto.Config = newValues.Config
		mergeInto.Config.BeenSet = true
	}
}
func getEnvOptionsServer(logger *zap.Logger) (*OptionsServer, error) {
	opt := OptionsServer{}
	if err := env.Parse(&opt); err != nil {
		return nil, err
	}
	return &opt, nil
}
func getOptionsServer(args []string, logger *zap.Logger) (*OptionsServer, error) {
	opt := &OptionsServer{}
	fs := flag.NewFlagSet("gophkeeper-server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&opt.ServerAddress, "a", "адрес и порт запуска этого сервера")
	fs.Var(&opt.DatabaseURI, "d", "connection string/dsn для postgres базы данных")
	fs.Var(&opt.MigrationsFolder, "m", "относительный путь до миграций, например ./migrations")
	fs.Var(&opt.AuthTokenSecret, "s", "секретный ключ для генерации токенов авторизации")
	fs.Var(&opt.PublicCertKey, "public-key", "путь до публичного ключа")
	fs.Var(&opt.PrivateKey, "private-key", "путь до приватного ключа")
	fs.Var(&opt.AuditLogPath, "audit-log", "путь до файла аудита (JSONL); пусто = выключено")
	fs.Var(&opt.Config, "config", "путь до файла с конфигурацией приложения")
	fs.Var(&opt.Config, "c", "путь до файла с конфигурацией приложения")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opt, nil
}
