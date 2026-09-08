package com.example.userdata.grpc;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@io.grpc.stub.annotations.GrpcGenerated
public final class UserDataServiceGrpc {

  private UserDataServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "user_info.UserDataService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.example.userdata.grpc.UserRegistrationRequest,
      com.example.userdata.grpc.UserRegistrationResponse> getUserRegistrationMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UserRegistration",
      requestType = com.example.userdata.grpc.UserRegistrationRequest.class,
      responseType = com.example.userdata.grpc.UserRegistrationResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.example.userdata.grpc.UserRegistrationRequest,
      com.example.userdata.grpc.UserRegistrationResponse> getUserRegistrationMethod() {
    io.grpc.MethodDescriptor<com.example.userdata.grpc.UserRegistrationRequest, com.example.userdata.grpc.UserRegistrationResponse> getUserRegistrationMethod;
    if ((getUserRegistrationMethod = UserDataServiceGrpc.getUserRegistrationMethod) == null) {
      synchronized (UserDataServiceGrpc.class) {
        if ((getUserRegistrationMethod = UserDataServiceGrpc.getUserRegistrationMethod) == null) {
          UserDataServiceGrpc.getUserRegistrationMethod = getUserRegistrationMethod =
              io.grpc.MethodDescriptor.<com.example.userdata.grpc.UserRegistrationRequest, com.example.userdata.grpc.UserRegistrationResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UserRegistration"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.example.userdata.grpc.UserRegistrationRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.example.userdata.grpc.UserRegistrationResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserDataServiceMethodDescriptorSupplier("UserRegistration"))
              .build();
        }
      }
    }
    return getUserRegistrationMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.example.userdata.grpc.UserLoginRequest,
      com.example.userdata.grpc.UserLoginResponse> getUserLoginMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UserLogin",
      requestType = com.example.userdata.grpc.UserLoginRequest.class,
      responseType = com.example.userdata.grpc.UserLoginResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.example.userdata.grpc.UserLoginRequest,
      com.example.userdata.grpc.UserLoginResponse> getUserLoginMethod() {
    io.grpc.MethodDescriptor<com.example.userdata.grpc.UserLoginRequest, com.example.userdata.grpc.UserLoginResponse> getUserLoginMethod;
    if ((getUserLoginMethod = UserDataServiceGrpc.getUserLoginMethod) == null) {
      synchronized (UserDataServiceGrpc.class) {
        if ((getUserLoginMethod = UserDataServiceGrpc.getUserLoginMethod) == null) {
          UserDataServiceGrpc.getUserLoginMethod = getUserLoginMethod =
              io.grpc.MethodDescriptor.<com.example.userdata.grpc.UserLoginRequest, com.example.userdata.grpc.UserLoginResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UserLogin"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.example.userdata.grpc.UserLoginRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.example.userdata.grpc.UserLoginResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserDataServiceMethodDescriptorSupplier("UserLogin"))
              .build();
        }
      }
    }
    return getUserLoginMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static UserDataServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserDataServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserDataServiceStub>() {
        @java.lang.Override
        public UserDataServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserDataServiceStub(channel, callOptions);
        }
      };
    return UserDataServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports all types of calls on the service
   */
  public static UserDataServiceBlockingV2Stub newBlockingV2Stub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserDataServiceBlockingV2Stub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserDataServiceBlockingV2Stub>() {
        @java.lang.Override
        public UserDataServiceBlockingV2Stub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserDataServiceBlockingV2Stub(channel, callOptions);
        }
      };
    return UserDataServiceBlockingV2Stub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static UserDataServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserDataServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserDataServiceBlockingStub>() {
        @java.lang.Override
        public UserDataServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserDataServiceBlockingStub(channel, callOptions);
        }
      };
    return UserDataServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static UserDataServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserDataServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserDataServiceFutureStub>() {
        @java.lang.Override
        public UserDataServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserDataServiceFutureStub(channel, callOptions);
        }
      };
    return UserDataServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void userRegistration(com.example.userdata.grpc.UserRegistrationRequest request,
        io.grpc.stub.StreamObserver<com.example.userdata.grpc.UserRegistrationResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUserRegistrationMethod(), responseObserver);
    }

    /**
     */
    default void userLogin(com.example.userdata.grpc.UserLoginRequest request,
        io.grpc.stub.StreamObserver<com.example.userdata.grpc.UserLoginResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUserLoginMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service UserDataService.
   */
  public static abstract class UserDataServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return UserDataServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service UserDataService.
   */
  public static final class UserDataServiceStub
      extends io.grpc.stub.AbstractAsyncStub<UserDataServiceStub> {
    private UserDataServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserDataServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserDataServiceStub(channel, callOptions);
    }

    /**
     */
    public void userRegistration(com.example.userdata.grpc.UserRegistrationRequest request,
        io.grpc.stub.StreamObserver<com.example.userdata.grpc.UserRegistrationResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUserRegistrationMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void userLogin(com.example.userdata.grpc.UserLoginRequest request,
        io.grpc.stub.StreamObserver<com.example.userdata.grpc.UserLoginResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUserLoginMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service UserDataService.
   */
  public static final class UserDataServiceBlockingV2Stub
      extends io.grpc.stub.AbstractBlockingStub<UserDataServiceBlockingV2Stub> {
    private UserDataServiceBlockingV2Stub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserDataServiceBlockingV2Stub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserDataServiceBlockingV2Stub(channel, callOptions);
    }

    /**
     */
    public com.example.userdata.grpc.UserRegistrationResponse userRegistration(com.example.userdata.grpc.UserRegistrationRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getUserRegistrationMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.example.userdata.grpc.UserLoginResponse userLogin(com.example.userdata.grpc.UserLoginRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getUserLoginMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do limited synchronous rpc calls to service UserDataService.
   */
  public static final class UserDataServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<UserDataServiceBlockingStub> {
    private UserDataServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserDataServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserDataServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public com.example.userdata.grpc.UserRegistrationResponse userRegistration(com.example.userdata.grpc.UserRegistrationRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUserRegistrationMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.example.userdata.grpc.UserLoginResponse userLogin(com.example.userdata.grpc.UserLoginRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUserLoginMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service UserDataService.
   */
  public static final class UserDataServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<UserDataServiceFutureStub> {
    private UserDataServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserDataServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserDataServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.example.userdata.grpc.UserRegistrationResponse> userRegistration(
        com.example.userdata.grpc.UserRegistrationRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUserRegistrationMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.example.userdata.grpc.UserLoginResponse> userLogin(
        com.example.userdata.grpc.UserLoginRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUserLoginMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_USER_REGISTRATION = 0;
  private static final int METHODID_USER_LOGIN = 1;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_USER_REGISTRATION:
          serviceImpl.userRegistration((com.example.userdata.grpc.UserRegistrationRequest) request,
              (io.grpc.stub.StreamObserver<com.example.userdata.grpc.UserRegistrationResponse>) responseObserver);
          break;
        case METHODID_USER_LOGIN:
          serviceImpl.userLogin((com.example.userdata.grpc.UserLoginRequest) request,
              (io.grpc.stub.StreamObserver<com.example.userdata.grpc.UserLoginResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getUserRegistrationMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.example.userdata.grpc.UserRegistrationRequest,
              com.example.userdata.grpc.UserRegistrationResponse>(
                service, METHODID_USER_REGISTRATION)))
        .addMethod(
          getUserLoginMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.example.userdata.grpc.UserLoginRequest,
              com.example.userdata.grpc.UserLoginResponse>(
                service, METHODID_USER_LOGIN)))
        .build();
  }

  private static abstract class UserDataServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserDataServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.example.userdata.grpc.Login.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("UserDataService");
    }
  }

  private static final class UserDataServiceFileDescriptorSupplier
      extends UserDataServiceBaseDescriptorSupplier {
    UserDataServiceFileDescriptorSupplier() {}
  }

  private static final class UserDataServiceMethodDescriptorSupplier
      extends UserDataServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    UserDataServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (UserDataServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new UserDataServiceFileDescriptorSupplier())
              .addMethod(getUserRegistrationMethod())
              .addMethod(getUserLoginMethod())
              .build();
        }
      }
    }
    return result;
  }
}
