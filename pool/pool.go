package pool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/KYVENetwork/supervysor/types"
)

// GetPoolHeight retrieves the KYVE pool height by using a list of endpoints (& optionally fallback endpoints)
// based on the provided chain and pool ID.
func GetPoolHeight(chainId string, poolId int, fallbackEndpoints string) (int, error) {
	var endpoints []string
	var err error

	if chainId == "korellia" {
		endpoints = types.KorelliaEndpoints
	} else if chainId == "kaon-1" {
		endpoints = types.KaonEndpoints
	} else if chainId == "kyve-1" {
		endpoints = types.MainnetEndpoints
	} else {
		return 0, fmt.Errorf("unknown chainId")
	}

	for _, endpoint := range append(endpoints, strings.Split(fallbackEndpoints, ",")...) {
		if endpoint != "" {
			if height, err := requestPoolHeight(poolId, endpoint); err == nil {
				return height, err
			}
		}
	}
	return 0, err
}

// requestPoolHeight retrieves KYVE pool height by making an GET request to the given endpoint.
func requestPoolHeight(poolId int, endpoint string) (int, error) {
	poolEndpoint := endpoint + "/kyve/query/v1beta1/pool/" + strconv.FormatInt(int64(poolId), 10)

	response, err := http.Get(poolEndpoint)
	if err != nil {
		return 0, fmt.Errorf("failed requesting KYVE endpoint: %s", err)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("failed reading KYVE endpoint response: %s", err)
	}

	var resp types.SettingsResponse
	err = json.Unmarshal(responseData, &resp)
	if err != nil {
		return 0, fmt.Errorf("failed unmarshalling KYVE endpoint response: %s", err)
	}

	var poolHeight int
	currentKey := resp.Pool.Data.CurrentKey

	if currentKey == "" {
		startKey := resp.Pool.Data.StartKey
		poolHeight, err = strconv.Atoi(startKey)
		if err != nil {
			return 0, fmt.Errorf("could not convert poolHeight from start_key to int: %s", err)
		}
	} else {
		poolHeight, err = strconv.Atoi(currentKey)
		if err != nil {
			return 0, fmt.Errorf("could not convert poolHeight from current_key to int: %s", err)
		}
	}

	return poolHeight, err
}
