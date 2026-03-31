package model

import (
	"message-pusher/common"
	"strconv"
	"strings"
)

// type Option struct {
// 	Key   string `json:"key" gorm:"primaryKey"`
// 	Value string `json:"value"`
// }

type Option struct {
	// [Key] 中文名：配置项键名（唯一标识）
	// 设计目的：作为数据库的主键 (primaryKey)。它代表了配置的具体名称，
	// 例如 "ServerAddress"（服务器地址）或 "GitHubClientId"（GitHub客户端ID）。
	// 程序通过这个唯一的 Key 在代码中查找对应的配置值。
	Key   string `json:"key" gorm:"primaryKey"`

	// [Value] 中文名：配置项的具体值
	// 设计目的：存储该配置项对应的实际内容。
	// 例如：当 Key 为 "AppVersion" 时，Value 可能是 "v0.5.1"。
	// 所有的配置值都以字符串形式存储，程序在使用时会根据需要转换为数字或布尔值。
	Value string `json:"value"`
}

func AllOption() ([]*Option, error) {
	var options []*Option
	var err error
	err = DB.Find(&options).Error
	return options, err
}

func InitOptionMap() {
	common.OptionMapRWMutex.Lock()
	common.OptionMap = make(map[string]string)
	common.OptionMap["FileUploadPermission"] = strconv.Itoa(common.FileUploadPermission)
	common.OptionMap["FileDownloadPermission"] = strconv.Itoa(common.FileDownloadPermission)
	common.OptionMap["ImageUploadPermission"] = strconv.Itoa(common.ImageUploadPermission)
	common.OptionMap["ImageDownloadPermission"] = strconv.Itoa(common.ImageDownloadPermission)
	common.OptionMap["PasswordLoginEnabled"] = strconv.FormatBool(common.PasswordLoginEnabled)
	common.OptionMap["PasswordRegisterEnabled"] = strconv.FormatBool(common.PasswordRegisterEnabled)
	common.OptionMap["EmailVerificationEnabled"] = strconv.FormatBool(common.EmailVerificationEnabled)
	common.OptionMap["GitHubOAuthEnabled"] = strconv.FormatBool(common.GitHubOAuthEnabled)
	common.OptionMap["WeChatAuthEnabled"] = strconv.FormatBool(common.WeChatAuthEnabled)
	common.OptionMap["TurnstileCheckEnabled"] = strconv.FormatBool(common.TurnstileCheckEnabled)
	common.OptionMap["RegisterEnabled"] = strconv.FormatBool(common.RegisterEnabled)
	common.OptionMap["MessagePersistenceEnabled"] = strconv.FormatBool(common.MessagePersistenceEnabled)
	common.OptionMap["MessageRenderEnabled"] = strconv.FormatBool(common.MessageRenderEnabled)
	common.OptionMap["SMTPServer"] = ""
	common.OptionMap["SMTPAccount"] = ""
	common.OptionMap["SMTPPort"] = strconv.Itoa(common.SMTPPort)
	common.OptionMap["SMTPToken"] = ""
	common.OptionMap["Notice"] = ""
	common.OptionMap["About"] = ""
	common.OptionMap["Footer"] = common.Footer
	common.OptionMap["HomePageLink"] = common.HomePageLink
	common.OptionMap["ServerAddress"] = ""
	common.OptionMap["GitHubClientId"] = ""
	common.OptionMap["GitHubClientSecret"] = ""
	common.OptionMap["WeChatServerAddress"] = ""
	common.OptionMap["WeChatServerToken"] = ""
	common.OptionMap["WeChatAccountQRCodeImageURL"] = ""
	common.OptionMap["TurnstileSiteKey"] = ""
	common.OptionMap["TurnstileSecretKey"] = ""
	common.OptionMapRWMutex.Unlock()
	options, _ := AllOption()
	for _, option := range options {
		updateOptionMap(option.Key, option.Value)
	}
}

func UpdateOption(key string, value string) error {
	// Save to database first
	option := Option{
		Key: key,
	}
	// https://gorm.io/docs/update.html#Save-All-Fields
	DB.FirstOrCreate(&option, Option{Key: key})
	option.Value = value
	// Save is a combination function.
	// If save value does not contain primary key, it will execute Create,
	// otherwise it will execute Update (with all fields).
	DB.Save(&option)
	// Update OptionMap
	updateOptionMap(key, value)
	return nil
}

func updateOptionMap(key string, value string) {
	common.OptionMapRWMutex.Lock()
	defer common.OptionMapRWMutex.Unlock()
	common.OptionMap[key] = value  // 内存的配置更新
	if strings.HasSuffix(key, "Permission") {  //
		intValue, _ := strconv.Atoi(value)
		switch key {
		case "FileUploadPermission":
			common.FileUploadPermission = intValue
		case "FileDownloadPermission":
			common.FileDownloadPermission = intValue
		case "ImageUploadPermission":
			common.ImageUploadPermission = intValue
		case "ImageDownloadPermission":
			common.ImageDownloadPermission = intValue
		}
	}
	if strings.HasSuffix(key, "Enabled") {
		boolValue := value == "true"
		switch key {
		case "PasswordRegisterEnabled":
			common.PasswordRegisterEnabled = boolValue
		case "PasswordLoginEnabled":
			common.PasswordLoginEnabled = boolValue
		case "EmailVerificationEnabled":
			common.EmailVerificationEnabled = boolValue
		case "GitHubOAuthEnabled":
			common.GitHubOAuthEnabled = boolValue
		case "WeChatAuthEnabled":
			common.WeChatAuthEnabled = boolValue
		case "TurnstileCheckEnabled":
			common.TurnstileCheckEnabled = boolValue
		case "RegisterEnabled":
			common.RegisterEnabled = boolValue
		case "MessagePersistenceEnabled":
			common.MessagePersistenceEnabled = boolValue
		case "MessageRenderEnabled":
			common.MessageRenderEnabled = boolValue
		}
	}
	switch key {
	case "SMTPServer":
		common.SMTPServer = value
	case "SMTPPort":
		intValue, _ := strconv.Atoi(value)
		common.SMTPPort = intValue
	case "SMTPAccount":
		common.SMTPAccount = value
	case "SMTPToken":
		common.SMTPToken = value
	case "ServerAddress":
		common.ServerAddress = value
	case "GitHubClientId":
		common.GitHubClientId = value
	case "GitHubClientSecret":
		common.GitHubClientSecret = value
	case "Footer":
		common.Footer = value
	case "HomePageLink":
		common.HomePageLink = value
	case "WeChatServerAddress":
		common.WeChatServerAddress = value
	case "WeChatServerToken":
		common.WeChatServerToken = value
	case "WeChatAccountQRCodeImageURL":
		common.WeChatAccountQRCodeImageURL = value
	case "TurnstileSiteKey":
		common.TurnstileSiteKey = value
	case "TurnstileSecretKey":
		common.TurnstileSecretKey = value
	}
}
