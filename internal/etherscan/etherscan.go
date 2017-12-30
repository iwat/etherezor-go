package etherscan

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
)

func BalanceOf(addr string) (*big.Int, error) {
	ret := (*big.Int)(nil)

	err := callGet("account", "balance", addr, func(body io.Reader) error {
		resp := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Result  string `json:"result"`
		}{}

		e := json.NewDecoder(body)
		err := e.Decode(&resp)
		if err != nil {
			return err
		}

		if resp.Status != "1" {
			return fmt.Errorf("etherscan error %s %s", resp.Status, resp.Message)
		}

		ret = big.NewInt(0)
		if _, ok := ret.SetString(resp.Result, 10); !ok {
			return fmt.Errorf("parse int failed %s", resp.Result)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func callGet(module, action, address string, cb func(body io.Reader) error) error {
	query := url.Values{}
	query.Set("module", module)
	query.Set("action", action)
	query.Set("address", address)
	query.Set("tag", "latest")
	query.Set("apikey", os.Getenv("ETHERSCAN_API_KEY"))

	resp, err := http.Get("https://api.etherscan.io/api?" + query.Encode())
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return cb(resp.Body)
}
