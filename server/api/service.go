package api

import (
	"zmessage/server/modules/media"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
)

// NewService 创建API服务
func NewService(
	userSvc user.Service,
	msgSvc message.Service,
	mediaSvc media.Service,
	wsMgr WSManager,
) Service {
	return &service{
		userSvc:  userSvc,
		msgSvc:   msgSvc,
		mediaSvc: mediaSvc,
		wsMgr:    wsMgr,
	}
}

// service API服务实现
type service struct {
	userSvc  user.Service
	msgSvc    message.Service
	mediaSvc  media.Service
	wsMgr     WSManager
}

// UserService 获取用户服务
func (s *service) UserService() user.Service {
	return s.userSvc
}

// MessageService 获取消息服务
func (s *service) MessageService() message.Service {
	return s.msgSvc
}

// MediaService 获取媒体服务
func (s *service) MediaService() media.Service {
	return s.mediaSvc
}

// WSManager 获取WebSocket管理器
func (s *service) WSManager() WSManager {
	return s.wsMgr
}
