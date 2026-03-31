package model

import (
	"errors"
	"message-pusher/common"
	"strings"
)

// User if you add sensitive fields, don't forget to clean them in setupLogin function.
// Otherwise, the sensitive information will be saved on local storage in plain text!

// type User struct {
// 	Id                    int    `json:"id"`
// 	Username              string `json:"username" gorm:"unique;index" validate:"max=12"`
// 	Password              string `json:"password" gorm:"not null;" validate:"min=8,max=20"`
// 	DisplayName           string `json:"display_name" gorm:"index" validate:"max=20"`
// 	Role                  int    `json:"role" gorm:"type:int;default:1"`   // admin, common
// 	Status                int    `json:"status" gorm:"type:int;default:1"` // enabled, disabled
// 	Token                 string `json:"token"`
// 	Email                 string `json:"email" gorm:"index" validate:"max=50"`
// 	GitHubId              string `json:"github_id" gorm:"column:github_id;index"`
// 	WeChatId              string `json:"wechat_id" gorm:"column:wechat_id;index"`
// 	VerificationCode      string `json:"verification_code" gorm:"-:all"` // this field is only for Email verification, don't save it to database!
// 	Channel               string `json:"channel"`
// 	SendEmailToOthers     int    `json:"send_email_to_others" gorm:"type:int;default:0"`
// 	SaveMessageToDatabase int    `json:"save_message_to_database" gorm:"type:int;default:0"`
// }

type User struct {
	// [Id] 中文名：用户唯一标识符
	// 设计目的：作为数据库主键，用于唯一区分系统中的每一个用户。在关联消息表或渠道表时，
	// 使用该 ID 作为外键，确保数据的一致性。
	Id                    int    `json:"id"`

	// [Username] 中文名：登录用户名
	// 设计目的：用户登录系统时使用的凭证。设置了唯一索引 (unique) 防止重复，
	// 并在业务上限制最大长度为 12 个字符，方便记忆和展示。
	Username              string `json:"username" gorm:"unique;index" validate:"max=12"`

	// [Password] 中文名：登录密码
	// 设计目的：身份验证的核心凭证。数据库中仅存储哈希后的密文（由 gorm:"not null" 约束），
	// 配合 validate 限制长度在 8-20 位，确保账号安全性。
	Password              string `json:"password" gorm:"not null;" validate:"min=8,max=20"`

	// [DisplayName] 中文名：显示名称/昵称
	// 设计目的：在系统界面（如导航栏、欢迎语）中展示给用户看的人性化名称，
	// 避免直接显示有时显得生硬的用户名。
	DisplayName           string `json:"display_name" gorm:"index" validate:"max=20"`

	// [Role] 中文名：用户角色
	// 设计目的：实现简单的权限控制（RBAC）。例如 100 代表超级管理员（Root），
	// 1 代表普通用户。决定了该用户是否能访问管理后台、添加新用户等功能。
	Role                  int    `json:"role" gorm:"type:int;default:1"`   // admin, common

	// [Status] 中文名：账号状态
	// 设计目的：管理账号的生命周期。1 代表启用（Enabled），0 代表禁用（Disabled）。
	// 当需要封禁违规用户时，只需将此值改为 0，该用户将无法通过登录拦截器。
	Status                int    `json:"status" gorm:"type:int;default:1"` // enabled, disabled

	// [Token] 中文名：身份令牌
	// 设计目的：用于 API 调用鉴权。当外部程序（如脚本、监控插件）通过接口推送消息时，
	// 无需提供账号密码，只需在请求头带上此 Token，系统即可识别其身份。
	Token                 string `json:"token"`

	// [Email] 中文名：电子邮箱
	// 设计目的：1. 作为重要的通知渠道；2. 用于账号找回密码或身份验证。
	// 在推送消息时，若未指定渠道，系统常默认推送到此邮箱。
	Email                 string `json:"email" gorm:"index" validate:"max=50"`

	// [GitHubId] 中文名：GitHub 第三方登录 ID
	// 设计目的：支持 OAuth2 第三方登录。存储 GitHub 账号的唯一标识，
	// 实现“点击 GitHub 图标直接登录”的功能，提升用户体验。
	GitHubId              string `json:"github_id" gorm:"column:github_id;index"`

	// [WeChatId] 中文名：微信唯一标识
	// 设计目的：用于关联微信账号。支持微信登录，或在通过“微信渠道”推送消息时，
	// 作为定位接收者的依据。
	WeChatId              string `json:"wechat_id" gorm:"column:wechat_id;index"`

	// [VerificationCode] 中文名：临时验证码
	// 设计目的：仅用于注册或重置密码时的邮箱验证过程。因为标记了 gorm:"-:all"，
	// 它只存在于内存中，不会被存入数据库，确保了敏感验证数据的即时性。
	VerificationCode      string `json:"verification_code" gorm:"-:all"` 

	// [Channel] 中文名：默认推送渠道
	// 设计目的：用户级别的偏好设置。如果推送请求中没有明确指定使用哪个渠道，
	// 系统会自动读取此处设置的默认值（如 "ding", "wechat"）。
	Channel               string `json:"channel"`

	// [SendEmailToOthers] 中文名：允许推送到他人邮箱
	// 设计目的：权限开关。控制该用户是否有权将消息发送到非本人绑定的邮箱，
	// 主要用于防止滥用系统发送垃圾邮件。
	SendEmailToOthers     int    `json:"send_email_to_others" gorm:"type:int;default:0"`

	// [SaveMessageToDatabase] 中文名：消息持久化开关
	// 设计目的：隐私与性能控制。开启后（值为1），该用户发送的所有推送历史都会存入消息表，
	// 允许用户在 Web 后端随时查阅历史记录；关闭则不留底。
	SaveMessageToDatabase int    `json:"save_message_to_database" gorm:"type:int;default:0"`
}

func GetMaxUserId() int {
	var user User
	DB.Last(&user)
	return user.Id
}

func GetAllUsers(startIdx int, num int) (users []*User, err error) {
	err = DB.Order("id desc").Limit(num).Offset(startIdx).Select([]string{"id", "username", "display_name", "role", "status", "email", "send_email_to_others", "save_message_to_database"}).Find(&users).Error
	return users, err
}

func SearchUsers(keyword string) (users []*User, err error) {
	err = DB.Select([]string{"id", "username", "display_name", "role", "status", "email"}).Where("id = ? or username LIKE ? or email LIKE ? or display_name LIKE ?", keyword, keyword+"%", keyword+"%", keyword+"%").Find(&users).Error
	return users, err
}

func GetUserById(id int, selectAll bool) (*User, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	user := User{Id: id}
	var err error = nil
	if selectAll {
		err = DB.First(&user, "id = ?", id).Error
	} else {
		err = DB.Select([]string{"id", "username", "display_name", "role", "status", "email", "wechat_id", "github_id",
			"channel", "token", "save_message_to_database",
		}).First(&user, "id = ?", id).Error
	}
	return &user, err
}

func DeleteUserById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	user := User{Id: id}
	return user.Delete()
}

func (user *User) Insert() error {
	var err error
	if user.Password != "" {
		user.Password, err = common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	err = DB.Create(user).Error
	if err == nil {
		common.UserCount += 1 // We don't need to use atomic here, because it's not a critical value
	}
	return err
}

func (user *User) Update(updatePassword bool) error {
	var err error
	if updatePassword {
		user.Password, err = common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	err = DB.Model(user).Updates(user).Error
	return err
}

func (user *User) Delete() error {
	if user.Id == 0 {
		return errors.New("id 为空！")
	}
	err := DB.Delete(user).Error
	return err
}

// ValidateAndFill check password & user status
func (user *User) ValidateAndFill() (err error) {
	// When querying with struct, GORM will only query with non-zero fields,
	// that means if your field’s value is 0, '', false or other zero values,
	// it won’t be used to build query conditions
	password := user.Password
	if user.Username == "" || password == "" {
		return errors.New("用户名或密码为空")
	}
	DB.Where(User{Username: user.Username}).First(user)
	okay := common.ValidatePasswordAndHash(password, user.Password)
	if !okay || user.Status != common.UserStatusEnabled {
		return errors.New("用户名或密码错误，或用户已被封禁")
	}
	return nil
}

func (user *User) FillUserById() error {
	if user.Id == 0 {
		return errors.New("id 为空！")
	}
	DB.Where(User{Id: user.Id}).First(user)
	return nil
}

func (user *User) FillUserByEmail() error {
	if user.Email == "" {
		return errors.New("email 为空！")
	}
	DB.Where(User{Email: user.Email}).First(user)
	return nil
}

func (user *User) FillUserByGitHubId() error {
	if user.GitHubId == "" {
		return errors.New("GitHub id 为空！")
	}
	DB.Where(User{GitHubId: user.GitHubId}).First(user)
	return nil
}

func (user *User) FillUserByWeChatId() error {
	if user.WeChatId == "" {
		return errors.New("WeChat id 为空！")
	}
	DB.Where(User{WeChatId: user.WeChatId}).First(user)
	return nil
}

func (user *User) FillUserByUsername() error {
	if user.Username == "" {
		return errors.New("username 为空！")
	}
	DB.Where(User{Username: user.Username}).First(user)
	return nil
}

func ValidateUserToken(token string) (user *User) {
	if token == "" {
		return nil
	}
	token = strings.Replace(token, "Bearer ", "", 1)
	user = &User{}
	if DB.Where("token = ?", token).First(user).RowsAffected == 1 {
		return user
	}
	return nil
}

func IsEmailAlreadyTaken(email string) bool {
	return DB.Where("email = ?", email).Find(&User{}).RowsAffected == 1
}

func IsWeChatIdAlreadyTaken(wechatId string) bool {
	return DB.Where("wechat_id = ?", wechatId).Find(&User{}).RowsAffected == 1
}

func IsGitHubIdAlreadyTaken(githubId string) bool {
	return DB.Where("github_id = ?", githubId).Find(&User{}).RowsAffected == 1
}

func IsUsernameAlreadyTaken(username string) bool {
	return DB.Where("username = ?", username).Find(&User{}).RowsAffected == 1
}

func ResetUserPasswordByEmail(email string, password string) error {
	if email == "" || password == "" {
		return errors.New("邮箱地址或密码为空！")
	}
	hashedPassword, err := common.Password2Hash(password)
	if err != nil {
		return err
	}
	err = DB.Model(&User{}).Where("email = ?", email).Update("password", hashedPassword).Error
	return err
}
