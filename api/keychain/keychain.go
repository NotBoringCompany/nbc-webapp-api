package keychain

// func loadKeychain(caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
// 	alchemyEthKey := os.Getenv("ALCHEMY_ETH_API_KEY")

// 	// connect to ETH Mainnet
// 	client, err := ethclient.Dial(fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", alchemyEthKey))
// 	if err != nil {
// 		return nil, err
// 	}

// 	kycAbiContentBytes, err := os.ReadFile("abi/Keychain.json")
// 	if err != nil {
// 		return nil, err
// 	}

// 	// converts the array of bytes obtained from reading the abi to a string
// 	kycAbiContent := string(kycAbiContentBytes)

// 	// Keychain address on ETH mainnet
// 	kycAddr := common.HexToAddress("0xBbD427AbbA5fA84d29fB1e9F09F12D0B7D2E017f")
// 	abi, err := abi.JSON(strings.NewReader(kycAbiContent))
// 	if err != nil {
// 		return nil, err
// 	}

// 	// checks if any of the parameters are empty
// 	if caller == nil {
// 		caller = client
// 	}
// 	if transactor == nil {
// 		transactor = client
// 	}
// 	if filterer == nil {
// 		filterer = client
// 	}
// }
