package com.duverge.userdata.repository;

import org.springframework.data.jpa.repository.JpaRepository;

import com.duverge.userdata.entity.User;

public interface UserRepository extends JpaRepository<User, Long> {
    public User findByAccountId(String accountId);
}
