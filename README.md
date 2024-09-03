# celestia-node-monitor

A simple monitor for the Celestia DA node, it support full, bridge, and light node.
# Features
* It will send alert in discord channel when your celestia DA node is down, not synced, or lack of balance. 
* It will check the status every 5 minutes.


# Pre-requisites

1. You have go 1.22+ installed.
2. You have set a discord webhook. You can follow this guide: https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks

# How to use

## Get your auth token
```
# mocha bridge node for example 
AUTH_TOKEN=$(celestia bridge auth admin --p2p.network mocha)
echo $AUTH_TOKEN
```

## Input your webhook url in .env file
`cp .env.example .env`
then edit .env file

## Edit config.toml
`cp config.toml.example config.toml`
then edit config.toml file
Configure the following parameters in the `config.toml` file:

* `[node]` section:
  - `standardConsensusRPC`: Celestia Network Public Consensus RPC address
  - `minimumBalance`: Minimum balance of your node (in utia). You'll receive an alert if the balance falls below this value.

* `[[node.APIs]]` section:
  - `URL`: Celestia DA Node API address, default is http://localhost:26658
  - `Token`: Your da node authentication token

* `[discord]` section:
  - `alertuserid`: Discord user ID. Set this if you want to be @ mentioned when receiving alerts.
  - `alertroleid`: Discord role ID. Set this if you want members with a specific role to be @ mentioned when receiving alerts.

You can add multiple `[[node.APIs]]` blocks to monitor multiple nodes.

For instructions on how to find Discord user ID and role ID:
- User ID: https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-
- Role ID: https://www.itgeared.com/how-to-get-role-id-on-discord/

## Build 
```shell
 go build -o celestia-node-monitor  ./cmd
```





## Run
```shell
./celestia-node-monitor
```

You can check the log in `celestia-node-monitor.log`


## Run as a service
```
sudo tee /etc/systemd/system/clelestia-node-monitor.service > /dev/null << EOF
[Unit]
Description=celestia-node-monitor
After=network-online.target

[Service]
User=$USER
WorkingDirectory=$HOME/celestia-node-monitor
ExecStart=$HOME/celestia-node-monitor/celestia-node-monitor
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target

EOF
```
start

```
sudo systemctl enable celestia-node-monitor
sudo systemctl daemon-reload
sudo systemctl start celestia-node-monitor
```
check logs
```
sudo journalctl -fu  celestia-node-monitor  -o cat
```

