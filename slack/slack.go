package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"encore.dev/beta/errs"
	"io"
	"net/http"
)

type NotifyParams struct {
	Text string `json:"text"`
}

//encore:api private
func Notify(ctx context.Context, p *NotifyParams) error {
	eb := errs.B()
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", secrets.SlackWebhookURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return eb.Code(errs.Unavailable).Msgf("notify slack: %s: %s", resp.Status, body).Err()
	}
	return nil
}

var secrets struct {
	SlackWebhookURL string
}
