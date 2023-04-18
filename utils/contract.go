package utils

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

/*
`LoadContract` loads a contract interface given the following parameters.

	`apiKey`is only required when the `rawClientUrl` requires an API key.
	`concat` checks whether the `rawClientUrl` and the `apiKey` should be concatenated or not. if false, it just returns `rawClientUrl` when connecting to the client.
*/
func LoadContract(
	apiKey string,
	concat bool,
	rawClientUrl string,
	abiPath string,
	contractAddress string,
	caller bind.ContractCaller,
	transactor bind.ContractTransactor,
	filterer bind.ContractFilterer,
) (*bind.BoundContract, error) {
	var getApiKey string
	if apiKey != "" {
		getApiKey = os.Getenv(apiKey)
	}

	// connects to the a client given the `rawClientUrl`. if `concat` is true, it concatenates the `rawClientUrl` and the `apiKey`.
	var client *ethclient.Client
	var err error
	if concat {
		client, err = ethclient.Dial(rawClientUrl + getApiKey)
		if err != nil {
			return nil, err
		}
	} else {
		client, err = ethclient.Dial(rawClientUrl)
		if err != nil {
			return nil, err
		}
	}

	abiContentBytes, err := os.ReadFile(abiPath)
	if err != nil {
		return nil, err
	}

	// converts the array of bytes obtained from reading the abi to a string
	abiContent := string(abiContentBytes)

	// contract address on the specific chain
	addr := common.HexToAddress(contractAddress)
	abi, err := abi.JSON(strings.NewReader(abiContent))
	if err != nil {
		return nil, err
	}

	// checks if `caller`, `transactor` or `filterer` are present. if not, default to `client`.
	if caller == nil {
		caller = client
	}
	if transactor == nil {
		transactor = client
	}
	if filterer == nil {
		filterer = client
	}

	contract := bind.NewBoundContract(addr, abi, caller, transactor, filterer)
	return contract, nil
}
