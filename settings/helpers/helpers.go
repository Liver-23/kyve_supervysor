package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"cosmossdk.io/log"
)

type Response struct {
	Pool struct {
		Data struct {
			CurrentKey     string `json:"current_key"`
			UploadInterval string `json:"upload_interval"`
			MaxBundleSize  string `json:"max_bundle_size"`
		} `json:"data"`
	} `json:"pool"`
}

var logger = log.NewLogger(os.Stdout)

func CheckBinaryPath(path string) error {
	_, err := exec.LookPath(path)
	if err != nil {
		return err
	}
	return nil
}

func CalculateKeepRecent(maxBundleSize int, uploadInterval int) int {
	return int(
		math.Round(
			float64(maxBundleSize) / float64(uploadInterval) * 60 * 60 * 24 * 7))
}

func CalculateMaxDifference(maxBundleSize int, uploadInterval int) int {
	return int(
		math.Round(
			float64(maxBundleSize) / float64(uploadInterval) * 60 * 60 * 24 * 5))
}

func GetPoolSettings(poolId int, chainId string) ([2]int, error) {
	var poolEndpoint string
	if chainId == "korellia" {
		poolEndpoint = "https://api.korellia.kyve.network/kyve/query/v1beta1/pool/" + strconv.FormatInt(int64(poolId), 10)
	} else if chainId == "kaon-1" {
		poolEndpoint = "https://api-eu-1.kaon.kyve.network/kyve/query/v1beta1/pool/" + strconv.FormatInt(int64(poolId), 10)
	} else if chainId == "kyve-1" {
		poolEndpoint = "https://api-eu-1.kyve.network/kyve/query/v1beta1/pool/" + strconv.FormatInt(int64(poolId), 10)
	} else {
		return [2]int{0, 0}, fmt.Errorf("unknown chainId (needs to be kyve-1, kaon-1 or korellia)")
	}
	response, err := http.Get(poolEndpoint)
	if err != nil {
		logger.Error("API isn't available", err.Error())
		return [2]int{}, err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error("got unexpected response", err.Error())
		return [2]int{}, err
	}

	var resp Response
	err = json.Unmarshal(responseData, &resp)
	if err != nil {
		logger.Error("couldn't unmarshal response", err.Error())
		return [2]int{}, err
	}

	uploadInterval := resp.Pool.Data.UploadInterval
	interval, err := strconv.Atoi(uploadInterval)
	if err != nil {
		logger.Error("couldn't convert uploadInterval to int", err.Error())
		return [2]int{}, err
	}

	maxBundleSize := resp.Pool.Data.MaxBundleSize
	size, err := strconv.Atoi(maxBundleSize)
	if err != nil {
		logger.Error("couldn't convert maxBundleSize to int", err.Error())
		return [2]int{}, err
	}

	return [2]int{size, interval}, nil
}
