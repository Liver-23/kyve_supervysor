package node

import (
	"fmt"

	"github.com/KYVENetwork/supervysor/node/helpers"
)

func InitialStart(logFile string, binaryPath string, addrBookPath string, seeds string, homeDir string) error {
	logger = helpers.InitLogger(logFile)

	logger.Info("starting initially")
	process, err := startNode(logFile, true, binaryPath, addrBookPath, seeds, homeDir)
	if err != nil {
		return fmt.Errorf("could not start node initially: %s", err)
	}

	logger.Info("initial process started", "pId", process.Pid)

	Process.Id = process.Pid
	Process.GhostMode = false

	return nil
}

func EnableGhostMode(logFile string, binaryPath string, addrBookPath string, homeDir string) error {
	logger = helpers.InitLogger(logFile)

	if !Process.GhostMode {
		if err := ShutdownNode(); err != nil {
			logger.Error("could not shutdown node", "err", err)
		}

		process, err := startGhostNode(logFile, binaryPath, addrBookPath, homeDir)
		if err != nil {
			return fmt.Errorf("Ghost Mode enabling failed: %s", err)
		} else {
			if process != nil && process.Pid > 0 {
				Process.Id = process.Pid
				Process.GhostMode = true
				logger.Info("node started in Ghost Mode")
			} else {
				return fmt.Errorf("Ghost Mode enabling failed: process is not defined")
			}
		}
	}
	return nil
}

func EnableNormalMode(logFile string, binaryPath string, addrBookPath string, seeds string, homeDir string) error {
	logger = helpers.InitLogger(logFile)

	if Process.GhostMode {
		if err := ShutdownNode(); err != nil {
			logger.Error("could not shutdown node", "err", err)
		}

		process, err := startNode(logFile, false, binaryPath, addrBookPath, seeds, homeDir)
		if err != nil {
			return fmt.Errorf("Ghost Mode disabling failed: %s", err)
		} else {
			if process != nil && process.Pid > 0 {
				Process.Id = process.Pid
				Process.GhostMode = false
				logger.Info("Node started in Normal Mode", "pId", process.Pid)
			} else {
				return fmt.Errorf("Ghost Mode disabling failed: process is not defined")
			}
		}
	}
	return nil
}
