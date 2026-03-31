package channel

import (
	"message-pusher/common"
	"message-pusher/model"
	"sync"
	"time"
)


// TokenStoreItem 定义了所有第三方平台令牌（如微信 AccessToken）必须具备的标准行为。
// 通过接口设计，系统可以无视具体的平台差异（微信、飞书、钉钉），统一进行自动化管理。
type TokenStoreItem interface {
	
	// [Key] 中文名：存储索引键
	// 目的：返回该 Token 在内存 Map 中的唯一标识（通常是 "渠道类型:渠道ID"）。
	// 作用：让 TokenStore 能够精准地定位、存取或更新特定的令牌，避免数据混淆。
	Key() string

	// [Token] 中文名：获取当前可用令牌
	// 目的：返回当前存储在内存中、且处于有效期内的 AccessToken 字符串。
	// 作用：发送消息的业务逻辑（如微信推送函数）会调用此方法获取“通关文牒”。
	Token() string

	// [Refresh] 中文名：执行令牌刷新
	// 目的：【核心方法】负责发起 HTTP 请求访问第三方 API（如微信官网），获取新 Token。
	// 作用：由后台协程定时触发。它会更新结构体内部的 Token 字符串及过期时间，
	// 实现“无感续期”，保证推送服务不中断。
	Refresh()

	// [IsFilled] 中文名：配置完整性检查
	// 目的：检查该渠道是否已经填写了必要的认证信息（如 AppId, AppSecret）。
	// 作用：作为 Refresh 的前置校验。如果配置不全，系统将跳过该渠道，
	// 防止产生无效的网络请求或因无效参数被第三方平台暂时封禁。
	IsFilled() bool

	// [IsShared] 中文名：共享模式标识
	// 目的：标识该 Token 是否在多个不同的发送渠道之间共享（基于相同的 AppId）。
	// 作用：用于优化频率限制（Rate Limit）。如果多个渠道共用一个 Token，
	// 系统会避免重复刷新，节省 API 调用额度，防止触发平台的频率惩罚。
	IsShared() bool
}

type tokenStore struct {
	Map               map[string]*TokenStoreItem
	Mutex             sync.RWMutex 
	ExpirationSeconds int  // 失效时间长度（s）
}

var s tokenStore  // 在内存中维护一个token存储表

func channel2item(channel_ *model.Channel) TokenStoreItem {
	switch channel_.Type {
	case model.TypeWeChatTestAccount:
		item := &WeChatTestAccountTokenStoreItem{
			AppID:     channel_.AppId,
			AppSecret: channel_.Secret,
		}
		return item
	case model.TypeWeChatCorpAccount:
		corpId, agentId, err := parseWechatCorpAccountAppId(channel_.AppId)
		if err != nil {
			common.SysError(err.Error())
			return nil
		}
		item := &WeChatCorpAccountTokenStoreItem{
			CorpId:      corpId,
			AgentSecret: channel_.Secret,
			AgentId:     agentId,
		}
		return item
	case model.TypeLarkApp:
		item := &LarkAppTokenStoreItem{
			AppID:     channel_.AppId,
			AppSecret: channel_.Secret,
		}
		return item
	}
	return nil
}

func channels2items(channels []*model.Channel) []TokenStoreItem {
	var items []TokenStoreItem
	for _, channel_ := range channels {
		item := channel2item(channel_)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

func TokenStoreInit() {  // 管理具有有效期的第三方令牌（Token）
	s.Map = make(map[string]*TokenStoreItem)
	// https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Get_access_token.html
	// https://developer.work.weixin.qq.com/document/path/91039
	s.ExpirationSeconds = 2 * 55 * 60 // 2 hours - 5 minutes
	go func() {  // 开启一个协程
		channels, err := model.GetTokenStoreChannels() // 去channels表中查找有token的通道（TypeWeChatCorpAccount, TypeWeChatTestAccount, TypeLarkApp），返回查到的结构体
		if err != nil {
			common.FatalLog(err.Error())
		}
		items := channels2items(channels)
		s.Mutex.RLock()
		for i := range items {
			// s.Map[item.Key()] = &item  // This is wrong, you are getting the address of a local variable!
			s.Map[items[i].Key()] = &items[i]
		}
		s.Mutex.RUnlock()
		for { // 不断检测token是否过期（每2 * 55 * 60s 更新一下） // 面试考点：读写分离的快照模式
			s.Mutex.RLock()
			var tmpMap = make(map[string]*TokenStoreItem)  // 副本表
			for k, v := range s.Map {
				tmpMap[k] = v
			}
			s.Mutex.RUnlock()
			for k := range tmpMap {
				(*tmpMap[k]).Refresh()  // 更新token，因为tmpmap的值是指针，可以实现直接修改原表中的token
			}
			s.Mutex.RLock()
			// we shouldn't directly replace the old map with the new map, cause the old map's keys may already change
			for k := range s.Map { 
				v, okay := tmpMap[k]
				if okay {
					s.Map[k] = v
				}
			}
			sleepDuration := common.Max(s.ExpirationSeconds, 60)
			s.Mutex.RUnlock()
			time.Sleep(time.Duration(sleepDuration) * time.Second)
		}
	}()
}

// TokenStoreAddItem It's okay to add an incomplete item.
func TokenStoreAddItem(item TokenStoreItem) {
	if !item.IsFilled() {
		return
	}
	item.Refresh()
	s.Mutex.RLock()
	s.Map[item.Key()] = &item
	s.Mutex.RUnlock()
}

func TokenStoreRemoveItem(item TokenStoreItem) {
	s.Mutex.RLock()
	delete(s.Map, item.Key())
	s.Mutex.RUnlock()
}

func TokenStoreAddUser(user *model.User) {
	channels, err := model.GetTokenStoreChannelsByUserId(user.Id)
	if err != nil {
		common.SysError(err.Error())
		return
	}
	items := channels2items(channels)
	for i := range items {
		TokenStoreAddItem(items[i])
	}
}

// TokenStoreRemoveUser
// user must be filled.
// It's okay to delete a user that don't have an item here.
func TokenStoreRemoveUser(user *model.User) {
	channels, err := model.GetTokenStoreChannelsByUserId(user.Id)
	if err != nil {
		common.SysError(err.Error())
		return
	}
	items := channels2items(channels)
	for i := range items {
		if items[i].IsShared() {
			continue
		}
		TokenStoreRemoveItem(items[i])
	}
}

func checkTokenStoreChannelType(channelType string) bool {
	return channelType == model.TypeWeChatTestAccount || channelType == model.TypeWeChatCorpAccount || channelType == model.TypeLarkApp
}

func TokenStoreAddChannel(channel *model.Channel) {
	if !checkTokenStoreChannelType(channel.Type) {
		return
	}
	item := channel2item(channel)
	if item != nil {
		// Do not check IsShared here, cause its useless
		TokenStoreAddItem(item)
	}
}

func TokenStoreRemoveChannel(channel *model.Channel) {
	if !checkTokenStoreChannelType(channel.Type) {
		return
	}
	item := channel2item(channel)
	if item != nil && !item.IsShared() {
		TokenStoreRemoveItem(item)
	}
}

func TokenStoreUpdateChannel(newChannel *model.Channel, oldChannel *model.Channel) {
	if oldChannel.Type != model.TypeWeChatTestAccount && oldChannel.Type != model.TypeWeChatCorpAccount {
		return
	}
	// Why so complicated? Because the given channel maybe incomplete.
	if oldChannel.Type == model.TypeWeChatTestAccount {
		// Only keep changed parts
		if newChannel.AppId == oldChannel.AppId {
			newChannel.AppId = ""
		}
		if newChannel.Secret == oldChannel.Secret {
			newChannel.Secret = ""
		}
		oldItem := WeChatTestAccountTokenStoreItem{
			AppID:     oldChannel.AppId,
			AppSecret: oldChannel.Secret,
		}
		// Yeah, it's a deep copy.
		newItem := oldItem
		// This means the user updated those fields.
		if newChannel.AppId != "" {
			newItem.AppID = newChannel.AppId
		}
		if newChannel.Secret != "" {
			newItem.AppSecret = newChannel.Secret
		}
		if !oldItem.IsShared() {
			TokenStoreRemoveItem(&oldItem)
		}
		TokenStoreAddItem(&newItem)
		return
	}
	if oldChannel.Type == model.TypeWeChatCorpAccount {
		// Only keep changed parts
		if newChannel.AppId == oldChannel.AppId {
			newChannel.AppId = ""
		}
		if newChannel.Secret == oldChannel.Secret {
			newChannel.Secret = ""
		}
		corpId, agentId, err := parseWechatCorpAccountAppId(oldChannel.AppId)
		if err != nil {
			common.SysError(err.Error())
			return
		}
		oldItem := WeChatCorpAccountTokenStoreItem{
			CorpId:      corpId,
			AgentSecret: oldChannel.Secret,
			AgentId:     agentId,
		}
		// Yeah, it's a deep copy.
		newItem := oldItem
		// This means the user updated those fields.
		if newChannel.AppId != "" {
			corpId, agentId, err := parseWechatCorpAccountAppId(oldChannel.AppId)
			if err != nil {
				common.SysError(err.Error())
				return
			}
			newItem.CorpId = corpId
			newItem.AgentId = agentId
		}
		if newChannel.Secret != "" {
			newItem.AgentSecret = newChannel.Secret
		}
		if !oldItem.IsShared() {
			TokenStoreRemoveItem(&oldItem)
		}
		TokenStoreAddItem(&newItem)
		return
	}
}

func TokenStoreGetToken(key string) string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	item, ok := s.Map[key]
	if ok {
		return (*item).Token()
	}
	common.SysError("token for " + key + " is blank!")
	return ""
}
