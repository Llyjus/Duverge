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

    public boolean setNewUser(String username, String password) {

        User user = new User();

        user.setAccountId(username);
        user.setPassword(password);

        userRepository.save(user);
        // deal with false of connection

        return true;
    }

    public boolean setUser(String username) {
        User user = userRepository.findByAccountId(username);
        if (user == null) {
            return false;
        }
        user.setAccountId(username);
        // deal with false of connection
        return true;
    }

    public boolean checkUser(String username, String password) {
        User user = userRepository.findByAccountId(username);
        if (user == null) {
            return false;
        }
        return user.getPassword().equals(password);
    }

    public User getUser(Long id) {
        return userRepository.findById(id).orElse(null);
    }
}