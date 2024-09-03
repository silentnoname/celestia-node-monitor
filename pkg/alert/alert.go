package alert

import (
	"bytes"
	"celestia-node-monitor/pkg/config"
	"celestia-node-monitor/pkg/log"
	"celestia-node-monitor/pkg/check"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	"net/http"
	"time"
)

type webhookPayload struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds"`
}

// SendMsgViaWebhook  send msg via discord websocket
func SendMsgViaWebhook(webhookUrl string, webhookPayload webhookPayload) error {
	b, err := json.Marshal(webhookPayload)
	if err != nil {
		return fmt.Errorf("failed to alert due to fail to marshal webhook payload: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to alert due to fail to create http request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to alert due to fail to send http request: %v", err)
	}
	if res.StatusCode != 204 {
		return fmt.Errorf("failed to alert,: %v", res)
	}
	return nil
}

// getEmbedAlert get Embed alert
func getEmbedAlert(Performances []monitor.Performance) []Embed {
	var embeds []Embed
	for _, Performance := range Performances {
		if !Performance.Sync.Synced {
			if Performance.Sync.Error == "" {
				embed := Embed{
					Title:       "Celestia DA Node Sync Alert",
					Description: "*Your node sync check show some problems, please check it*",
					Color:       16711680,
					Fields: []*EmbedField{
						{
							Name:   "URL",
							Value:  Performance.NodeURL,
							Inline: false,
						},
						{
							Name:   "Error",
							Value:  "Your node not synced",
							Inline: false,
						},
						{
							Name:   "Time",
							Value:  "<t:" + fmt.Sprint(time.Now().Unix()) + ">",
							Inline: false,
						},
					},
					Thumbnail: &EmbedThumbnail{
						URL: "https://i.imgur.com/5NmtLLy.jpg",
					},
				}
				embeds = append(embeds, embed)
			}
			if Performance.Sync.Error != "" {
				embed := Embed{
					Title:       "Celestia DA Node Sync Alert",
					Description: "*Your node sync check show some problems, please check it*",
					Color:       16711680,
					Fields: []*EmbedField{
						{
							Name:   "URL",
							Value:  Performance.NodeURL,
							Inline: false,
						},
						{
							Name:   "Error",
							Value:  Performance.Sync.Error,
							Inline: false,
						},
						{
							Name:   "Time",
							Value:  "<t:" + fmt.Sprint(time.Now().Unix()) + ">",
							Inline: false,
						},
					},
					Thumbnail: &EmbedThumbnail{
						URL: "https://i.imgur.com/5NmtLLy.jpg",
					},
				}
				embeds = append(embeds, embed)
			}
		}
		if !Performance.MinBalance.Enough {
			if Performance.MinBalance.Error == "" {
				embed := Embed{
					Title:       "Celestia DA Node Min balance check Alert",
					Description: "*Your node minimum balance check found some problem, please check it*",
					Color:       16711680,
					Fields: []*EmbedField{
						{
							Name:   "URL",
							Value:  Performance.NodeURL,
							Inline: false,
						},
						{
							Name:   "Error",
							Value:  "Your node balance is lower than minimum balance you set",
							Inline: false,
						},
						{
							Name:   "Time",
							Value:  "<t:" + fmt.Sprint(time.Now().Unix()) + ">",
							Inline: false,
						},
					},
					Thumbnail: &EmbedThumbnail{
						URL: "https://i.imgur.com/5NmtLLy.jpg",
					},
				}
				embeds = append(embeds, embed)
			}
			if Performance.MinBalance.Error != "" {
				embed := Embed{
					Title:       "Celestia DA Node Min balance check Alert",
					Description: "*Your node minimum balance check found some problem, please check it*",
					Color:       16711680,
					Fields: []*EmbedField{
						{
							Name:   "URL",
							Value:  Performance.NodeURL,
							Inline: false,
						},
						{
							Name:   "Error",
							Value:  Performance.MinBalance.Error,
							Inline: false,
						},
						{
							Name:   "Time",
							Value:  "<t:" + fmt.Sprint(time.Now().Unix()) + ">",
							Inline: false,
						},
					},
					Thumbnail: &EmbedThumbnail{
						URL: "https://i.imgur.com/5NmtLLy.jpg",
					},
				}
				embeds = append(embeds, embed)
			}
		}
	}
	return embeds
}

// SendAlertViaDiscord send alert
func SendAlertViaDiscord(config config.Config, badperformances []monitor.Performance) error {

	embeds := getEmbedAlert(badperformances)
	if len(embeds) == 0 {
		log.Log.Info("There is no alert to send")
		return nil
	}
	log.Log.Info("Sending alert via Discord")
	var mention string
	if len(config.Discord.AlertUserID) != 0 && config.Discord.AlertUserID[0] != "" {
		for _, userid := range config.Discord.AlertUserID {
			mention += "<@" + userid + "> "
		}
	}
	if len(config.Discord.AlertRoleID) != 0 && config.Discord.AlertRoleID[0] != "" {
		for _, roleid := range config.Discord.AlertRoleID {
			mention += "<@&" + roleid + "> "
		}
	}
	payload := webhookPayload{
		Content: "" + mention,
		Embeds:  embeds,
	}
	var err error
	retry.Do(
		func() error {
			err = SendMsgViaWebhook(config.Discord.WebhookUrl, payload)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Delay(time.Second*3),
		retry.Attempts(3),
	)
	if err != nil {
		return fmt.Errorf("failed to send alert via discord: %v", err)
	}
	log.Log.Info("send alert via discord successfully")
	return nil
}
