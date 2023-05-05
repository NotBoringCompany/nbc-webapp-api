package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

/*
Fetches the wallet from `sessionToken` and checks if it matches `walletToCheck`.
Returns true if the wallet matches, false otherwise.
*/
func CheckWalletMatchFromSessionToken(sessionToken, walletToCheck string) (bool, error) {
	res, err := http.Get(fmt.Sprintf(`https://nbc-webapp-api-ts-production.up.railway.app/backend-account/fetch-wallet-from-session-token/%s`, sessionToken))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	var responseBody struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Data    struct {
			WalletAddress string `json:"walletAddress"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
		return false, err
	}

	if responseBody.Status != 200 {
		return false, fmt.Errorf("unable to fetch wallet address from session token: %s", responseBody.Message)
	}

	fmt.Println("wallet address from session token:", responseBody.Data.WalletAddress)
	fmt.Println("wallet to check:", walletToCheck)
	fmt.Println("wallet address from session token (lowercase):", strings.ToLower(responseBody.Data.WalletAddress))

	walletAddress := responseBody.Data.WalletAddress
	// check if the wallet address matches the staker's wallet address.
	if walletAddress != strings.ToLower(walletToCheck) {
		return false, nil
	}

	return true, nil
}
