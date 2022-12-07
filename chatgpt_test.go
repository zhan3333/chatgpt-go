package chatgpt_go_test

import (
	chatgpt_go "chatgpt-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var sessionToken = os.Getenv("SESSION_KEY")

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestChatGPT_SendMessage(t *testing.T) {
	timeout := time.Second * 30
	client, err := chatgpt_go.NewChatGPT(sessionToken, chatgpt_go.ChatGPTOptions{
		Log:     logrus.NewEntry(logrus.StandardLogger()),
		Timeout: &timeout,
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	conversation := client.NewConversation("", "")
	resp, err := conversation.SendMessage("hello")
	if assert.NoError(t, err) {
		t.Logf("resp: %s", resp)
	} else {
		t.FailNow()
	}

	cid := conversation.ConversationId

	resp, err = conversation.SendMessage("what's your name")
	if assert.NoError(t, err) {
		t.Logf("resp: %s", resp)
		assert.Equal(t, cid, conversation.ConversationId)
	} else {
		t.FailNow()
	}
}

func TestChatGPT_RefreshAccessToken(t *testing.T) {
	client, err := chatgpt_go.NewChatGPT(sessionToken, chatgpt_go.ChatGPTOptions{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = client.RefreshAccessToken()
	if assert.NoError(t, err) {
		t.Logf("accessToken: %s", client.AccessToken)
	}
}
