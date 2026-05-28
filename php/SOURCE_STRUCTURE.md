# Complete Valkyrja PHP Source Directory Structure

**Base Path:** `/Users/melechmizrachi/Dropbox/Sites/Valkyrja/php/valkyrja/src/Valkyrja`

## Statistics

- **Total PHP Files:** 1,140
- **Total Directories:** 619
- **Major Modules:** 27

---

## Directory Tree by Module

### 1. API Module

*13 files, 12 directories*

```
├── Constant/
│   └── Status.php
├── Manager/
│   ├── Api.php
│   └── Contract/
│       └── ApiContract.php
├── Middleware/
│   └── ApiThrowableCaughtMiddleware.php
├── Model/
│   ├── Json.php
│   ├── JsonData.php
│   └── Contract/
│       ├── JsonContract.php
│       └── JsonDataContract.php
├── Provider/
│   ├── ApiComponentProvider.php
│   └── ApiServiceProvider.php
└── Throwable/
    ├── Contract/
    │   └── ApiThrowable.php
    └── Exception/
        └── Abstract/
            ├── ApiInvalidArgumentException.php
            └── ApiRuntimeException.php
```

### 2. Application Module

*26 files, 16 directories*

```
├── Constant/
│   └── ApplicationInfo.php
├── Data/
│   ├── CliConfig.php
│   ├── Config.php
│   ├── HttpConfig.php
│   └── Contract/
│       ├── CliConfigContract.php
│       ├── ConfigContract.php
│       └── HttpConfigContract.php
├── Directory/
│   └── Directory.php
├── Entry/
│   ├── Cli.php
│   ├── Http.php
│   └── Abstract/
│       ├── App.php
│       └── WorkerHttp.php
├── Env/
│   └── Env.php
├── Kernel/
│   ├── Valkyrja.php
│   ├── ChildApplication.php
│   └── Contract/
│       └── ApplicationContract.php
├── Provider/
│   ├── ApplicationComponentProvider.php
│   ├── CliApplicationComponentProvider.php
│   ├── CliWithHttpApplicationComponentProvider.php
│   ├── HttpApplicationComponentProvider.php
│   └── Contract/
│       ├── ComponentProviderContract.php
│       └── PublishableComponentProviderContract.php
├── Throwable/
│   ├── Contract/
│   │   └── ApplicationThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── ApplicationInvalidArgumentException.php
│           └── ApplicationRuntimeException.php
├── README.md
├── APPLICATION_STRUCTURE.md
├── GETTING_STARTED.md
└── LIFECYCLE.md
```

### 3. Attribute Module

*10 files, 10 directories*

```
├── Collector/
│   ├── Collector.php
│   └── Contract/
│       └── CollectorContract.php
├── Contract/
│   └── ReflectionAwareAttributeContract.php
├── Provider/
│   ├── AttributeComponentProvider.php
│   └── AttributeServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── AttributeThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── AttributeInvalidArgumentException.php
│           └── AttributeRuntimeException.php
├── Trait/
│   └── ReflectionAwareAttribute.php
└── README.md
```

### 4. Auth Module

*72 files, 23 directories*

```
├── Authenticator/
│   └── SessionAuthenticator.php
├── Constant/
│   ├── RouteName.php
│   ├── SessionItemId.php
│   └── UserField.php
├── Data/
│   └── AuthenticatedUsers.php
├── Entity/
│   ├── User.php
│   ├── VerifiableUser.php
│   ├── LockableUser.php
│   ├── MailableUser.php
│   ├── Contract/
│   │   ├── UserContract.php
│   │   ├── VerifiableUserContract.php
│   │   ├── LockableUserContract.php
│   │   ├── MailableUserContract.php
│   │   ├── TwoFactorUserContract.php
│   │   ├── PinUserContract.php
│   │   ├── DeviceAuthenticatedUserContract.php
│   │   ├── AntiPhishCodeUserContract.php
│   │   ├── LastOnlineUserContract.php
│   │   ├── PermissibleUserContract.php
│   │   └── UserDeviceContract.php
│   └── Trait/
│       ├── UserFields.php
│       ├── UserMethods.php
│       ├── VerifiableUserFields.php
│       ├── VerifiableUserMethods.php
│       ├── LockableUserFields.php
│       └── LockableUserMethods.php
├── Hasher/
│   ├── PhpPasswordHasher.php
│   └── Contract/
│       └── PasswordHasherContract.php
├── Provider/
│   ├── AuthComponentProvider.php
│   └── AuthServiceProvider.php
├── Store/
│   ├── InMemoryStore.php
│   ├── NullStore.php
│   └── OrmStore.php
├── Throwable/
│   ├── Contract/
│   │   └── AuthThrowable.php
│   └── Exception/
│       ├── AuthInvalidAuthenticatedUsersSessionValueException.php
│       ├── AuthInvalidAuthenticationException.php
│       ├── AuthInvalidCurrentAuthenticationException.php
│       ├── AuthInvalidPasswordConfirmationException.php
│       ├── AuthInvalidRegistrationException.php
│       ├── AuthInvalidRetrievableUserException.php
│       ├── AuthInvalidUnserializedAuthenticatedUsersException.php
│       ├── AuthMissingTokenizableUserRequiredFieldsException.php
│       ├── AuthNoCurrentUserException.php
│       ├── AuthNoImpersonatedUserException.php
│       ├── AuthTokenizationException.php
│       ├── AuthUnexpectedPasswordValueException.php
│       ├── AuthUnexpectedUsernameValueException.php
│       └── [Additional exception classes...]
├── Attempt/
├── Retrieval/
├── README.md
└── [Authenticator-related files]
```

### 5. Broadcast Module

*13 files, 10 directories*

```
├── Broadcaster/
│   ├── PusherBroadcaster.php
│   ├── CryptPusherBroadcaster.php
│   ├── LogBroadcaster.php
│   ├── NullBroadcaster.php
│   └── Contract/
│       └── BroadcasterContract.php
├── Data/
│   ├── Message.php
│   └── Contract/
│       └── MessageContract.php
├── Provider/
│   ├── BroadcastComponentProvider.php
│   └── BroadcastServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── BroadcastThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── BroadcastInvalidArgumentException.php
│           └── BroadcastRuntimeException.php
└── README.md
```

### 6. Cache Module

*12 files, 10 directories*

```
├── Manager/
│   ├── RedisCache.php
│   ├── LogCache.php
│   ├── NullCache.php
│   └── Contract/
│       └── CacheContract.php
├── Tagger/
│   ├── Tagger.php
│   └── Contract/
│       └── TaggerContract.php
├── Provider/
│   ├── CacheComponentProvider.php
│   └── CacheServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── CacheThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── CacheInvalidArgumentException.php
│           └── CacheRuntimeException.php
└── README.md
```

### 7. CLI Module

*172 files, 85 directories — **Large Module***

```
├── Interaction/
│   ├── Argument/
│   │   └── Argument.php
│   ├── Data/
│   │   ├── CliInteractionConfig.php
│   │   └── Contract/
│   │       └── CliInteractionConfigContract.php
│   ├── Enum/
│   │   ├── BackgroundColor.php
│   │   ├── ExitCode.php
│   │   ├── OptionType.php
│   │   ├── Style.php
│   │   └── TextColor.php
│   ├── Format/
│   │   ├── Format.php
│   │   ├── BackgroundColorFormat.php
│   │   ├── StyleFormat.php
│   │   ├── TextColorFormat.php
│   │   └── Contract/
│   │       └── FormatContract.php
│   ├── Formatter/
│   │   ├── Formatter.php
│   │   ├── ErrorFormatter.php
│   │   ├── SuccessFormatter.php
│   │   ├── WarningFormatter.php
│   │   ├── HighlightedTextFormatter.php
│   │   └── QuestionFormatter.php
│   ├── Input/
│   │   └── Input.php
│   ├── Message/
│   │   ├── Message.php
│   │   ├── Messages.php
│   │   ├── Banner.php
│   │   ├── Answer.php
│   │   ├── ErrorMessage.php
│   │   ├── SuccessMessage.php
│   │   ├── WarningMessage.php
│   │   ├── Question.php
│   │   ├── NewLine.php
│   │   └── Progress.php
│   ├── Option/
│   │   └── Option.php
│   ├── Output/
│   │   ├── Output.php
│   │   ├── StreamOutput.php
│   │   ├── FileOutput.php
│   │   ├── PlainOutput.php
│   │   └── EmptyOutput.php
│   ├── Writer/
│   │   └── QuestionWriter.php
│   ├── Provider/
│   │   ├── CliInteractionComponentProvider.php
│   │   └── CliInteractionServiceProvider.php
│   └── [Additional interaction files]
├── Middleware/
│   ├── Handler/
│   │   ├── ExitedHandler.php
│   │   ├── InputReceivedHandler.php
│   │   ├── RouteDispatchedHandler.php
│   │   ├── RouteMatchedHandler.php
│   │   ├── RouteNotMatchedHandler.php
│   │   └── ThrowableCaughtHandler.php
│   └── Contract/
│       ├── ExitedMiddlewareContract.php
│       ├── InputReceivedMiddlewareContract.php
│       ├── RouteNotMatchedMiddlewareContract.php
│       └── ThrowableCaughtMiddlewareContract.php
├── [Command, Dispatcher, Collector, Controller, etc.]
└── README.md
```

### 8. Container Module

*17 files, 12 directories*

```
├── Manager/
│   ├── Container.php
│   ├── ChildContainer.php
│   ├── NativeChildContainer.php
│   └── Contract/
│       ├── ContainerContract.php
│       └── ProvidersAwareContract.php
├── Manager/Trait/
│   └── ProvidersAware.php
├── Data/
│   └── ContainerData.php
├── Enum/
│   └── InvalidReferenceMode.php
├── Provider/
│   ├── ContainerComponentProvider.php
│   ├── ContainerServiceProvider.php
│   └── Contract/
│       └── ServiceProviderContract.php
├── Throwable/
│   ├── Contract/
│   │   └── ContainerThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── ContainerInvalidArgumentException.php
│       │   └── ContainerRuntimeException.php
│       ├── ContainerInvalidPublishCallbackException.php
│       └── ContainerInvalidReferenceException.php
└── README.md
```

### 9. Crypt Module

*14 files, 8 directories*

```
├── Manager/
│   ├── SodiumCrypt.php
│   ├── NullCrypt.php
│   └── Contract/
│       └── CryptContract.php
├── Provider/
│   ├── CryptComponentProvider.php
│   └── CryptServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── CryptThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── CryptInvalidArgumentException.php
│       │   └── CryptRuntimeException.php
│       ├── CryptDecodeFailureException.php
│       ├── CryptEncryptionFailureException.php
│       ├── CryptKeyToBytesException.php
│       ├── CryptTamperedMessageException.php
│       └── CryptTruncatedMessageException.php
└── README.md
```

### 10. Dispatch Module

*35 files, 12 directories*

```
├── Data/
│   ├── CallableDispatch.php
│   ├── ClassDispatch.php
│   ├── ConstantDispatch.php
│   ├── GlobalVariableDispatch.php
│   ├── MethodDispatch.php
│   ├── PropertyDispatch.php
│   └── Contract/
│       ├── CallableDispatchContract.php
│       ├── ClassDispatchContract.php
│       ├── ConstantDispatchContract.php
│       ├── DispatchContract.php
│       ├── GlobalVariableDispatchContract.php
│       ├── MethodDispatchContract.php
│       └── PropertyDispatchContract.php
├── Dispatcher/
│   ├── Dispatcher.php
│   └── Contract/
│       └── DispatcherContract.php
├── Factory/
│   └── DispatchFactory.php
├── Provider/
│   ├── DispatchComponentProvider.php
│   └── DispatchServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── DispatchThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── DispatchInvalidArgumentException.php
│       │   └── DispatchRuntimeException.php
│       ├── DispatchCallableMissingClassNameException.php
│       ├── DispatchCallableMissingMethodNameException.php
│       ├── DispatchCallableNonStringClassNameException.php
│       ├── DispatchInvalidClosureException.php
│       ├── DispatchInvalidDispatchCapabilityException.php
│       ├── DispatchInvalidFunctionException.php
│       ├── DispatchInvalidMethodException.php
│       ├── DispatchInvalidPropertyException.php
│       ├── DispatchInvalidReflectionFunctionException.php
│       ├── DispatchNoClassException.php
│       ├── DispatchUnsupportedCallableException.php
│       └── DispatchUnsupportedDispatchException.php
└── README.md
```

### 11. Event Module

*20 files, 17 directories*

```
├── Attribute/
│   ├── Listener.php
│   └── ListenerHandler.php
├── Collection/
│   ├── ListenerCollection.php
│   └── Contract/
│       └── ListenerCollectionContract.php
├── Collector/
│   ├── AttributeListenerCollector.php
│   └── Contract/
│       └── ListenerCollectorContract.php
├── Contract/
│   ├── ArgumentsCapableEventContract.php
│   ├── DispatchCollectableEventContract.php
│   └── (Additional contracts)
├── Data/
│   ├── EventData.php
│   ├── Listener.php
│   └── Contract/
│       └── ListenerContract.php
├── Dispatcher/
│   ├── EventDispatcher.php
│   └── Contract/
│       └── EventDispatcherContract.php
├── Provider/
│   ├── EventComponentProvider.php
│   ├── EventServiceProvider.php
│   └── Contract/
│       └── ListenerProviderContract.php
├── Throwable/
│   ├── Contract/
│   │   └── EventThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── EventInvalidArgumentException.php
│           └── EventRuntimeException.php
└── README.md
```

### 12. Filesystem Module

*17 files, 10 directories*

```
├── Manager/
│   ├── FlysystemFilesystem.php
│   ├── LocalFlysystemFilesystem.php
│   ├── S3FlysystemFilesystem.php
│   ├── InMemoryFilesystem.php
│   ├── NullFilesystem.php
│   └── Contract/
│       └── FilesystemContract.php
├── Data/
│   ├── InMemoryFile.php
│   └── InMemoryMetadata.php
├── Enum/
│   └── Visibility.php
├── Provider/
│   ├── FilesystemComponentProvider.php
│   └── FilesystemServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── FilesystemThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── FilesystemInvalidArgumentException.php
│       │   └── FilesystemRuntimeException.php
│       ├── FilesystemResourceReadException.php
│       └── FilesystemUnableToReadContentsException.php
└── README.md
```

### 13. HTTP Module

*297 files, 158 directories — **Largest Module***

```
├── Client/
│   ├── Manager/
│   │   ├── GuzzleClient.php
│   │   ├── LogClient.php
│   │   ├── NullClient.php
│   │   └── Contract/
│   │       └── ClientContract.php
│   ├── Provider/
│   │   ├── HttpClientComponentProvider.php
│   │   └── HttpClientServiceProvider.php
│   └── Throwable/
│       └── Contract/
│           └── HttpClientThrowable.php
├── Middleware/
│   ├── Handler/
│   │   ├── Abstract/
│   │   │   └── Handler.php
│   │   ├── RequestReceivedHandler.php
│   │   ├── RouteMatchedHandler.php
│   │   ├── RouteDispatchedHandler.php
│   │   ├── RouteNotMatchedHandler.php
│   │   ├── SendingResponseHandler.php
│   │   ├── TerminatedHandler.php
│   │   ├── ThrowableCaughtHandler.php
│   │   └── Contract/
│   │       ├── HandlerContract.php
│   │       ├── RequestReceivedHandlerContract.php
│   │       ├── RouteDispatchedHandlerContract.php
│   │       ├── RouteMatchedHandlerContract.php
│   │       ├── RouteNotMatchedHandlerContract.php
│   │       ├── SendingResponseHandlerContract.php
│   │       ├── TerminatedHandlerContract.php
│   │       └── ThrowableCaughtHandlerContract.php
│   ├── Provider/
│   │   ├── HttpMiddlewareComponentProvider.php
│   │   └── HttpMiddlewareServiceProvider.php
│   ├── Contract/
│   │   ├── RequestReceivedMiddlewareContract.php
│   │   ├── RouteDispatchedMiddlewareContract.php
│   │   ├── RouteMatchedMiddlewareContract.php
│   │   ├── RouteNotMatchedMiddlewareContract.php
│   │   ├── SendingResponseMiddlewareContract.php
│   │   ├── TerminatedMiddlewareContract.php
│   │   └── ThrowableCaughtMiddlewareContract.php
│   └── Throwable/
│       └── Contract/
│           └── HttpMiddlewareThrowable.php
├── Routing/
│   ├── Controller/
│   │   ├── Controller.php
│   │   └── ApiController.php
│   ├── Constant/
│   │   └── Regex.php
│   ├── Data/
│   │   ├── Route.php
│   │   ├── DynamicRoute.php
│   │   ├── Parameter.php
│   │   ├── HttpRoutingData.php
│   │   └── Contract/
│   │       ├── RouteContract.php
│   │       ├── DynamicRouteContract.php
│   │       └── ParameterContract.php
│   ├── Factory/
│   │   ├── RoutingResponseFactory.php
│   │   └── Contract/
│   │       └── RoutingResponseFactoryContract.php
│   ├── Url/
│   │   ├── Url.php
│   │   └── Contract/
│   │       └── UrlContract.php
│   └── [Additional routing files]
├── Throwable/
│   ├── Contract/
│   │   └── HttpThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── HttpInvalidArgumentException.php
│           └── [Additional HTTP exception files...]
├── Request/
├── Response/
├── Attribute/
├── [Many more HTTP-related modules...]
└── README.md
```

### 14. JWT Module

*10 files, 9 directories*

```
├── Enum/
│   └── Algorithm.php
├── Manager/
│   ├── FirebaseJwt.php
│   ├── NullJwt.php
│   └── Contract/
│       └── JwtContract.php
├── Provider/
│   ├── JwtComponentProvider.php
│   └── JwtServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── JwtThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── JwtInvalidArgumentException.php
│           └── JwtRuntimeException.php
└── README.md
```

### 15. Log Module

*12 files, 10 directories*

```
├── Enum/
│   └── LogLevel.php
├── Logger/
│   ├── Abstract/
│   │   └── Logger.php
│   ├── PsrLogger.php
│   ├── NullLogger.php
│   └── Contract/
│       └── LoggerContract.php
├── Provider/
│   ├── LogComponentProvider.php
│   └── LogServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── LogThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── LogInvalidArgumentException.php
│       │   └── LogRuntimeException.php
│       └── LogInvalidLogLevelException.php
└── README.md
```

### 16. Mail Module

*17 files, 10 directories*

```
├── Data/
│   ├── Message.php
│   ├── Recipient.php
│   ├── Attachment.php
│   └── Contract/
│       ├── MessageContract.php
│       ├── RecipientContract.php
│       └── AttachmentContract.php
├── Mailer/
│   ├── PhpMailer.php
│   ├── MailgunMailer.php
│   ├── LogMailer.php
│   ├── NullMailer.php
│   └── Contract/
│       └── MailerContract.php
├── Provider/
│   ├── MailComponentProvider.php
│   └── MailServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── MailThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── MailInvalidArgumentException.php
│           └── MailRuntimeException.php
└── README.md
```

### 17. ORM Module

*90 files, 32 directories — **Complex Module***

```
├── Constant/
│   ├── DateFormat.php
│   └── Statement.php
├── Data/
│   ├── Id.php
│   ├── Value.php
│   ├── Where.php
│   ├── WhereGroup.php
│   ├── OrderBy.php
│   ├── Join.php
│   ├── EntityCast.php
│   ├── Where/
│   │   ├── AndWhere.php
│   │   ├── AndNotWhere.php
│   │   └── OrWhere.php
│   └── Join/
│       ├── InnerJoin.php
│       ├── LeftJoin.php
│       ├── RightJoin.php
│       ├── OuterJoin.php
│       └── FullOuterJoin.php
├── Enum/
│   ├── Comparison.php
│   ├── JoinType.php
│   ├── JoinOperator.php
│   ├── SortOrder.php
│   └── WhereType.php
├── Factory/
│   ├── DateFactory.php
│   └── [Other factories...]
├── Manager/
│   ├── Abstract/
│   │   └── PdoManager.php
│   ├── MysqlManager.php
│   ├── PgsqlManager.php
│   ├── SqliteManager.php
│   ├── NullManager.php
│   └── Contract/
│       └── ManagerContract.php
├── Middleware/
│   └── EntityRouteMatchedMiddleware.php
├── Provider/
│   ├── OrmComponentProvider.php
│   └── OrmServiceProvider.php
├── QueryBuilder/
│   ├── SqlSelectQueryBuilder.php
│   ├── SqlInsertQueryBuilder.php
│   ├── SqlUpdateQueryBuilder.php
│   └── SqlDeleteQueryBuilder.php
├── Repository/
│   └── Repository.php
├── Schema/
│   ├── Abstract/
│   │   ├── Migration.php
│   │   ├── SqlFileMigration.php
│   │   └── TransactionalMigration.php
│   ├── Contract/
│   │   ├── MigrationContract.php
│   │   ├── SchemaContract.php
│   │   ├── TableContract.php
│   │   ├── ColumnContract.php
│   │   ├── ConstraintContract.php
│   │   └── IndexContract.php
│   └── [Schema implementation files...]
├── Statement/
│   ├── PdoStatement.php
│   ├── NullStatement.php
│   └── Contract/
│       └── StatementContract.php
├── README.md
└── [Additional ORM files...]
```

### 18. Reflection Module

*9 files, 8 directories*

```
├── Provider/
│   ├── ReflectionComponentProvider.php
│   └── ReflectionServiceProvider.php
├── Reflector/
│   ├── Reflector.php
│   └── Contract/
│       └── ReflectorContract.php
├── Throwable/
│   ├── Contract/
│   │   └── ReflectionThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── ReflectionInvalidArgumentException.php
│       │   └── ReflectionRuntimeException.php
│       └── ReflectionInvalidClassConstantException.php
└── README.md
```

### 19. Session Module

*28 files, 17 directories*

```
├── Manager/
│   ├── Abstract/
│   │   └── Session.php
│   ├── PhpSession.php
│   ├── CacheSession.php
│   ├── LogSession.php
│   ├── NullSession.php
│   ├── Cookie/
│   │   ├── CookieSession.php
│   │   └── EncryptedCookieSession.php
│   ├── Jwt/
│   │   ├── Http/
│   │   │   ├── HeaderJwtSession.php
│   │   │   └── EncryptedHeaderJwtSession.php
│   │   └── Cli/
│   │       ├── OptionJwtSession.php
│   │       └── EncryptedOptionJwtSession.php
│   ├── Token/
│   │   ├── Http/
│   │   │   ├── HeaderTokenSession.php
│   │   │   └── EncryptedHeaderTokenSession.php
│   │   └── Cli/
│   │       ├── OptionTokenSession.php
│   │       └── EncryptedOptionTokenSession.php
│   └── Contract/
│       └── SessionContract.php
├── Data/
│   └── CookieParams.php
├── Provider/
│   ├── SessionComponentProvider.php
│   └── [Additional provider files...]
├── Throwable/
│   ├── Contract/
│   │   └── SessionThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── SessionInvalidArgumentException.php
│       │   └── SessionRuntimeException.php
│       ├── SessionIdFailureException.php
│       ├── SessionInvalidCsrfTokenException.php
│       ├── SessionInvalidSessionIdException.php
│       ├── SessionNameFailureException.php
│       └── SessionStartFailureException.php
└── README.md
```

### 20. SMS Module

*12 files, 10 directories*

```
├── Data/
│   ├── Message.php
│   └── Contract/
│       └── MessageContract.php
├── Messenger/
│   ├── VonageMessenger.php
│   ├── LogMessenger.php
│   ├── NullMessenger.php
│   └── Contract/
│       └── MessengerContract.php
├── Provider/
│   ├── SmsComponentProvider.php
│   └── SmsServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── SmsThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── SmsInvalidArgumentException.php
│           └── SmsRuntimeException.php
└── README.md
```

### 21. Support Module

*6 files, 6 directories*

```
├── Generator/
│   ├── Abstract/
│   │   └── FileGenerator.php
│   ├── Enum/
│   │   └── GenerateStatus.php
│   └── Contract/
│       └── FileGeneratorContract.php
├── Time/
│   ├── Time.php
│   └── Microtime.php
└── README.md
```

### 22. Throwable Module

*7 files, 7 directories*

```
├── Contract/
│   └── ValkyrjaThrowable.php
├── Exception/
│   └── Abstract/
│       ├── ValkyrjaInvalidArgumentException.php
│       └── ValkyrjaRuntimeException.php
├── Handler/
│   ├── Abstract/
│   │   └── ThrowableHandler.php
│   ├── WhoopsThrowableHandler.php
│   └── Contract/
│       └── ThrowableHandlerContract.php
└── README.md
```

### 23. Type Module

*168 files, 83 directories — **Large Type System***

```
├── Abstract/
│   └── Type.php
├── Array/
│   ├── ArrayT.php
│   └── NonEmptyArray.php
├── Bool/
│   ├── BoolT.php
│   ├── TrueT.php
│   └── FalseT.php
├── Collection/
│   └── Collection.php
├── Data/
│   ├── Cast.php
│   ├── ArrayCast.php
│   ├── OriginalCast.php
│   └── OriginalArrayCast.php
├── Enum/
│   ├── Type.php
│   └── CastType.php
├── Float/
│   └── FloatT.php
├── Id/
│   ├── Id.php
│   ├── IntId.php
│   └── StringId.php
├── Int/
│   └── IntT.php
├── Json/
│   ├── Json.php
│   └── JsonObject.php
├── Null/
│   └── NullT.php
├── Object/
│   ├── ObjectT.php
│   └── SerializedObject.php
├── String/
│   ├── StringT.php
│   └── NonEmptyString.php
├── Uid/
│   └── Uid.php
├── Ulid/
│   ├── Ulid.php
│   └── README.md
├── Uuid/
│   ├── Uuid.php
│   ├── UuidV1.php
│   ├── UuidV3.php
│   ├── UuidV4.php
│   ├── UuidV5.php
│   ├── UuidV6.php
│   ├── UuidV7.php
│   ├── UuidV8.php
│   └── README.md
├── Vlid/
│   ├── Vlid.php
│   ├── VlidV1.php
│   ├── VlidV2.php
│   ├── VlidV3.php
│   ├── VlidV4.php
│   ├── Contract/
│   │   ├── VlidContract.php
│   │   ├── VlidV1Contract.php
│   │   ├── VlidV2Contract.php
│   │   └── VlidV4Contract.php
│   └── README.md
├── Contract/
│   ├── TypeContract.php
│   └── [Multiple type-specific contracts...]
└── [Additional type system files...]
```

### 24. Validation Module

*33 files, 16 directories*

```
├── Constant/
│   └── ErrorMessage.php
├── Rule/
│   ├── Abstract/
│   │   └── Rule.php
│   ├── Contract/
│   │   └── RuleContract.php
│   ├── Is/
│   │   ├── Required.php
│   │   ├── NotEmpty.php
│   │   ├── IsEmpty.php
│   │   ├── Email.php
│   │   ├── Equal.php
│   │   ├── NotEqual.php
│   │   ├── IsBool.php
│   │   ├── IsString.php
│   │   └── IsNumeric.php
│   ├── String/
│   │   ├── Min.php
│   │   ├── Max.php
│   │   ├── Regex.php
│   │   ├── Alpha.php
│   │   ├── Lowercase.php
│   │   ├── Uppercase.php
│   │   ├── StartsWith.php
│   │   ├── EndsWith.php
│   │   └── Contains.php
│   ├── Int/
│   │   ├── GreaterThan.php
│   │   └── LessThan.php
│   └── Orm/
│       ├── Abstract/
│       │   └── EntityRule.php
│       ├── EntityExists.php
│       └── EntityNotExists.php
├── Validator/
│   ├── Validator.php
│   └── Contract/
│       └── ValidatorContract.php
├── Throwable/
│   ├── Contract/
│   │   └── ValidationThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── ValidationInvalidArgumentException.php
│       │   └── ValidationRuntimeException.php
│       └── ValidationRuleFailureException.php
└── README.md
```

### 25. View Module

*58 files, 26 directories*

```
├── Factory/
│   ├── ViewResponseFactory.php
│   └── Contract/
│       └── ViewResponseFactoryContract.php
├── Orka/
│   ├── Constant/
│   │   └── OrkaReplacement.php
│   └── Replacement/
│       ├── Block/
│       │   ├── Block.php
│       │   ├── StartBlock.php
│       │   ├── EndBlock.php
│       │   └── TrimBlock.php
│       ├── Comment/
│       │   ├── SingleLine.php
│       │   ├── StartMultiline.php
│       │   └── EndMultiline.php
│       ├── Debug/
│       │   └── Dd.php
│       ├── Layout.php
│       ├── Partial/
│       │   ├── Partial.php
│       │   ├── PartialWithVariables.php
│       │   ├── TrimPartial.php
│       │   └── TrimPartialWithVariables.php
│       ├── Statement/
│       │   ├── Break_.php
│       │   ├── Conditional/
│       │   │   ├── If_.php
│       │   │   ├── ElseIf_.php
│       │   │   ├── Else_.php
│       │   │   ├── Unless.php
│       │   │   ├── ElseUnless.php
│       │   │   ├── EndIf_.php
│       │   │   ├── Isset_.php
│       │   │   └── Empty_.php
│       │   └── Iterate/
│       │       ├── For_.php
│       │       ├── EndFor_.php
│       │       ├── Foreach_.php
│       │       └── EndForeach_.php
│       ├── Variable/
│       │   ├── Escaped.php
│       │   ├── Unescaped.php
│       │   ├── SetVariable.php
│       │   └── SetVariables.php
│       └── Contract/
│           └── ReplacementContract.php
├── Provider/
│   ├── ViewComponentProvider.php
│   └── ViewServiceProvider.php
├── Renderer/
│   ├── PhpRenderer.php
│   ├── TwigRenderer.php
│   ├── OrkaRenderer.php
│   └── Contract/
│       └── RendererContract.php
├── Template/
│   ├── Template.php
│   └── Contract/
│       └── TemplateContract.php
├── Throwable/
│   ├── Contract/
│   │   └── ViewThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── ViewInvalidArgumentException.php
│       │   └── ViewRuntimeException.php
│       ├── ViewEscapeEncodingFailureException.php
│       ├── ViewInvalidPathException.php
│       ├── ViewOrkaCacheFailureException.php
│       ├── ViewRenderFailureException.php
│       └── [Additional view exceptions...]
└── README.md
```

### 26. Documentation Files (Root Level)

```
├── README.md
├── APPLICATION_STRUCTURE.md
├── GETTING_STARTED.md
├── LIFECYCLE.md
└── VERSIONING_AND_RELEASE_PROCESS.md
```

---

## Key Architectural Patterns

### 1. Manager Pattern

Most modules have `Manager/` subdirectories with:

- Multiple implementations (e.g., `RedisCache`, `LogCache`, `NullCache`)
- Contract/Interface definitions
- Provider pattern for service registration

### 2. Service Providers

Every module has `Provider/` containing:

- `ComponentProvider` (dependency injection setup)
- `ServiceProvider` (service registration)
- Contract definitions

### 3. Exception Handling

Structured exception hierarchy:

- Abstract base exceptions (`InvalidArgumentException`, `RuntimeException`)
- Module-specific exceptions
- `ThrowableContract` interfaces

### 4. Type System

Extensive Type module with:

- Basic types (`Bool`, `Int`, `Float`, `String`, `Null`)
- Collection types (`Array`, `Collection`)
- Unique ID types (`UUID`, `ULID`, `VLID`)
- JSON and serialization support

### 5. HTTP Routing

Comprehensive HTTP request/response handling:

- Route matching and dispatch
- Middleware pipeline
- Controller abstractions
- Response factories

### 6. ORM Layer

Complete database abstraction:

- Multiple database drivers (MySQL, PostgreSQL, SQLite)
- Query builders for different operations
- Schema migrations
- Entity mapping

### 7. Session Management

Multiple session implementations:

- Cookie-based
- Cache-based
- JWT-based
- Token-based (both HTTP headers and CLI options)
