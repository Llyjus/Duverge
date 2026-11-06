package com.duverge.userdata.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.duverge.userdata.entity.Message;

public interface MessageRepository extends JpaRepository<Message, Long> {
    Optional<Message> findByMessageId(String messageId);
}
