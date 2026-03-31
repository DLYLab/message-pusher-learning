package model

import (
	"errors"
	"message-pusher/common"
)

const (
	TypeEmail             = "email"
	TypeWeChatTestAccount = "test"
	TypeWeChatCorpAccount = "corp_app"
	TypeCorp              = "corp"
	TypeLark              = "lark"
	TypeDing              = "ding"
	TypeTelegram          = "telegram"
	TypeDiscord           = "discord"
	TypeBark              = "bark"
	TypeClient            = "client"
	TypeNone              = "none"
	TypeOneBot            = "one_bot"
	TypeGroup             = "group"
	TypeLarkApp           = "lark_app"
	TypeCustom            = "custom"
	TypeTencentAlarm      = "tencent_alarm"
)

// type Channel struct {
// 	Id          int     `json:"id"`
// 	Type        string  `json:"type" gorm:"type:varchar(32)"`
// 	UserId      int     `json:"user_id" gorm:"uniqueIndex:name_user_id;index"`
// 	Name        string  `json:"name" gorm:"type:varchar(32);uniqueIndex:name_user_id"`
// 	Description string  `json:"description"`
// 	Status      int     `json:"status" gorm:"default:1"` // enabled, disabled
// 	Secret      string  `json:"secret" gorm:"index"`
// 	AppId       string  `json:"app_id"`
// 	AccountId   string  `json:"account_id"`
// 	URL         string  `json:"url" gorm:"column:url"`
// 	Other       string  `json:"other"`
// 	CreatedTime int64   `json:"created_time" gorm:"bigint"`
// 	Token       *string `json:"token" gorm:"token"`
// }

type Channel struct {
	// [Id] 中文名：渠道唯一 ID
	// 设计目的：数据库自增主键，用于唯一标识每一个推送配置。
	Id          int     `json:"id"`

	// [Type] 中文名：渠道类型
	// 设计目的：标识该渠道的具体平台，例如 "ding"（钉钉）、"wechat"（企业微信）、"email"（邮件）。
	// 程序根据这个类型来决定调用哪套发送逻辑。
	Type        string  `json:"type" gorm:"type:varchar(32)"`

	// [UserId] 中文名：所属用户 ID
	// 设计目的：标识该配置属于哪个用户。配合 Name 组成联合唯一索引 (name_user_id)，
	// 确保同一个用户下不能有重名的渠道。
	UserId      int     `json:"user_id" gorm:"uniqueIndex:name_user_id;index"`

	// [Name] 中文名：渠道自定义名称
	// 设计目的：用户给这个配置起的外号，比如“我的运维报警群”。推送时可以直接通过名称调用。
	Name        string  `json:"name" gorm:"type:varchar(32);uniqueIndex:name_user_id"`

	// [Description] 中文名：渠道描述
	// 设计目的：用户对该配置的备注说明，方便在管理界面区分多个相似的配置。
	Description string  `json:"description"`

	// [Status] 中文名：渠道启用状态
	// 设计目的：控制该推送配置是否生效。1 为启用，0 为禁用。
	Status      int     `json:"status" gorm:"default:1"` // enabled, disabled

	// [Secret] 中文名：安全密钥 / Secret
	// 设计目的：存储第三方平台提供的密钥。例如钉钉机器人的加签密钥、
	// 企业微信的 Secret 等，用于发送请求时的加密认证。
	Secret      string  `json:"secret" gorm:"index"`

	// [AppId] 中文名：应用 ID / AgentId
	// 设计目的：存储第三方平台的应用标识。例如企业微信的 AgentId 
	// 或公众号的 AppId。
	AppId       string  `json:"app_id"`

	// [AccountId] 中文名：账号标识 / 企业 ID
	// 设计目的：存储平台级的 ID。例如企业微信的 CorpId 或某些平台的 API Key。
	AccountId   string  `json:"account_id"`

	// [URL] 中文名：Webhook 地址 / API 终点
	// 设计目的：存储推送的目标 URL。例如钉钉机器人的完整 Webhook 链接，
	// 或者自建网关的 API 地址。
	URL         string  `json:"url" gorm:"column:url"`

	// [Other] 中文名：备用/扩展配置
	// 设计目的：一个“杂物箱”字段。如果某个平台有特殊的额外参数（如端口号、SMTP服务器），
	// 通常会以 JSON 格式或特殊字符串存在这里。
	Other       string  `json:"other"`

	// [CreatedTime] 中文名：创建时间
	// 设计目的：记录该渠道配置建立的时间戳，方便管理和排序。
	CreatedTime int64   `json:"created_time" gorm:"bigint"`

	// [Token] 中文名：访问令牌 (Access Token)
	// 设计目的：存储某些平台需要持久化保存的 Token。使用指针类型 (*string) 
	// 是为了方便处理数据库中的 NULL 值。
	Token       *string `json:"token" gorm:"token"`
}

type BriefChannel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetChannelById(id int, userId int, selectAll bool) (*Channel, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id 或 userId 为空！")
	}
	c := Channel{Id: id, UserId: userId}
	var err error
	if selectAll {
		err = DB.Where(c).First(&c).Error
	} else {
		err = DB.Omit("secret").Where(c).First(&c).Error
	}
	return &c, err
}

func GetChannelByName(name string, userId int) (*Channel, error) {
	if name == "" || userId == 0 {
		return nil, errors.New("name 或 userId 为空！")
	}
	c := Channel{Name: name, UserId: userId}
	err := DB.Where(c).First(&c).Error
	return &c, err
}

func GetTokenStoreChannels() (channels []*Channel, err error) {
	err = DB.Where("type in ?", []string{TypeWeChatCorpAccount, TypeWeChatTestAccount, TypeLarkApp}).Find(&channels).Error
	return channels, err
}

func GetTokenStoreChannelsByUserId(userId int) (channels []*Channel, err error) {
	err = DB.Where("user_id = ?", userId).Where("type = ? or type = ?", TypeWeChatCorpAccount, TypeWeChatTestAccount).Find(&channels).Error
	return channels, err
}

func GetChannelsByUserId(userId int, startIdx int, num int) (channels []*Channel, err error) {
	err = DB.Omit("secret").Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&channels).Error
	return channels, err
}

func GetBriefChannelsByUserId(userId int) (channels []*BriefChannel, err error) {
	err = DB.Model(&Channel{}).Select("id", "name", "description").Where("user_id = ? and status = ?", userId, common.ChannelStatusEnabled).Find(&channels).Error
	return channels, err
}

func SearchChannels(userId int, keyword string) (channels []*Channel, err error) {
	err = DB.Omit("secret").Where("user_id = ?", userId).Where("id = ? or name LIKE ?", keyword, keyword+"%").Find(&channels).Error
	return channels, err
}

func DeleteChannelById(id int, userId int) (c *Channel, err error) {
	// Why we need userId here? In case user want to delete other's c.
	if id == 0 || userId == 0 {
		return nil, errors.New("id 或 userId 为空！")
	}
	c = &Channel{Id: id, UserId: userId}
	err = DB.Where(c).First(&c).Error
	if err != nil {
		return nil, err
	}
	return c, c.Delete()
}

func (channel *Channel) Insert() error {
	var err error
	err = DB.Create(channel).Error
	return err
}

func (channel *Channel) UpdateStatus(status int) error {
	err := DB.Model(channel).Update("status", status).Error
	return err
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (channel *Channel) Update() error {
	var err error
	err = DB.Model(channel).Select("type", "name", "description", "secret", "app_id", "account_id", "url", "other", "status", "token").Updates(channel).Error
	return err
}

func (channel *Channel) Delete() error {
	err := DB.Delete(channel).Error
	return err
}
