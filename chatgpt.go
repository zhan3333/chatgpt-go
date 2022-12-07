package chatgpt_go

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"

type ChatGPT struct {
	SessionToken       string
	AccessToken        string
	AccessTokenExpires time.Time
	Log                *logrus.Entry
	Timeout            time.Duration
}

type ChatGPTOptions struct {
	Log     *logrus.Entry
	Timeout *time.Duration
}

func NewChatGPT(sessionToken string, options ChatGPTOptions) (*ChatGPT, error) {
	if sessionToken == "" {
		return nil, fmt.Errorf("sessionToken must set")
	}
	c := &ChatGPT{
		SessionToken: sessionToken,
		Log:          options.Log,
	}
	if options.Timeout != nil {
		c.Timeout = *options.Timeout
	} else {
		c.Timeout = time.Second * 10
	}
	return c, nil
}

type SessionResult struct {
	User struct {
		Id       string        `json:"id"`
		Name     string        `json:"name"`
		Email    string        `json:"email"`
		Image    string        `json:"image"`
		Picture  string        `json:"picture"`
		Groups   []interface{} `json:"groups"`
		Features []interface{} `json:"features"`
	} `json:"user"`
	Expires     time.Time `json:"expires"`
	AccessToken string    `json:"accessToken"`
}

func (c *ChatGPT) IsAccessTokenExpired() bool {
	return time.Now().After(c.AccessTokenExpires)
}

func (c *ChatGPT) RefreshAccessToken() error {
	if c.AccessToken == "" || c.IsAccessTokenExpired() {
		req, err := http.NewRequest(http.MethodGet, "https://chat.openai.com/api/auth/session", nil)
		if err != nil {
			return err
		}
		req.Header.Set("cookie", fmt.Sprintf("__Secure-next-auth.session-token=%s", c.SessionToken))
		req.Header.Set("user-agent", UserAgent)
		resp, err := (&http.Client{Timeout: c.Timeout}).Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("response status=%d not 200", resp.StatusCode)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		respJson := SessionResult{}
		if err := json.Unmarshal(b, &respJson); err != nil {
			return err
		}
		if respJson.AccessToken == "" {
			return fmt.Errorf("invalid resp: %s", string(b))
		}
		c.AccessToken = respJson.AccessToken
		c.AccessTokenExpires = respJson.Expires
	}
	return nil
}

type Conversation struct {
	ChatGPT         *ChatGPT
	ConversationId  string
	ParentMessageId string
}

func (c *ChatGPT) NewConversation(conversationId string, parentMessageId string) *Conversation {
	return &Conversation{
		ChatGPT:         c,
		ConversationId:  conversationId,
		ParentMessageId: parentMessageId,
	}
}

type ConversationBodyMessage struct {
	Id      string `json:"id"`
	Role    string `json:"role"`
	Content struct {
		ContentType string   `json:"content_type"`
		Parts       []string `json:"parts"`
	} `json:"content"`
}

type ConversationBody struct {
	Action          string                    `json:"action"`
	Messages        []ConversationBodyMessage `json:"messages"`
	ParentMessageId string                    `json:"parent_message_id"`
	Model           string                    `json:"model"`
	ConversationId  string                    `json:"conversation_id,omitempty"`
}

type ConversationResult struct {
	Message struct {
		Id         string      `json:"id"`
		Role       string      `json:"role"`
		User       interface{} `json:"user"`
		CreateTime interface{} `json:"create_time"`
		UpdateTime interface{} `json:"update_time"`
		Content    struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
		EndTurn  interface{} `json:"end_turn"`
		Weight   float64     `json:"weight"`
		Metadata struct {
		} `json:"metadata"`
		Recipient string `json:"recipient"`
	} `json:"message"`
	ConversationId string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}

func (r *ConversationResult) GetMessage() (string, error) {
	return r.Message.Content.Parts[0], nil
}

func (r *ConversationResult) JSON() []byte {
	bs, _ := json.Marshal(r)
	return bs
}

func (b *ConversationBody) Reader() (io.Reader, error) {
	bs, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bs), nil
}

func (b *ConversationBody) JSON() []byte {
	bs, _ := json.Marshal(b)
	return bs
}

func (c *Conversation) SendMessage(message string) (string, error) {
	if c.ParentMessageId == "" {
		c.ParentMessageId = uuid.NewString()
	}
	if err := c.ChatGPT.RefreshAccessToken(); err != nil {
		return "", err
	}
	body := ConversationBody{
		Action: "next",
		Messages: []ConversationBodyMessage{{
			Id:   uuid.NewString(),
			Role: "user",
			Content: struct {
				ContentType string   `json:"content_type"`
				Parts       []string `json:"parts"`
			}{
				ContentType: "text",
				Parts:       []string{message},
			},
		}},
		ParentMessageId: c.ParentMessageId,
		Model:           "text-davinci-002-render",
	}
	if c.ConversationId != "" {
		body.ConversationId = c.ConversationId
	}
	bodyReader, err := body.Reader()
	if c.ChatGPT.Log != nil {
		c.ChatGPT.Log.WithField("body", string(body.JSON())).Debug("send_request")
	}
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, "https://chat.openai.com/backend-api/conversation", bodyReader)
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", c.ChatGPT.AccessToken)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", UserAgent)
	resp, err := (&http.Client{Timeout: c.ChatGPT.Timeout}).Do(req)
	if err != nil {
		return "", err
	}

	respMessage := ""
	br := bufio.NewReader(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	delim := []byte{':', ' '}

	for {
		bs, err := br.ReadBytes('\n')

		if err != nil && err != io.EOF {
			return "", err
		}

		if len(bs) < 2 {
			continue
		}

		spl := bytes.SplitN(bs, delim, 2)

		if len(spl) < 2 {
			continue
		}

		value := strings.TrimSuffix(string(spl[1]), "\n")

		if err == io.EOF || value == "[DONE]" {
			break
		}
		respMessage = value
	}

	result := ConversationResult{}
	if err := json.Unmarshal([]byte(respMessage), &result); err != nil {
		return "", err
	}

	if c.ChatGPT.Log != nil {
		c.ChatGPT.Log.WithField("body", string(result.JSON())).Debug("send_response")
	}

	c.ParentMessageId = result.Message.Id
	c.ConversationId = result.ConversationId

	return result.GetMessage()
}
