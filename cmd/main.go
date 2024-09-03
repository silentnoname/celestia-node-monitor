package main

import (
	"celestia-node-monitor/pkg/alert"
	"celestia-node-monitor/pkg/config"
	"celestia-node-monitor/pkg/log"
	"celestia-node-monitor/pkg/check"
	"strconv"
	"strings"
	"time"
	"go.uber.org/zap"
)

func main() {
	log.InitLog()
	log.Log.Info("Start celestia-node-monitor")
	log.Log.Info("Loading config")
	cfg, err := config.LoadConfig("config.toml", ".env")
	if err != nil {
		panic(err)
	}
	log.Log.Info("Standard consensus rpc is " + cfg.Node.StandardConsensusRPC)
	nodeURLs := make([]string, len(cfg.Node.APIs))
	for i, api := range cfg.Node.APIs {
		nodeURLs[i] = api.URL
	}
	log.Log.Info("Nodes to check: " + strings.Join(nodeURLs, ", "))
	log.Log.Info("Will alert when node balance is less than " + strconv.Itoa(cfg.Node.MinimumBalance) + " utia")
	log.Log.Info("Will check node performance every 5 minutes")
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		log.Log.Info("Start to check node performance")
		nodePerformances:= monitor.CheckNodes(*cfg)
		log.Log.Info("Start to check and send alert")
		alerterr := alert.SendAlertViaDiscord(*cfg, nodePerformances)
		if alerterr != nil {
			log.Log.Error("Failed to send alert", zap.Error(err))
		}
	}

}
