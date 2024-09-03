package monitor

import (
	"celestia-node-monitor/pkg/config"
	"celestia-node-monitor/pkg/log"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	client "github.com/celestiaorg/celestia-openrpc"
)

type Performance struct {
	NodeURL string
	Token   string
	Sync    struct {
		Synced bool
		Error  string
	}
	MinBalance struct {
		Enough bool
		Error  string
	}
}

type CelestiaRpcStatusResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		NodeInfo struct {
			ProtocolVersion struct {
				P2P   string `json:"p2p"`
				Block string `json:"block"`
				App   string `json:"app"`
			} `json:"protocol_version"`
			ID         string `json:"id"`
			ListenAddr string `json:"listen_addr"`
			Network    string `json:"network"`
			Version    string `json:"version"`
			Channels   string `json:"channels"`
			Moniker    string `json:"moniker"`
			Other      struct {
				TxIndex    string `json:"tx_index"`
				RPCAddress string `json:"rpc_address"`
			} `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash     string    `json:"latest_block_hash"`
			LatestAppHash       string    `json:"latest_app_hash"`
			LatestBlockHeight   string    `json:"latest_block_height"`
			LatestBlockTime     time.Time `json:"latest_block_time"`
			EarliestBlockHash   string    `json:"earliest_block_hash"`
			EarliestAppHash     string    `json:"earliest_app_hash"`
			EarliestBlockHeight string    `json:"earliest_block_height"`
			EarliestBlockTime   time.Time `json:"earliest_block_time"`
			CatchingUp          bool      `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address string `json:"address"`
			PubKey  struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
}

type BalanceResponse struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func exponentialBackoff(retry int) {
	time.Sleep(time.Duration((1<<retry)*1000) * time.Millisecond)
}

func getLatestBlockFromRPC(standardRPC string) (int, error) {
	timeout := 5
	req, err := http.NewRequest("GET", standardRPC+"/status", nil)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	var response CelestiaRpcStatusResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	standardrpcLatestBlock, _ := strconv.Atoi(response.Result.SyncInfo.LatestBlockHeight)
	return standardrpcLatestBlock, nil
}

func retryGetLatestBlockFromRPC(standardRPC string, maxRetries int) (int, error) {
	var lastErr error
	var latestBlock int

	for retry := 0; retry < maxRetries; retry++ {
		latestBlock, lastErr = getLatestBlockFromRPC(standardRPC)

		if lastErr == nil {
			return latestBlock, nil
		}
		if retry < maxRetries-1 {
			exponentialBackoff(retry)
		}
	}

	return 0, lastErr
}

func getLatestBlockFromAPI(url string, token string) (int, error) {
	ctx := context.Background()
	client, err := client.NewClient(ctx, url, token)
	if err != nil {
		return 0, err
	}
	block, err := client.Header.LocalHead(ctx)
	if err != nil {
		return 0, err
	}
	return int(block.Height()), nil
}

func retryGetLatestBlockFromAPI(url string, token string, maxRetries int) (int, error) {
	var lastErr error
	var latestBlock int

	for retry := 0; retry < maxRetries; retry++ {
		latestBlock, lastErr = getLatestBlockFromAPI(url, token)

		if lastErr == nil {
			return latestBlock, nil
		}

		if retry < maxRetries-1 {
			exponentialBackoff(retry)
		}
	}

	return 0, lastErr
}

func getBalanceFromAPI(url string, token string) (int, error) {
	ctx := context.Background()
	client, err := client.NewClient(ctx, url, token)
	if err != nil {
		return 0, err
	}
	balance, err := client.State.Balance(ctx)
	if err != nil {
		return 0, err
	}
	amount := int(balance.Amount.Int64())
	return amount, nil
}

func retryGetBalanceFromAPI(url string, token string, maxRetries int) (int, error) {
	var lastErr error
	var balance int

	for retry := 0; retry < maxRetries; retry++ {
		balance, lastErr = getBalanceFromAPI(url, token)

		if lastErr == nil {
			return balance, nil
		}

		if retry < maxRetries-1 {
			exponentialBackoff(retry)
		}
	}

	return 0, lastErr
}

// checkSyncStatus checks if the gateway API is synced with the standard RPC, false means not synced, true means synced
func checkSyncStatus(API string, token string, standardRPC string) (bool, error) {
	apiBlock, err := retryGetLatestBlockFromAPI(API, token, 3)
	if err != nil {
		return false, fmt.Errorf("failed to get latest block from node API: %v", err)
	}

	rpcBlock, err := retryGetLatestBlockFromRPC(standardRPC, 3)
	if err != nil {
		return false, fmt.Errorf("failed to get latest block from RPC: %v", err)
	}
	// We consider the max behind block is 5
	if (rpcBlock - apiBlock) > 5 {
		log.Log.Error("Node " + API + " is not synced, gateway height is " + strconv.Itoa(apiBlock) + " standard RPC block is " + strconv.Itoa(rpcBlock))
		return false, nil
	}
	return true, nil
}

// minBalanceCheck checks if the node has enough balance
func minBalanceCheck(API string, token string, minBalance int) (bool, error) {
	balance, err := retryGetBalanceFromAPI(API, token, 3)
	if err != nil {
		return false, fmt.Errorf("failed to get balance from node API: %v", err)
	}

	if balance < minBalance {
		log.Log.Error(fmt.Sprintf("Node %s does not have enough balance, balance is %d, minimum balance is %d", API, balance, minBalance))
		return false, nil
	}
	return true, nil
}

func CheckNodes(config config.Config) ([]Performance) {
	var NodesPerformance []Performance
	for _, apiConfig := range config.Node.APIs {
		log.Log.Info("Checking node: " + apiConfig.URL)
		Node := Performance{
			NodeURL: apiConfig.URL,
			Token:   apiConfig.Token,
		}

		// check sync status
		isSynced, err := checkSyncStatus(apiConfig.URL, apiConfig.Token, config.Node.StandardConsensusRPC)
		if err != nil {
			log.Log.Error("Failed to check node " + apiConfig.URL + "sync status: " + err.Error())
			Node.Sync.Synced = false
			Node.Sync.Error = err.Error()
		} else {
			Node.Sync.Synced = isSynced
		}

		// check balance
		hasEnoughBalance, err := minBalanceCheck(apiConfig.URL, apiConfig.Token, config.Node.MinimumBalance)
		if err != nil {
			log.Log.Error("Failed to check node balance" + apiConfig.URL + err.Error())
			Node.MinBalance.Enough = false
			Node.MinBalance.Error = err.Error()
		} else {
			Node.MinBalance.Enough = hasEnoughBalance
		}
		NodesPerformance = append(NodesPerformance, Node)

	}
	return NodesPerformance
}
