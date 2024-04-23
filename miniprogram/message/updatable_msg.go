package message

import (
	"fmt"

	"github.com/silenceper/wechat/v2/miniprogram/context"
	"github.com/silenceper/wechat/v2/util"
)

const (
	// createActivityURL 创建activity_id
	createActivityURL = "https://api.weixin.qq.com/cgi-bin/message/wxopen/activityid/create?access_token=%s"
	// SendUpdatableMsgURL 修改动态消息
	setUpdatableMsgURL = "https://api.weixin.qq.com/cgi-bin/message/wxopen/updatablemsg/send?access_token=%s"
)

// UpdatableTargetState 动态消息状态
type UpdatableTargetState int

const (
	// TargetStateNotStarted 未开始
	TargetStateNotStarted UpdatableTargetState = 0
	// TargetStateStarted 已开始
	TargetStateStarted UpdatableTargetState = 1
	// TargetStateFinished 已结束
	TargetStateFinished UpdatableTargetState = 2
)

// UpdatableMessage 动态消息
type UpdatableMessage struct {
	*context.Context
}

// NewUpdatableMessage 实例化
func NewUpdatableMessage(ctx *context.Context) *UpdatableMessage {
	return &UpdatableMessage{
		Context: ctx,
	}
}

// CreateActivityID 创建activity_id
func (updatableMessage *UpdatableMessage) CreateActivityID() (res CreateActivityIDResponse, err error) {
	accessToken, err := updatableMessage.GetAccessToken()
	if err != nil {
		return
	}

	uri := fmt.Sprintf(createActivityURL, accessToken)
	response, err := util.HTTPGet(uri)
	if err != nil {
		return
	}
	err = util.DecodeWithError(response, &res, "CreateActivityID")
	return
}

// SetUpdatableMsg 修改动态消息
func (updatableMessage *UpdatableMessage) SetUpdatableMsg(activityID string, targetState UpdatableTargetState, template UpdatableMsgTemplate) (err error) {
	accessToken, err := updatableMessage.GetAccessToken()
	if err != nil {
		return
	}

	uri := fmt.Sprintf(setUpdatableMsgURL, accessToken)
	data := SendUpdatableMsgReq{
		ActivityID:   activityID,
		TargetState:  targetState,
		TemplateInfo: template,
	}

	response, err := util.PostJSON(uri, data)
	if err != nil {
		return
	}
	return util.DecodeWithCommonError(response, "SendUpdatableMsg")
}

// CreateActivityIDResponse 创建activity_id 返回
type CreateActivityIDResponse struct {
	util.CommonError

	ActivityID     string `json:"activity_id"`
	ExpirationTime int64  `json:"expiration_time"`
}

// UpdatableMsgTemplate 动态消息模板
type UpdatableMsgTemplate struct {
	ParameterList []UpdatableMsgParameter `json:"parameter_list"`
}

// UpdatableMsgParameter 动态消息参数
type UpdatableMsgParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SendUpdatableMsgReq 修改动态消息参数
type SendUpdatableMsgReq struct {
	ActivityID   string               `json:"activity_id"`
	TemplateInfo UpdatableMsgTemplate `json:"template_info"`
	TargetState  UpdatableTargetState `json:"target_state"`
}
