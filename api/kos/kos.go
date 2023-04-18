package kos

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// loads the Key Of Salvation contract (ETH mainnet)
func loadKOS(caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	alchemyEthKey := os.Getenv("ALCHEMY_ETH_API_KEY")

	// connects to the ETH Mainnet
	client, err := ethclient.Dial(fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", alchemyEthKey))
	if err != nil {
		return nil, err
	}

	kosAbiContentBytes, err := os.ReadFile("abi/KeyOfSalvation.json")
	if err != nil {
		return nil, err
	}

	// converts the array of bytes obtained from reading the abi to a string
	kosAbiContent := string(kosAbiContentBytes)

	// KOS address on ETH mainnet
	kosAddr := common.HexToAddress("0x34BFF2Dbf20cF39dB042cb68D42D6d06fdbd85D3")
	abi, err := abi.JSON(strings.NewReader(kosAbiContent))
	if err != nil {
		return nil, err
	}

	// checks if any of the parameters are empty
	if caller == nil {
		caller = client
	}
	if transactor == nil {
		transactor = client
	}
	if filterer == nil {
		filterer = client
	}

	contract := bind.NewBoundContract(kosAddr, abi, client, client, client)

	return contract, nil
}
