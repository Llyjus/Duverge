package com.duverge.userdata.repository;

import java.util.Optional;
import java.time.LocalDateTime;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import com.duverge.userdata.entity.Message;

public interface MessageRepository extends JpaRepository<Message, Long> {
    Optional<Message> findByMessageId(String messageId);

    @Modifying
    @Query("delete from Message message where message.createdAt < :cutoff")
    int deleteCreatedBefore(@Param("cutoff") LocalDateTime cutoff);
}
