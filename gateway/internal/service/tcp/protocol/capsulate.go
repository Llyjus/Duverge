package protocol

import (
	"encoding/json"
)

// CapsulateErrorResponse wraps an error response in a message structure.
func CapsulateErrorResponse(
	errorCode string,
	detail string,
) (
	[]byte,
	error) {
	// capsulate the error response
	errorResponse := ErrorResponse{
		ErrorCode: errorCode,
		Detail:    detail,
	}

	errorResponseData, err := json.Marshal(errorResponse)
	if err != nil {
		return nil, err
	}

	response := Packet{
		Version: Version,
		Type:    "response",
		Success: false,
		Data:    errorResponseData,
	}
	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func CapsulateSuccessResponse() (
	[]byte,
	error) {
	// capsulate the success response
	response := Packet{
		Version: Version,
		Type:    "response",
		Success: true,
		Data:    nil,
	}
	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

// CapsulateChatMessage wraps a chat message in a response structure.
func CapsulateChatMessage(
	requestID string,
	userID string,
	targetUserID string,
	message_type string,
	content string,
) (
	[]byte,
	error) {
	// capsulate the message to send
	chatPost := ChatPost{
		MessageID: requestID,
		SenderID:  userID,
		Type:      message_type,
		Content:   content,
	}

	chatPostData, err := json.Marshal(chatPost)
	if err != nil {
		return nil, err
	}

	response := Packet{
		Version: Version,
		Type:    "chat",
		Success: true,
		Data:    chatPostData,
	}
	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}
