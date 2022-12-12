# ChatGPT golang 接口对接

2022-12-12 开始，chatgpt 使用了 Cloudflare 来检查机器人。作为应对措施，我们需要在 cookie 中额外复制 cf_clearance
作为参数传入构建方法中。

## ⚠️⚠️⚠️注意

- 保持登录与运行环境的 ip、user-agent 一致: 更换 ip 或使用其他的 user-agent 会导致接口 403。如果
  user-agent、session_token、clearance_token 都确认与浏览器一致，那么需要确认 ip 是否与浏览器的一致。
- cf_clearance 有效期2小时，意味着需要经常更换它
- 运行机器人后，关闭浏览器的 chatgpt: 浏览器会更新你的 sessionToken，更新后旧的就无法使用
- 据观察，消息短则数秒得到响应，长时会到分钟级，设置超时时间为 3分钟 比较稳妥

## 特性

- 支持会话
- 支持消息上下文
- 支持刷新 accessToken
- 支持 2022-12-12 更新后机器人校验 token 的附加

## 运行测试

1. 已经登录好的 ChatGPT 账号
   从登录好的网站 https://chat.openai.com/chat 控制台获取 __Secure-next-auth.session-token session 的值
   ![img](./assets/img.png)
2. 设置程序测试的环境变量
   SESSION_KEY={上边获取到的 cookie 值}
3. 运行测试代码 go test -run TestChatGPT_SendMessage
   ![img](./assets/img_1.png)

## 使用

```go
import "github.com/zhan3333/chatgpt-go"
```

## 代码示例

```go
package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	chatgpt_go "github.com/zhan3333/chatgpt-go"
	"os"
	"time"
)

func main() {
	var sessionToken = os.Getenv("SESSION_KEY")
	var clearanceToken = os.Getenv("CLEARANCE_TOKEN")
	var userAgent = os.Getenv("USER_AGENT")
	timeout := time.Second * 60
	client, err := chatgpt_go.NewChatGPT(os.Getenv("SESSION_KEY"), chatgpt_go.ChatGPTOptions{
		SessionToken:   sessionToken,
		ClearanceToken: clearanceToken,
		UserAgent:      userAgent,
		Log:            logrus.NewEntry(logrus.StandardLogger()),
		Timeout:        &timeout,
	})
	if err != nil {
		panic(err)
	}
	conversation := client.NewConversation("", "")
	resp, err := conversation.SendMessage("hello")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
```
