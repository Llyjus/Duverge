package com.duverge.userdata.grpc;

import org.springframework.grpc.server.service.GrpcService;

import com.duverge.userdata.error.PasswordHashUnavailableException;
import com.duverge.userdata.error.DatabaseCapacityUnavailableException;
import com.duverge.userdata.internal.RandomSession;
import com.duverge.userdata.service.UserService;
import com.example.userdata.grpc.UserDataServiceGrpc;
import com.example.userdata.grpc.UserLoginRequest;
import com.example.userdata.grpc.UserLoginResponse;
import com.example.userdata.grpc.UserRegistrationRequest;
import com.example.userdata.grpc.UserRegistrationResponse;

import io.grpc.stub.StreamObserver;
import io.grpc.Status;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;

@GrpcService
public class UserServer extends UserDataServiceGrpc.UserDataServiceImplBase {
    private final UserService userService;
    private final MeterRegistry meterRegistry;

    public UserServer(UserService userService, MeterRegistry meterRegistry) {
        this.userService = userService;
        this.meterRegistry = meterRegistry;
    }

    @Override
    public void userLogin(
            UserLoginRequest request,
            StreamObserver<UserLoginResponse> responseObserver) {

        boolean userCheckResult;
        try {
            userCheckResult = recordLatency("userLogin", () -> userService.checkUser(
                    request.getAccountId(), request.getPassword()));
        } catch (PasswordHashUnavailableException e) {
            responseObserver.onError(Status.RESOURCE_EXHAUSTED
                    .withDescription("Password hashing capacity is temporarily exhausted")
                    .asRuntimeException());
            return;
        } catch (DatabaseCapacityUnavailableException e) {
            responseObserver.onError(Status.RESOURCE_EXHAUSTED
                    .withDescription("Database capacity is temporarily exhausted")
                    .asRuntimeException());
            return;
        }
        
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

        boolean userSetResult;
        try {
            userSetResult = recordLatency("userRegistration", () -> userService.setNewUser(
                    request.getAccountId(), request.getPassword()));
        } catch (PasswordHashUnavailableException e) {
            responseObserver.onError(Status.RESOURCE_EXHAUSTED
                    .withDescription("Password hashing capacity is temporarily exhausted")
                    .asRuntimeException());
            return;
        } catch (DatabaseCapacityUnavailableException e) {
            responseObserver.onError(Status.RESOURCE_EXHAUSTED
                    .withDescription("Database capacity is temporarily exhausted")
                    .asRuntimeException());
            return;
        }

        UserRegistrationResponse response = UserRegistrationResponse.newBuilder()
                .setResult(userSetResult)
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    private boolean recordLatency(String method, java.util.function.BooleanSupplier operation) {
        Timer timer = Timer.builder("userdata.grpc.server.requests")
                .description("UserData gRPC request duration")
                .tag("method", method)
                .register(meterRegistry);
        return timer.record(operation::getAsBoolean);
    }
}
