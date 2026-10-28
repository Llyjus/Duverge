package com.duverge.userdata.grpc;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.duverge.userdata.service.UserService;
import com.example.userdata.grpc.UserLoginRequest;
import com.example.userdata.grpc.UserLoginResponse;
import com.example.userdata.grpc.UserRegistrationRequest;
import com.example.userdata.grpc.UserRegistrationResponse;
import io.grpc.stub.StreamObserver;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.junit.jupiter.api.Test;

class UserServerMessageIdTests {

    @Test
    void echoesMessageIdForLoginAndRegistrationResponses() {
        var server = new UserServer(new SuccessfulUserService(), new SimpleMeterRegistry());
        var loginObserver = new RecordingObserver<UserLoginResponse>();
        var registrationObserver = new RecordingObserver<UserRegistrationResponse>();

        server.userLogin(UserLoginRequest.newBuilder()
                .setMessageId("login-request-1")
                .setAccountId("alice")
                .setPassword("test")
                .build(), loginObserver);
        server.userRegistration(UserRegistrationRequest.newBuilder()
                .setMessageId("register-request-1")
                .setAccountId("bob")
                .setPassword("test")
                .build(), registrationObserver);

        assertNull(loginObserver.error);
        assertTrue(loginObserver.completed);
        assertEquals("login-request-1", loginObserver.value.getMessageId());
        assertNull(registrationObserver.error);
        assertTrue(registrationObserver.completed);
        assertEquals("register-request-1", registrationObserver.value.getMessageId());
    }

    private static final class SuccessfulUserService extends UserService {
        private SuccessfulUserService() {
            super(null, null, null, null);
        }

        @Override
        public boolean checkUser(String accountId, String password) {
            return true;
        }

        @Override
        public boolean setNewUser(String accountId, String password) {
            return true;
        }
    }

    private static final class RecordingObserver<T> implements StreamObserver<T> {
        private T value;
        private Throwable error;
        private boolean completed;

        @Override
        public void onNext(T value) {
            this.value = value;
        }

        @Override
        public void onError(Throwable error) {
            this.error = error;
        }

        @Override
        public void onCompleted() {
            completed = true;
        }
    }
}
