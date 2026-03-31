package model

import (
	"errors"
	"message-pusher/common"
	"time"
)

// type Message struct {
// 	Id          int    `json:"id"`
// 	UserId      int    `json:"user_id" gorm:"index"`
// 	Title       string `json:"title"`
// 	Description string `json:"description"`
// 	Content     string `json:"content"`
// 	URL         string `json:"url" gorm:"column:url"`
// 	Channel     string `json:"channel"`
// 	Token       string `json:"token" gorm:"-:all"`
// 	HTMLContent string `json:"html_content"  gorm:"-:all"`
// 	Timestamp   int64  `json:"timestamp" gorm:"type:bigint"`
// 	Link        string `json:"link" gorm:"unique;index"`
// 	To          string `json:"to" gorm:"column:to"`           // if specified, will send to this user(s)
// 	Status      int    `json:"status" gorm:"default:0;index"` // pending, sent, failed
// 	OpenId      string `json:"openid" gorm:"-:all"`           // alias for to
// 	Desp        string `json:"desp" gorm:"-:all"`             // alias for content
// 	Short       string `json:"short" gorm:"-:all"`            // alias for description
// 	Async       bool   `json:"async" gorm:"-"`                // if true, will send message asynchronously
// 	RenderMode  string `json:"render_mode" gorm:"raw"`        // markdown (default), code, raw
// }

type Message struct {
	// [Id] 中文名：消息唯一 ID
	// 设计目的：数据库自增主键，用于唯一标识系统中的每一条消息记录。
	Id          int    `json:"id"`

	// [UserId] 中文名：所属用户 ID
	// 设计目的：外键字段，关联 User 表的 Id。设置了索引 (index) 方便
	// 快速查询某个用户发送的所有历史消息。
	UserId      int    `json:"user_id" gorm:"index"`

	// [Title] 中文名：消息标题
	// 设计目的：推送消息的主标题，通常在通知栏的第一行显示。
	Title       string `json:"title"`

	// [Description] 中文名：消息描述/摘要
	// 设计目的：消息的简短说明。在某些渠道（如微信卡片）中作为正文预览展示。
	Description string `json:"description"`

	// [Content] 中文名：消息详情内容
	// 设计目的：存储消息的完整主体。支持长文本，是推送中最核心的信息部分。
	Content     string `json:"content"`

	// [URL] 中文名：点击跳转链接
	// 设计目的：用户点击推送通知后，浏览器或 APP 应该跳转到的网页地址。
	URL         string `json:"url" gorm:"column:url"`

	// [Channel] 中文名：发送渠道名称
	// 设计目的：记录该消息是通过哪个渠道发送的（如 "email", "ding", "wechat"）。
	Channel     string `json:"channel"`

	// [Token] 中文名：临时鉴权令牌
	// 设计目的：API 调用时传入的验证字符串。因为标记了 gorm:"-:all"，
	// 它不会存入数据库，仅在接收请求时用于身份校验。
	Token       string `json:"token" gorm:"-:all"`

	// [HTMLContent] 中文名：HTML 渲染内容
	// 设计目的：在网页预览消息时，将 Markdown 转换后的 HTML 源码。
	// 不存入数据库 (gorm:"-:all")，由程序在展示时动态生成。
	HTMLContent string `json:"html_content"  gorm:"-:all"`

	// [Timestamp] 中文名：发送时间戳
	// 设计目的：记录消息创建的精确时间（秒）。使用 bigint 类型确保兼容性，
	// 方便在前端展示“xx分钟前”或进行时间范围筛选。
	Timestamp   int64  `json:"timestamp" gorm:"type:bigint"`

	// [Link] 中文名：消息详情访问短链接
	// 设计目的：系统生成的唯一 UUID 字符串。用于拼接成类似于 "/message/uuid" 的链接，
	// 设置了唯一索引，用户无需登录即可通过此链接查看消息网页版。
	Link        string `json:"link" gorm:"unique;index"`

	// [To] 中文名：指定接收者
	// 设计目的：如果消息不是发给自己，而是发给特定的群组或他人（如邮件地址、OpenID），
	// 则在此处记录接收方的标识。
	To          string `json:"to" gorm:"column:to"` 

	// [Status] 中文名：发送状态
	// 设计目的：监控消息生命周期。0:等待中(pending), 1:已送达(sent), 2:发送失败(failed)。
	Status      int    `json:"status" gorm:"default:0;index"`

	// [OpenId] 中文名：微信 OpenID (别名)
	// 设计目的：为了兼容其他推送接口协议。逻辑上等同于 To，不存入数据库。
	OpenId      string `json:"openid" gorm:"-:all"` 

	// [Desp] 中文名：消息长文本 (别名)
	// 设计目的：主要为了兼容 ServerChan 的接口参数 `desp`。逻辑上对应 Content。
	Desp        string `json:"desp" gorm:"-:all"` 

	// [Short] 中文名：消息短文本 (别名)
	// 设计目的：主要为了兼容老版本接口参数 `short`。逻辑上对应 Description。
	Short       string `json:"short" gorm:"-:all"` 

	// [Async] 中文名：是否异步发送
	// 设计目的：控制发送行为。如果为 true，后端会立即返回结果并将发送任务丢入后台队列，
	// 避免因第三方接口超时导致用户请求卡死。不存入数据库。
	Async       bool   `json:"async" gorm:"-"` 

	// [RenderMode] 中文名：内容渲染模式
	// 设计目的：决定消息在网页端如何展示。
	// markdown (默认解析), code (当做代码块展示), raw (原始文本展示)。
	RenderMode  string `json:"render_mode" gorm:"raw"` 
}

func GetMessageByIds(id int, userId int) (*Message, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id 或 userId 为空！")
	}
	message := Message{Id: id, UserId: userId}
	err := DB.Where(message).First(&message).Error
	return &message, err
}

func GetMessageById(id int) (*Message, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	message := Message{Id: id}
	err := DB.Where(message).First(&message).Error
	return &message, err
}

func GetAsyncPendingMessageIds() (ids []int, err error) {
	err = DB.Model(&Message{}).Where("status = ?", common.MessageSendStatusAsyncPending).Pluck("id", &ids).Error
	return ids, err
}

func GetMessageByLink(link string) (*Message, error) {
	if link == "" {
		return nil, errors.New("link 为空！")
	}
	message := Message{Link: link}
	err := DB.Where(message).First(&message).Error
	return &message, err
}

func GetMessageStatusByLink(link string) (int, error) {
	if link == "" {
		return common.MessageSendStatusUnknown, errors.New("link 为空！")
	}
	message := Message{}
	err := DB.Where("link = ?", link).Select("status").First(&message).Error
	return message.Status, err
}

func GetMessagesByUserId(userId int, startIdx int, num int) (messages []*Message, err error) {
	err = DB.Select([]string{"id", "title", "channel", "timestamp", "status"}).
		Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&messages).Error
	return messages, err
}

func SearchMessages(keyword string) (messages []*Message, err error) {
	err = DB.Select([]string{"id", "title", "channel", "timestamp", "status"}).
		Where("id = ? or title LIKE ? or description LIKE ? or content LIKE ?", keyword, keyword+"%", keyword+"%", keyword+"%").
		Order("id desc").
		Find(&messages).Error
	return messages, err
}

func DeleteMessageById(id int, userId int) (err error) {
	// Why we need userId here? In case user want to delete other's message.
	if id == 0 || userId == 0 {
		return errors.New("id 或 userId 为空！")
	}
	message := Message{Id: id, UserId: userId}
	err = DB.Where(message).First(&message).Error
	if err != nil {
		return err
	}
	return message.Delete()
}

func DeleteAllMessages() error {
	return DB.Exec("DELETE FROM messages").Error
}

func (message *Message) UpdateAndInsert(userId int) error {
	message.Timestamp = time.Now().Unix()
	message.UserId = userId
	message.Status = common.MessageSendStatusPending
	var err error
	err = DB.Create(message).Error
	return err
}

func (message *Message) UpdateStatus(status int) error {
	err := DB.Model(message).Update("status", status).Error
	return err
}

func (message *Message) Delete() error {
	err := DB.Delete(message).Error
	return err
}
