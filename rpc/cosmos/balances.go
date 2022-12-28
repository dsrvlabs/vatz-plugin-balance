package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Bank struct {
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
	Pagination struct {
		NextKey interface{} `json:"next_key"`
		Total   string      `json:"total"`
	} `json:"pagination"`
}

type Auth struct {
	Account struct {
		Type        string `json:"@type"`
		BaseAccount struct {
			Address string `json:"address"`
			PubKey  struct {
				Type string `json:"@type"`
				Key  string `json:"key"`
			} `json:"pub_key"`
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
		} `json:"base_account"`
		CodeHash string `json:"code_hash"`
	} `json:"account"`
}

func GetBalances(apiPort uint, account string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/cosmos/bank/v1beta1/balances/%s", apiPort, account)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err.Error(), err
	}

	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err.Error(), err
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error(), err
	}

	if resp.StatusCode != http.StatusOK {
		return strconv.Itoa(resp.StatusCode), fmt.Errorf("request failed %s", string(rawBody))
	}

	bank := Bank{}
	err = json.Unmarshal(rawBody, &bank)

	if err != nil {
		return err.Error(), err
	}

	return bank.Balances[0].Amount, nil
}

func GetAccountInfo(apiPort uint, account string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/cosmos/auth/v1beta1/accounts/%s", apiPort, account)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err.Error(), err
	}

	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err.Error(), err
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error(), err
	}

	if resp.StatusCode != http.StatusOK {
		return strconv.Itoa(resp.StatusCode), fmt.Errorf("request failed %s", string(rawBody))
	}

	auth := Auth{}
	err = json.Unmarshal(rawBody, &auth)

	if err != nil {
		return err.Error(), err
	}

	return auth.Account.Type, nil
}
