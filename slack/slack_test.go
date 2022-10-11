package slack

import (
	"context"
	"gopkg.in/h2non/gock.v1"
	"testing"
)

func createMock() *gock.Request {
	return gock.New("https://hooks.slack.com").Post("/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX")
}

func callNotify() error {
	return NotifyRaw(context.Background(), "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX", &NotifyParams{
		Text: "Hello world!",
	})
}

func TestSuccessfulNotify(t *testing.T) {
	createMock().Reply(200).BodyString("ok")
	defer gock.Off()
	if err := callNotify(); err != nil {
		t.Fatal("should have succeeded", err)
	}
}

func TestFailedNotify_InvalidToken(t *testing.T) {
	createMock().Reply(403).BodyString("invalid_token")
	defer gock.Off()
	if err := callNotify(); err == nil {
		t.Fatal("should have failed", err)
	}
}

func TestFailedNotify_InvalidPayload(t *testing.T) {
	createMock().Reply(400).BodyString("invalid_payload")
	defer gock.Off()
	if err := callNotify(); err == nil {
		t.Fatal("should have failed", err)
	}
}
