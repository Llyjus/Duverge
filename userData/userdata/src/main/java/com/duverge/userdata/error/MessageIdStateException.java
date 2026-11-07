package com.duverge.userdata.error;

/** Raised when a Message ID is missing from Redis or is not in the expected initial state. */
public class MessageIdStateException extends RuntimeException {
    public MessageIdStateException(String message) {
        super(message);
    }
}
