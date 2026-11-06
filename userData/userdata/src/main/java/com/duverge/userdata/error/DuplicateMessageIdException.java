package com.duverge.userdata.error;

/** Raised when a request tries to persist a Message ID that is already recorded. */
public class DuplicateMessageIdException extends RuntimeException {
    public DuplicateMessageIdException(String message, Throwable cause) {
        super(message, cause);
    }
}
