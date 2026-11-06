package com.duverge.userdata.service;

import org.springframework.stereotype.Service;

import com.duverge.userdata.error.DuplicateMessageIdException;

/** Checks whether a request Message ID has already been recorded. */
@Service
public class IdempotencyService {
    private final MessageService messageService;

    public IdempotencyService(MessageService messageService) {
        this.messageService = messageService;
    }

    /**
     * Rejects a previously recorded ID before business work starts. The unique
     * database constraint remains the final guard against concurrent requests.
     */
    public void ensureNewMessageId(String messageId) {
        if (messageService.findByMessageId(messageId).isPresent()) {
            throw new DuplicateMessageIdException("Message ID has already been recorded", null);
        }
    }
}
