package channel

import (
	"message-pusher/common"
	"message-pusher/model"
)

// [AsyncMessageQueue] 中文名：异步消息队列
// 设计目的：充当“缓冲区”。当大量的推送请求瞬间涌入时，系统不直接去调外部接口（慢），
// 而是把消息的 ID 丢进这个通道里“排队”。利用 chan 的特性保证并发安全。
var AsyncMessageQueue chan int

// [AsyncMessageQueueSize] 中文名：队列容量/长度
// 设计目的：规定了队列最多能积压多少条消息。这里设为 128，意味着如果瞬时有 129 条
// 异步消息，第 129 条会阻塞或报错，防止内存被无限堆积的消息撑爆。
var AsyncMessageQueueSize = 128

// [AsyncMessageSenderNum] 中文名：异步发送者（消费者）数量
// 设计目的：规定了系统同时开启多少个“工位”来处理消息。这里设为 2，意味着系统会
// 永久运行 2 个协程，专门从队列里取任务并执行真正的发送逻辑。
var AsyncMessageSenderNum = 2


func init() {
	AsyncMessageQueue = make(chan int, AsyncMessageQueueSize)
	for i := 0; i < AsyncMessageSenderNum; i++ {
		go asyncMessageSender()
	}
}

// LoadAsyncMessages loads async pending messages from database.
// We have to wait the database connection is ready.
func LoadAsyncMessages() {
	ids, err := model.GetAsyncPendingMessageIds()
	if err != nil {
		common.FatalLog("failed to load async pending messages: " + err.Error())
	}
	for _, id := range ids {
		AsyncMessageQueue <- id
	}
}

func asyncMessageSenderHelper(message *model.Message) error {
	user, err := model.GetUserById(message.UserId, false)
	if err != nil {
		return err
	}
	channel_, err := model.GetChannelByName(message.Channel, user.Id)
	if err != nil {
		return err
	}
	return SendMessage(message, user, channel_)
}

func asyncMessageSender() {
	for {
		id := <-AsyncMessageQueue
		message, err := model.GetMessageById(id)
		if err != nil {
			common.SysError("async message sender error: " + err.Error())
			continue
		}
		err = asyncMessageSenderHelper(message)
		status := common.MessageSendStatusFailed
		if err != nil {
			common.SysError("async message sender error: " + err.Error())
		} else {
			status = common.MessageSendStatusSent
		}
		err = message.UpdateStatus(status)
		if err != nil {
			common.SysError("async message sender error: " + err.Error())
		}
	}
}
