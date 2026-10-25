package handler

import (
	"errors"
	"io"
	"net"

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

	case errors.Is(err, protocol.ErrServerClosed):
		return protocol.ErrCodeServerClosed, "服务器已关闭"

	default:
		return protocol.ErrCodeInternal, "内部服务器错误"
	}
}

func mapAuthError(err error) (string, string) {
	switch {
	case errors.Is(err, protocol.ErrInvalidMessage):
		return protocol.ErrCodeInvalidMessage, "非法认证信息"

	case errors.Is(err, protocol.ErrAuthenticationFailed):
		return protocol.ErrCodeAuthentication, "身份验证失败"

	case errors.Is(err, protocol.ErrInternal):
		return protocol.ErrCodeInternal, "内部服务器错误"

	default:
		return protocol.ErrCodeInternal, "内部服务器错误"
	}
}

func mapConnectionError(err error) (string, string) {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return protocol.ErrCodeServerClosed, "连接已关闭"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return protocol.ErrCodeServerClosed, "连接超时"
	}
	return protocol.ErrCodeInternal, "服务器连接错误，请稍后重试"
}
