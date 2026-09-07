package user_info;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@io.grpc.stub.annotations.GrpcGenerated
public final class UsersLoginServiceGrpc {

  private UsersLoginServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "user_info.UsersLoginService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<user_info.Login.UserRegistrationRequest,
      user_info.Login.UserRegistrationResponse> getUserRegistrationMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UserRegistration",
      requestType = user_info.Login.UserRegistrationRequest.class,
      responseType = user_info.Login.UserRegistrationResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<user_info.Login.UserRegistrationRequest,
      user_info.Login.UserRegistrationResponse> getUserRegistrationMethod() {
    io.grpc.MethodDescriptor<user_info.Login.UserRegistrationRequest, user_info.Login.UserRegistrationResponse> getUserRegistrationMethod;
    if ((getUserRegistrationMethod = UsersLoginServiceGrpc.getUserRegistrationMethod) == null) {
      synchronized (UsersLoginServiceGrpc.class) {
        if ((getUserRegistrationMethod = UsersLoginServiceGrpc.getUserRegistrationMethod) == null) {
          UsersLoginServiceGrpc.getUserRegistrationMethod = getUserRegistrationMethod =
              io.grpc.MethodDescriptor.<user_info.Login.UserRegistrationRequest, user_info.Login.UserRegistrationResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UserRegistration"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  user_info.Login.UserRegistrationRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  user_info.Login.UserRegistrationResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UsersLoginServiceMethodDescriptorSupplier("UserRegistration"))
              .build();
        }
      }
    }
    return getUserRegistrationMethod;
  }

  private static volatile io.grpc.MethodDescriptor<user_info.Login.UserLoginRequest,
      user_info.Login.UserLoginResponse> getUserLoginMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UserLogin",
      requestType = user_info.Login.UserLoginRequest.class,
      responseType = user_info.Login.UserLoginResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<user_info.Login.UserLoginRequest,
      user_info.Login.UserLoginResponse> getUserLoginMethod() {
    io.grpc.MethodDescriptor<user_info.Login.UserLoginRequest, user_info.Login.UserLoginResponse> getUserLoginMethod;
    if ((getUserLoginMethod = UsersLoginServiceGrpc.getUserLoginMethod) == null) {
      synchronized (UsersLoginServiceGrpc.class) {
        if ((getUserLoginMethod = UsersLoginServiceGrpc.getUserLoginMethod) == null) {
          UsersLoginServiceGrpc.getUserLoginMethod = getUserLoginMethod =
              io.grpc.MethodDescriptor.<user_info.Login.UserLoginRequest, user_info.Login.UserLoginResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UserLogin"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  user_info.Login.UserLoginRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  user_info.Login.UserLoginResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UsersLoginServiceMethodDescriptorSupplier("UserLogin"))
              .build();
        }
      }
    }
    return getUserLoginMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static UsersLoginServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceStub>() {
        @java.lang.Override
        public UsersLoginServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UsersLoginServiceStub(channel, callOptions);
        }
      };
    return UsersLoginServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports all types of calls on the service
   */
  public static UsersLoginServiceBlockingV2Stub newBlockingV2Stub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceBlockingV2Stub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceBlockingV2Stub>() {
        @java.lang.Override
        public UsersLoginServiceBlockingV2Stub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UsersLoginServiceBlockingV2Stub(channel, callOptions);
        }
      };
    return UsersLoginServiceBlockingV2Stub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static UsersLoginServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceBlockingStub>() {
        @java.lang.Override
        public UsersLoginServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UsersLoginServiceBlockingStub(channel, callOptions);
        }
      };
    return UsersLoginServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static UsersLoginServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UsersLoginServiceFutureStub>() {
        @java.lang.Override
        public UsersLoginServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UsersLoginServiceFutureStub(channel, callOptions);
        }
      };
    return UsersLoginServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void userRegistration(user_info.Login.UserRegistrationRequest request,
        io.grpc.stub.StreamObserver<user_info.Login.UserRegistrationResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUserRegistrationMethod(), responseObserver);
    }

    /**
     */
    default void userLogin(user_info.Login.UserLoginRequest request,
        io.grpc.stub.StreamObserver<user_info.Login.UserLoginResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUserLoginMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service UsersLoginService.
   */
  public static abstract class UsersLoginServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return UsersLoginServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service UsersLoginService.
   */
  public static final class UsersLoginServiceStub
      extends io.grpc.stub.AbstractAsyncStub<UsersLoginServiceStub> {
    private UsersLoginServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UsersLoginServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UsersLoginServiceStub(channel, callOptions);
    }

    /**
     */
    public void userRegistration(user_info.Login.UserRegistrationRequest request,
        io.grpc.stub.StreamObserver<user_info.Login.UserRegistrationResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUserRegistrationMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void userLogin(user_info.Login.UserLoginRequest request,
        io.grpc.stub.StreamObserver<user_info.Login.UserLoginResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUserLoginMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service UsersLoginService.
   */
  public static final class UsersLoginServiceBlockingV2Stub
      extends io.grpc.stub.AbstractBlockingStub<UsersLoginServiceBlockingV2Stub> {
    private UsersLoginServiceBlockingV2Stub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UsersLoginServiceBlockingV2Stub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UsersLoginServiceBlockingV2Stub(channel, callOptions);
    }

    /**
     */
    public user_info.Login.UserRegistrationResponse userRegistration(user_info.Login.UserRegistrationRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getUserRegistrationMethod(), getCallOptions(), request);
    }

    /**
     */
    public user_info.Login.UserLoginResponse userLogin(user_info.Login.UserLoginRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getUserLoginMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do limited synchronous rpc calls to service UsersLoginService.
   */
  public static final class UsersLoginServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<UsersLoginServiceBlockingStub> {
    private UsersLoginServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UsersLoginServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UsersLoginServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public user_info.Login.UserRegistrationResponse userRegistration(user_info.Login.UserRegistrationRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUserRegistrationMethod(), getCallOptions(), request);
    }

    /**
     */
    public user_info.Login.UserLoginResponse userLogin(user_info.Login.UserLoginRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUserLoginMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service UsersLoginService.
   */
  public static final class UsersLoginServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<UsersLoginServiceFutureStub> {
    private UsersLoginServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UsersLoginServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UsersLoginServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<user_info.Login.UserRegistrationResponse> userRegistration(
        user_info.Login.UserRegistrationRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUserRegistrationMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<user_info.Login.UserLoginResponse> userLogin(
        user_info.Login.UserLoginRequest request) {
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
          serviceImpl.userRegistration((user_info.Login.UserRegistrationRequest) request,
              (io.grpc.stub.StreamObserver<user_info.Login.UserRegistrationResponse>) responseObserver);
          break;
        case METHODID_USER_LOGIN:
          serviceImpl.userLogin((user_info.Login.UserLoginRequest) request,
              (io.grpc.stub.StreamObserver<user_info.Login.UserLoginResponse>) responseObserver);
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
              user_info.Login.UserRegistrationRequest,
              user_info.Login.UserRegistrationResponse>(
                service, METHODID_USER_REGISTRATION)))
        .addMethod(
          getUserLoginMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              user_info.Login.UserLoginRequest,
              user_info.Login.UserLoginResponse>(
                service, METHODID_USER_LOGIN)))
        .build();
  }

  private static abstract class UsersLoginServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UsersLoginServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return user_info.Login.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("UsersLoginService");
    }
  }

  private static final class UsersLoginServiceFileDescriptorSupplier
      extends UsersLoginServiceBaseDescriptorSupplier {
    UsersLoginServiceFileDescriptorSupplier() {}
  }

  private static final class UsersLoginServiceMethodDescriptorSupplier
      extends UsersLoginServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    UsersLoginServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (UsersLoginServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new UsersLoginServiceFileDescriptorSupplier())
              .addMethod(getUserRegistrationMethod())
              .addMethod(getUserLoginMethod())
              .build();
        }
      }
    }
    return result;
  }
}
