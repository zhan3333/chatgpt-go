# ChatGPT golang 接口对接

## 特性

- 支持会话
- 支持消息上下文
- 支持刷新 accessToken

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
import "github.com/gin-gonic/gin"
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
	timeout := time.Second * 60
	client, err := chatgpt_go.NewChatGPT(os.Getenv("SESSION_KEY"), chatgpt_go.ChatGPTOptions{
		Log:     logrus.NewEntry(logrus.StandardLogger()),
		Timeout: &timeout,
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

## 注意

- 据观察，accessToken 有效期在1年左右，sessionToken 有效期估计在一天左右，下一步会更新 accessToken 创建对象的方法
- 据观察，消息短则数秒得到响应，长时会到分钟级，设置超时时间为 1分钟 比较稳妥