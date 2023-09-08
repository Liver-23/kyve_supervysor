package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/KYVENetwork/supervysor/cmd/supervysor/helpers"
	"github.com/KYVENetwork/supervysor/types"

	"github.com/KYVENetwork/supervysor/settings"

	"github.com/pelletier/go-toml/v2"

	"github.com/spf13/cobra"
)

var (
	abciEndpoint      string
	binaryPath        string
	chainId           string
	fallbackEndpoints string
	homePath          string
	metrics           bool
	poolId            int
	seeds             string
	pruningInterval   int

	cfg types.SupervysorConfig
)

func init() {
	initCmd.Flags().StringVar(&chainId, "chain-id", "", "KYVE chain-id")
	if err := initCmd.MarkFlagRequired("chain-id"); err != nil {
		panic(fmt.Errorf("flag 'chain-id' should be required: %w", err))
	}

	initCmd.Flags().StringVar(&binaryPath, "binary-path", "", "path to chain binaries or cosmovisor (e.g. /root/go/bin/cosmovisor)")
	if err := initCmd.MarkFlagRequired("binary-path"); err != nil {
		panic(fmt.Errorf("flag 'binary-path' should be required: %w", err))
	}

	initCmd.Flags().StringVar(&homePath, "home-path", "", "path to home directory (e.g. /root/.osmosisd)")
	if err := initCmd.MarkFlagRequired("home-path"); err != nil {
		panic(fmt.Errorf("flag 'home-path' should be required: %w", err))
	}

	initCmd.Flags().IntVar(&poolId, "pool-id", 0, "KYVE pool-id")
	if err := initCmd.MarkFlagRequired("pool-id"); err != nil {
		panic(fmt.Errorf("flag 'pool-id' should be required: %w", err))
	}

	initCmd.Flags().StringVar(&seeds, "seeds", "", "seeds for the node to connect")
	if err := initCmd.MarkFlagRequired("seeds"); err != nil {
		panic(fmt.Errorf("flag 'seeds' should be required: %w", err))
	}

	initCmd.Flags().StringVar(&fallbackEndpoints, "fallback-endpoints", "", "additional endpoints to query KYVE pool height")

	initCmd.Flags().IntVar(&pruningInterval, "pruning-interval", 24, "block-pruning interval (hours)")

	initCmd.Flags().BoolVar(&metrics, "metrics", true, "exposing Prometheus metrics (true or false)")

	initCmd.Flags().StringVar(&abciEndpoint, "abci-endpoint", "http://127.0.0.1:26657", "ABCI Endpoint to request node information")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize supervysor",
	RunE: func(cmd *cobra.Command, args []string) error {
		return InitializeSupervysor()
	},
}

// InitializeSupervysor initializes the required supervysor config and performs some basic checks.
func InitializeSupervysor() error {
	if homePath == "" {
		logger.Error("home directory can not be empty")
		return fmt.Errorf("empty home directory path")
	}

	if pruningInterval <= 6 {
		logger.Error("pruning-interval should be higher than 6 hours")
	}

	if err := settings.InitializeSettings(binaryPath, homePath, poolId, false, seeds, chainId, fallbackEndpoints); err != nil {
		logger.Error("could not initialize settings", "err", err)
		return err
	}
	logger.Info("successfully initialized settings")

	configPath, err := helpers.GetSupervysorDir()
	if err != nil {
		logger.Error("could not get supervysor directory path", "err", err)
		return err
	}

	if _, err = os.Stat(configPath + "/config.toml"); err == nil {
		logger.Info(fmt.Sprintf("supervysor was already initialized and is editable under %s/config.toml", configPath))
		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		if _, err = os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(configPath, 0o755)
			if err != nil {
				return err
			}
		}
		logger.Info("initializing supverysor...")

		config := types.SupervysorConfig{
			ABCIEndpoint:        abciEndpoint,
			BinaryPath:          binaryPath,
			ChainId:             chainId,
			HomePath:            homePath,
			Interval:            10,
			Metrics:             metrics,
			PoolId:              poolId,
			Seeds:               seeds,
			FallbackEndpoints:   fallbackEndpoints,
			PruningInterval:     pruningInterval,
			HeightDifferenceMax: settings.Settings.MaxDifference,
			HeightDifferenceMin: settings.Settings.MaxDifference / 2,
			StateRequests:       false,
		}
		b, err := toml.Marshal(config)
		if err != nil {
			logger.Error("could not unmarshal config", "err", err)
			return err
		}

		err = os.WriteFile(configPath+"/config.toml", b, 0o755)
		if err != nil {
			logger.Error("could not write config file", "err", err)
			return err
		}

		_, err = getSupervysorConfig()
		if err != nil {
			logger.Error("could not load config file", "err", err)
			return err
		}

		logger.Info(fmt.Sprintf("successfully initialized: config available at %s/config.toml", configPath))
		return nil
	} else {
		logger.Error("could not get supervysor directory")
		return err
	}
}

// getSupervysorConfig returns the supervysor config.toml file.
func getSupervysorConfig() (*types.SupervysorConfig, error) {
	configPath, err := helpers.GetSupervysorDir()
	if err != nil {
		return nil, fmt.Errorf("could not get supervysor directory path: %s", err)
	}

	data, err := os.ReadFile(configPath + "/config.toml")
	if err != nil {
		return nil, fmt.Errorf("could not find config. Please initialize again: %s", err)
	}

	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not unsmarshal config: %s", err)
	}

	return &cfg, nil
}
