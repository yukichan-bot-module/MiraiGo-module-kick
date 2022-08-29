package kick

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

var instance *kick
var logger = utils.GetModuleLogger("com.aimerneige.kick")

type kick struct {
}

func init() {
	instance = &kick{}
	bot.RegisterModule(instance)
}

func (k *kick) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "com.aimerneige.kick",
		Instance: instance,
	}
}

// Init 初始化过程
// 在此处可以进行 Module 的初始化配置
// 如配置读取
func (k *kick) Init() {
}

// PostInit 第二次初始化
// 再次过程中可以进行跨 Module 的动作
// 如通用数据库等等
func (k *kick) PostInit() {
}

// Serve 注册服务函数部分
func (k *kick) Serve(b *bot.Bot) {
	b.GroupMessageEvent.Subscribe(func(c *client.QQClient, msg *message.GroupMessage) {
		// 格式一定不对的返回
		if len(msg.Elements) < 2 {
			return
		}
		groupCode := msg.GroupCode
		// 检查发送者管理员权限
		senderMemberInfo, err := c.GetMemberInfo(groupCode, msg.Sender.Uin)
		if err != nil {
			errMsg := fmt.Sprintf("在群「%d」获取成员「%d」的用户数据时发成错误，详情请查阅后台日志。", groupCode, msg.Sender.Uin)
			logger.WithError(err).Errorf(errMsg)
			c.SendGroupMessage(groupCode, simpleText(errMsg))
			return
		}
		// 发送者没有管理员权限，忽略消息
		if senderMemberInfo.Permission != client.Administrator && senderMemberInfo.Permission != client.Owner {
			return
		}
		// 检查 bot 管理员权限
		botPermission := true
		botMemberInfo, err := c.GetMemberInfo(groupCode, c.Uin)
		if err != nil {
			errMsg := fmt.Sprintf("在群「%d」获取成员「%d」的用户数据时发成错误，详情请查阅后台日志。", groupCode, msg.Sender.Uin)
			logger.WithError(err).Errorf(errMsg)
			c.SendGroupMessage(groupCode, simpleText(errMsg))
			return
		}
		if botMemberInfo.Permission != client.Administrator && botMemberInfo.Permission != client.Owner {
			botPermission = false
		}
		// 解析指令
		isAt := false
		isKick := false
		var target int64
		for _, ele := range msg.Elements {
			switch e := ele.(type) {
			case *message.AtElement:
				isAt = true
				target = e.Target
			case *message.TextElement:
				contentStr := e.Content
				contentStr = strings.TrimSpace(contentStr)
				if contentStr == "踢" || contentStr == "kick" || contentStr == "Kick" {
					isKick = true
				}
			}
		}
		if isAt && isKick && target != 0 {
			if botPermission == false {
				c.SendGroupMessage(groupCode, simpleText("请先授予机器人管理员权限。"))
				return
			}
			if target == c.Uin {
				c.SendGroupMessage(groupCode, simpleText("不要踢掉我啊~"))
				return
			}
			targetMemberInfo, err := c.GetMemberInfo(groupCode, target)
			if err != nil {
				errMsg := fmt.Sprintf("在群「%d」获取成员「%d」的用户数据时发成错误，详情请查阅后台日志。", groupCode, target)
				logger.WithError(err).Errorf(errMsg)
				c.SendGroupMessage(groupCode, simpleText(errMsg))
				return
			}
			if targetMemberInfo.Permission == client.Owner {
				c.SendGroupMessage(groupCode, simpleText("你居然想踢掉群主？真是危险的想法呢~"))
				return
			}
			if targetMemberInfo.Permission == client.Administrator {
				c.SendGroupMessage(groupCode, simpleText("管理员是踢不掉的呢~"))
				return
			}
			if err := targetMemberInfo.Kick("", false); err != nil {
				errMsg := fmt.Sprintf("在将成员「%d」移出群「%d」的过程中发生错误，详情请查阅后台日志。", target, groupCode)
				logger.WithError(err).Errorf(errMsg)
				c.SendGroupMessage(groupCode, simpleText(errMsg))
				return
			}
			c.SendGroupMessage(groupCode, simpleText(fmt.Sprintf("群成员「%d」已被管理员「%d」移出群。", target, msg.Sender.Uin)))
		}
	})
}

// Start 此函数会新开携程进行调用
// ```go
//
//	go exampleModule.Start()
//
// ```
// 可以利用此部分进行后台操作
// 如 http 服务器等等
func (k *kick) Start(b *bot.Bot) {
}

// Stop 结束部分
// 一般调用此函数时，程序接收到 os.Interrupt 信号
// 即将退出
// 在此处应该释放相应的资源或者对状态进行保存
func (k *kick) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
}

func simpleText(s string) *message.SendingMessage {
	return message.NewSendingMessage().Append(message.NewText(s))
}
