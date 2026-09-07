package com.duverge.userdata.grpc;

import org.springframework.grpc.server.service.GrpcService;

import com.example.userdata.grpc.UserDataServiceGrpc;

@GrpcService
public class GrpcServer extends UserDataServiceGrpc.UserDataServiceImplBase {
    
}