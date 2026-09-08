package com.duverge.userdata.test;

import com.example.userdata.grpc.UserDataServiceGrpc;
import com.example.userdata.grpc.UserRegistrationRequest;
import com.example.userdata.grpc.UserRegistrationResponse;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

public class WriteInTester {

    public static void main(String[] args) {

        ManagedChannel channel = ManagedChannelBuilder
                .forAddress("localhost", 9090)
                .usePlaintext()
                .build();

        UserDataServiceGrpc.UserDataServiceBlockingStub stub =
                UserDataServiceGrpc.newBlockingStub(channel);

        UserRegistrationRequest request =
                UserRegistrationRequest.newBuilder()
                        .setAccountId("lin")
                        .setPassword("123")
                        .build();

        UserRegistrationResponse response =
                stub.userRegistration(request);

        System.out.println("result = " + response.getResult());

        channel.shutdown();
    }
}