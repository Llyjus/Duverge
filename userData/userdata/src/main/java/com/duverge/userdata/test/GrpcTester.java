package com.duverge.userdata.test;

import com.example.userdata.grpc.UserDataServiceGrpc;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

public class GrpcTester {

    public static void main(String[] args) {

        ManagedChannel channel = ManagedChannelBuilder
                .forAddress("localhost", 9090)
                .usePlaintext()
                .build();

        UserDataServiceGrpc.UserDataServiceBlockingStub stub =
                UserDataServiceGrpc.newBlockingStub(channel);

        System.out.println("gRPC connected!");

        channel.shutdown();
    }
}