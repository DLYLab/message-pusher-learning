package model

import (
	"errors"
)

// WebhookConstructRule Keep compatible with Message
type WebhookConstructRule struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	URL         string `json:"url"`
}

// type Webhook struct {
// 	Id            int    `json:"id"`
// 	UserId        int    `json:"user_id" gorm:"index"`
// 	Name          string `json:"name" gorm:"type:varchar(32);index"`
// 	Status        int    `json:"status" gorm:"default:1"` // enabled, disabled
// 	Link          string `json:"link" gorm:"type:char(32);uniqueIndex"`
// 	CreatedTime   int64  `json:"created_time" gorm:"bigint"`
// 	ExtractRule   string `json:"extract_rule" gorm:"not null"`              // how we extract key info from the request
// 	ConstructRule string `json:"construct_rule" gorm:"not null"`            // how we construct message with the extracted info
// 	Channel       string `json:"channel" gorm:"type:varchar(32); not null"` // which channel to send our message
// }

type Webhook struct {
	// [Id] 中文名：Webhook 配置唯一 ID
	// 设计目的：数据库自增主键，用于区分不同的 Webhook 转换逻辑配置。
	Id            int    `json:"id"`

	// [UserId] 中文名：所属用户 ID
	// 设计目的：标识该 Webhook 配置属于哪个用户。加了索引 (index) 
	// 方便在用户后台快速拉取其创建的所有 Webhook 列表。
	UserId        int    `json:"user_id" gorm:"index"`

	// [Name] 中文名：Webhook 名称
	// 设计目的：用户给这个转换规则起的名称（如 "GitHub提交提醒"），
	// 方便在管理界面识别。
	Name          string `json:"name" gorm:"type:varchar(32);index"`

	// [Status] 中文名：启用状态
	// 设计目的：控制该 Webhook 接口是否处于激活状态。1 为正常接收，0 为停止响应。
	Status        int    `json:"status" gorm:"default:1"` // enabled, disabled

	// [Link] 中文名：唯一的 Webhook 接收地址后缀
	// 设计目的：系统会生成一个唯一的随机字符串（如 UUID），拼接到固定域名后
	// 形成一个独一无二的 URL（如 /receive/abc-123）。设置为唯一索引以防冲突。
	Link          string `json:"link" gorm:"type:char(32);uniqueIndex"`

	// [CreatedTime] 中文名：创建时间
	// 设计目的：记录该 Webhook 配置的创建时刻。
	CreatedTime   int64  `json:"created_time" gorm:"bigint"`

	// [ExtractRule] 中文名：数据提取规则 (JSONPath/正则)
	// 设计目的：【核心逻辑】定义如何从第三方发来的 JSON 报文中提取关键信息。
	// 例如：从 GitHub 的 Payload 中提取 "repository.name" 和 "pusher.name"。
	ExtractRule   string `json:"extract_rule" gorm:"not null"`

	// [ConstructRule] 中文名：消息构建规则 (模板)
	// 设计目的：【核心逻辑】定义如何利用提取出的变量生成最终的推送文本。
	// 例如：将变量填入模板 "项目 {repo} 被 {user} 提交了代码"。
	ConstructRule string `json:"construct_rule" gorm:"not null"`

	// [Channel] 中文名：转发目标渠道
	// 设计目的：指定转换后的消息最终通过哪个渠道（Channel）发出去。
	// 例如：GitHub 消息提取后，自动通过“钉钉渠道”发送。
	Channel       string `json:"channel" gorm:"type:varchar(32); not null"`
}

func GetWebhookById(id int, userId int) (*Webhook, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id 或 userId 为空！")
	}
	c := Webhook{Id: id, UserId: userId}
	err := DB.Where(c).First(&c).Error
	return &c, err
}

func GetWebhookByLink(link string) (*Webhook, error) {
	if link == "" {
		return nil, errors.New("link 为空！")
	}
	c := Webhook{Link: link}
	err := DB.Where(c).First(&c).Error
	return &c, err
}

func GetWebhooksByUserId(userId int, startIdx int, num int) (webhooks []*Webhook, err error) {
	err = DB.Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&webhooks).Error
	return webhooks, err
}

func SearchWebhooks(userId int, keyword string) (webhooks []*Webhook, err error) {
	err = DB.Where("user_id = ?", userId).Where("id = ? or link = ? or name LIKE ?", keyword, keyword, keyword+"%").Find(&webhooks).Error
	return webhooks, err
}

func DeleteWebhookById(id int, userId int) (c *Webhook, err error) {
	// Why we need userId here? In case user want to delete other's c.
	if id == 0 || userId == 0 {
		return nil, errors.New("id 或 userId 为空！")
	}
	c = &Webhook{Id: id, UserId: userId}
	err = DB.Where(c).First(&c).Error
	if err != nil {
		return nil, err
	}
	return c, c.Delete()
}

func (webhook *Webhook) Insert() error {
	var err error
	err = DB.Create(webhook).Error
	return err
}

func (webhook *Webhook) UpdateStatus(status int) error {
	err := DB.Model(webhook).Update("status", status).Error
	return err
}

// Update Make sure your token's fields is completed, because this will update zero values
func (webhook *Webhook) Update() error {
	var err error
	err = DB.Model(webhook).Select("status", "name", "extract_rule", "construct_rule", "channel").Updates(webhook).Error
	return err
}

func (webhook *Webhook) Delete() error {
	err := DB.Delete(webhook).Error
	return err
}
