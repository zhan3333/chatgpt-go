package chatgpt_go_test

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	chatgpt_go "github.com/zhan3333/chatgpt-go"
	"os"
	"testing"
	"time"
)

var sessionToken = os.Getenv("SESSION_KEY")
var clearanceToken = os.Getenv("CLEARANCE_TOKEN")
var userAgent = os.Getenv("USER_AGENT")

func TestMain(m *testing.M) {
	if sessionToken == "" {
		panic("env SESSION_KEY not set")
	}
	if clearanceToken == "" {
		panic("env CLEARANCE_TOKEN not set")
	}
	if userAgent == "" {
		panic("env USER_AGENT not set")
	}
	logrus.SetLevel(logrus.DebugLevel)
	m.Run()
}

func TestChatGPT_SendMessage(t *testing.T) {
	t.Logf("sessionToken: %s", sessionToken)
	t.Logf("clearanceToken: %s", clearanceToken)
	t.Logf("userAgent: %s", userAgent)

	timeout := time.Second * 60
	client, err := chatgpt_go.NewChatGPT(chatgpt_go.ChatGPTOptions{
		SessionToken:   sessionToken,
		ClearanceToken: clearanceToken,
		UserAgent:      userAgent,
		Log:            logrus.NewEntry(logrus.StandardLogger()),
		Timeout:        &timeout,
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	conversation := client.NewConversation("", "")
	resp, err := conversation.SendMessage("你好")
	if assert.NoError(t, err) {
		t.Logf("resp: %s", resp)
	} else {
		t.FailNow()
	}

	cid := conversation.ConversationId

	resp, err = conversation.SendMessage("你叫什么名字")
	if assert.NoError(t, err) {
		t.Logf("resp: %s", resp)
		assert.Equal(t, cid, conversation.ConversationId)
	} else {
		t.FailNow()
	}
}

func TestChatGPT_RefreshAccessToken(t *testing.T) {
	t.Logf("sessionToken: %s", sessionToken)
	t.Logf("clearanceToken: %s", clearanceToken)
	t.Logf("userAgent: %s", userAgent)
	client, err := chatgpt_go.NewChatGPT(chatgpt_go.ChatGPTOptions{
		SessionToken:   sessionToken,
		ClearanceToken: clearanceToken,
		UserAgent:      userAgent,
		Log:            logrus.NewEntry(logrus.StandardLogger()),
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = client.RefreshAccessToken()
	if assert.NoError(t, err) {
		t.Logf("accessToken: %s", client.AccessToken)
	}
}
