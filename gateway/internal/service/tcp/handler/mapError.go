package handler

import (
	"errors"

	"gateway/internal/service/tcp/protocol"
)

func mapChatError(err error) (string, string) {
	switch {
	case errors.Is(err, protocol.ErrInvalidMessage):
		return protocol.ErrCodeInvalidMessage, "非法聊天信息"

	case errors.Is(err, protocol.ErrForbidden):
		return protocol.ErrCodeForbidden, "禁止的操作"

	case errors.Is(err, protocol.ErrNotFound):
		return protocol.ErrCodeNotFound, "资源未找到"

	case errors.Is(err, protocol.ServerErr):
		return protocol.ErrCodeServerErr, "服务器错误"

	default:
		return protocol.ErrCodeInternal, "内部服务器错误"
	}
}
