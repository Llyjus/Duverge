package com.duverge.userdata.entity;

import java.time.LocalDateTime;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;

@Entity
@Table(
    name = "messages",
    uniqueConstraints = @UniqueConstraint(
        name = "uk_messages_id",
        columnNames = "message_id"
    )
)
public class Message {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(
        name = "message_id",
        nullable = false,
        updatable = false
    )
    private String messageId;

    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;

    protected Message() {
    }

    public Message(String messageId) {
        this.messageId = messageId;
        this.createdAt = LocalDateTime.now();
    }

    public Long getId() {
        return id;
    }

    public String getMessageId() {
        return messageId;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }
}
