package com.duverge.userdata.error;

/** Raised when a BCrypt operation cannot enter the bounded waiting area in time. */
public class PasswordHashUnavailableException extends RuntimeException {
    public PasswordHashUnavailableException(String message) {
        super(message);
    }

    public PasswordHashUnavailableException(String message, Throwable cause) {
        super(message, cause);
    }
}
