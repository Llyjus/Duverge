package com.duverge.userdata.grpc;

import org.springframework.grpc.server.service.GrpcService;

import com.duverge.userdata.service.UserService;
import com.example.userdata.grpc.UserDataServiceGrpc;
import com.example.userdata.grpc.UserLoginRequest;
import com.example.userdata.grpc.UserLoginResponse;

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
            

        UserLoginResponse response = UserLoginResponse.newBuilder()
                .setResult(userCheckResult)
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void userRegistration(
            com.example.userdata.grpc.UserRegistrationRequest request,
            StreamObserver<com.example.userdata.grpc.UserRegistrationResponse> responseObserver) {

        boolean userSetResult = userService.setNewUser(request.getAccountId(), request.getPassword());

        com.example.userdata.grpc.UserRegistrationResponse response = com.example.userdata.grpc.UserRegistrationResponse.newBuilder()
                .setResult(userSetResult)
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }
}