package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pingan/hydra/internal/model"
)

// DingTalkChannel sends alert notifications via DingTalk webhook.
type DingTalkChannel struct {
	webhookURL string
	secret     string // optional HMAC signing secret
	client     *http.Client
}

// NewDingTalkChannel creates a DingTalk notification channel.
func NewDingTalkChannel(webhookURL, secret string) *DingTalkChannel {
	return &DingTalkChannel{
		webhookURL: webhookURL,
		secret:     secret,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *DingTalkChannel) Name() string { return "dingtalk" }

// dingTalkMessage is the Markdown message payload for DingTalk bot.
type dingTalkMessage struct {
	MsgType  string            `json:"msgtype"`
	Markdown dingTalkMarkdown  `json:"markdown"`
	At       *dingTalkAt       `json:"at,omitempty"`
}

type dingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type dingTalkAt struct {
	IsAtAll bool `json:"isAtAll"`
}

// Send delivers an alert as a DingTalk Markdown card.
func (d *DingTalkChannel) Send(alert *model.Alert) error {
	text := d.buildMarkdown(alert)

	msg := dingTalkMessage{
		MsgType: "markdown",
		Markdown: dingTalkMarkdown{
			Title: fmt.Sprintf("[%s] %s", alert.Severity, alert.RuleName),
			Text:  text,
		},
	}

	// At all members for critical alerts
	if alert.Severity == "critical" {
		msg.At = &dingTalkAt{IsAtAll: true}
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal dingtalk message: %w", err)
	}

	url := d.webhookURL
	if d.secret != "" {
		url = d.signURL(url)
	}

	resp, err := d.client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("dingtalk post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk returned %d", resp.StatusCode)
	}

	log.Printf("dingtalk: sent alert %s (severity=%s)", alert.ID, alert.Severity)
	return nil
}

func (d *DingTalkChannel) buildMarkdown(alert *model.Alert) string {
	severityEmoji := map[string]string{
		"critical": "!",
		"warning":  "!",
		"info":     "i",
	}
	emoji := severityEmoji[alert.Severity]
	if emoji == "" {
		emoji = "i"
	}

	return fmt.Sprintf(`### [%s] %s

> %s

- **Metric**: %s
- **Value**: %.2f (threshold: %.2f)
- **Host**: %s
- **Region**: %s
- **Time**: %s`,
		alert.Severity, alert.RuleName,
		alert.Message,
		alert.Metric, alert.Value, alert.Threshold,
		alert.Host, alert.Region,
		alert.FiredAt.Format("2006-01-02 15:04:05"),
	)
}

// signURL appends HMAC-SHA256 signature parameters to the webhook URL.
func (d *DingTalkChannel) signURL(url string) string {
	ts := time.Now().UnixMilli()
	signStr := fmt.Sprintf("%d\n%s", ts, d.secret)
	mac := hmac.New(sha256.New, []byte(d.secret))
	mac.Write([]byte(signStr))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s&timestamp=%d&sign=%s", url, ts, sign)
}
