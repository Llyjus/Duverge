package com.duverge.userdata.error;

/** Raised when a database operation cannot enter the bounded admission queue. */
public class DatabaseCapacityUnavailableException extends RuntimeException {
    public DatabaseCapacityUnavailableException(String message) {
        super(message);
    }

    public DatabaseCapacityUnavailableException(String message, Throwable cause) {
        super(message, cause);
    }
}
