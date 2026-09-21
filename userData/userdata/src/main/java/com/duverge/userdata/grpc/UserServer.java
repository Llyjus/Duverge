package com.duverge.userdata.grpc;

import org.springframework.grpc.server.service.GrpcService;

import com.duverge.userdata.internal.RandomSession;
import com.duverge.userdata.service.UserService;
import com.example.userdata.grpc.UserDataServiceGrpc;
import com.example.userdata.grpc.UserLoginRequest;
import com.example.userdata.grpc.UserLoginResponse;
import com.example.userdata.grpc.UserRegistrationRequest;
import com.example.userdata.grpc.UserRegistrationResponse;

import io.grpc.stub.StreamObserver;

@GrpcService
public class UserServer extends UserDataServiceGrpc.UserDataServiceImplBase {
    private final UserService userService;

    public UserServer(UserService userService) {
        this.userService = userService;
    }

    @Override
    public void userLogin(
            UserLoginRequest request,
            StreamObserver<UserLoginResponse> responseObserver) {

        boolean userCheckResult = userService.checkUser(request.getAccountId(), request.getPassword());
        
        var sessionId = "";
        if (userCheckResult) {
            // If the user check is successful, create a session ID
            sessionId = RandomSession.generateSessionID();
            
        }

        UserLoginResponse response = UserLoginResponse.newBuilder()
                .setResult(userCheckResult)
                .setSessionId(sessionId)
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void userRegistration(
            UserRegistrationRequest request,
            StreamObserver<UserRegistrationResponse> responseObserver) {

        boolean userSetResult = userService.setNewUser(request.getAccountId(), request.getPassword());

        UserRegistrationResponse response = UserRegistrationResponse.newBuilder()
                .setResult(userSetResult)
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }
}