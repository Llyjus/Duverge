package com.duverge.userdata.service;

import org.springframework.stereotype.Service;

import com.duverge.userdata.entity.User;
import com.duverge.userdata.repository.UserRepository;

@Service
public class UserService {

    final UserRepository userRepository;

    public UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    public boolean setNewUser(String accountId, String password) {

        User user = new User();

        user.setAccountId(accountId);
        user.setPassword(password);

        userRepository.save(user);
        // deal with false of connection

        return true;
    }

    public boolean setUser(String accountId) {
        User user = userRepository.findByAccountId(accountId);
        if (user == null) {
            return false;
        }
        user.setAccountId(accountId);
        // deal with false of connection
        return true;
    }

    public boolean checkUser(String accountId, String password) {
        User user = userRepository.findByAccountId(accountId);
        if (user == null) {
            return false;
        }
        return user.getPassword().equals(password);
    }

    public User getUser(String accountId) {
        return userRepository.findByAccountId(accountId);
    }

    public User getUserById(Long id) {
        return userRepository.findById(id).orElse(null);
    }
}