# Complete Valkyrja PHP Source Directory Structure

**Base Path:** `src/Valkyrja` in the `valkyrjaio/valkyrja-php` repository

## Statistics

- **Total Files:** 1,214
- **Total PHP Files:** 1,182
- **Total Markdown Files:** 32
- **Total Directories:** 619
- **Modules:** 24

Each tree below shows every directory and every file in the module. The file
total counts every file below the base path, whatever the extension. The PHP
total and the Markdown total add up to that file total. The directory total
counts every directory below the base path. The directory total does not count
the base path.

A module count reads `_N files, M directories_`. `N` counts every file in the
module. `M` counts the module directory and every directory below it. Each tree
starts inside the module directory and never lists that directory, so a tree
holds `N` plus `M` minus one entries.

---

## Directory Tree by Module

### 1. API Module (`Api/`)

_13 files, 12 directories_

```
├── Constant/
│   └── Status.php
├── Manager/
│   ├── Contract/
│   │   └── ApiContract.php
│   └── Api.php
├── Middleware/
│   └── ApiThrowableCaughtMiddleware.php
├── Model/
│   ├── Contract/
│   │   ├── JsonContract.php
│   │   └── JsonDataContract.php
│   ├── Json.php
│   └── JsonData.php
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

### 2. Application Module (`Application/`)

_27 files, 18 directories_

```
├── Constant/
│   └── ApplicationInfo.php
├── Data/
│   ├── Contract/
│   │   ├── CliConfigContract.php
│   │   ├── ConfigContract.php
│   │   └── HttpConfigContract.php
│   ├── CliConfig.php
│   ├── Config.php
│   └── HttpConfig.php
├── Directory/
│   └── Directory.php
├── Entry/
│   ├── Abstract/
│   │   ├── App.php
│   │   └── WorkerHttp.php
│   ├── FrankenPhp/
│   │   └── FrankenPhpHttp.php
│   ├── OpenSwoole/
│   │   └── OpenSwooleHttp.php
│   ├── RoadRunner/
│   │   └── RoadRunnerHttp.php
│   ├── Cli.php
│   └── Http.php
├── Kernel/
│   ├── Contract/
│   │   └── ApplicationContract.php
│   ├── ChildApplication.php
│   └── Valkyrja.php
├── Provider/
│   ├── Contract/
│   │   └── ComponentProviderContract.php
│   ├── ApplicationComponentProvider.php
│   ├── CliApplicationComponentProvider.php
│   ├── CliWithHttpApplicationComponentProvider.php
│   └── HttpApplicationComponentProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── ApplicationThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── ApplicationInvalidArgumentException.php
│           └── ApplicationRuntimeException.php
└── README.md
```

### 3. Attribute Module (`Attribute/`)

_10 files, 10 directories_

```
├── Collector/
│   ├── Contract/
│   │   └── CollectorContract.php
│   └── Collector.php
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

### 4. Auth Module (`Auth/`)

_75 files, 23 directories_

```
├── Authenticator/
│   ├── Abstract/
│   │   └── Authenticator.php
│   ├── Contract/
│   │   └── AuthenticatorContract.php
│   └── SessionAuthenticator.php
├── Constant/
│   ├── RouteName.php
│   ├── SessionItemId.php
│   └── UserField.php
├── Data/
│   ├── Attempt/
│   │   ├── Contract/
│   │   │   ├── AuthenticationAttemptContract.php
│   │   │   ├── ForgotPasswordAttemptContract.php
│   │   │   ├── LockAttemptContract.php
│   │   │   ├── ResetPasswordAttemptContract.php
│   │   │   └── UnlockAttemptContract.php
│   │   ├── AuthenticationAttempt.php
│   │   ├── ForgotPasswordAttempt.php
│   │   ├── LockAttempt.php
│   │   ├── ResetPasswordAttempt.php
│   │   └── UnlockAttempt.php
│   ├── Contract/
│   │   ├── AuthConfigContract.php
│   │   └── AuthenticatedUsersContract.php
│   ├── Retrieval/
│   │   ├── Contract/
│   │   │   └── RetrievalContract.php
│   │   ├── RetrievalById.php
│   │   ├── RetrievalByIdAndUsername.php
│   │   ├── RetrievalByResetToken.php
│   │   └── RetrievalByUsername.php
│   ├── AuthConfig.php
│   ├── AuthSessionConfig.php
│   └── AuthenticatedUsers.php
├── Entity/
│   ├── Contract/
│   │   ├── AntiPhishCodeUserContract.php
│   │   ├── DeviceAuthenticatedUserContract.php
│   │   ├── LastOnlineUserContract.php
│   │   ├── LockableUserContract.php
│   │   ├── MailableUserContract.php
│   │   ├── PermissibleUserContract.php
│   │   ├── PinUserContract.php
│   │   ├── TwoFactorUserContract.php
│   │   ├── UserContract.php
│   │   ├── UserDeviceContract.php
│   │   ├── UserRecoveryCodeContract.php
│   │   └── VerifiableUserContract.php
│   ├── Trait/
│   │   ├── LockableUserFields.php
│   │   ├── LockableUserMethods.php
│   │   ├── MailableUserFields.php
│   │   ├── MailableUserMethods.php
│   │   ├── UserFields.php
│   │   ├── UserMethods.php
│   │   ├── VerifiableUserFields.php
│   │   └── VerifiableUserMethods.php
│   ├── LockableUser.php
│   ├── MailableUser.php
│   ├── User.php
│   └── VerifiableUser.php
├── Hasher/
│   ├── Contract/
│   │   └── PasswordHasherContract.php
│   └── PhpPasswordHasher.php
├── Provider/
│   ├── AuthComponentProvider.php
│   └── AuthServiceProvider.php
├── Store/
│   ├── Contract/
│   │   └── StoreContract.php
│   ├── InMemoryStore.php
│   ├── NullStore.php
│   └── OrmStore.php
├── Throwable/
│   ├── Contract/
│   │   └── AuthThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── AuthInvalidArgumentException.php
│       │   └── AuthRuntimeException.php
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
│       └── AuthUnexpectedUsernameValueException.php
└── README.md
```

### 5. Broadcast Module (`Broadcast/`)

_19 files, 10 directories_

```
├── Broadcaster/
│   ├── Contract/
│   │   └── BroadcasterContract.php
│   ├── CryptPusherBroadcaster.php
│   ├── LogBroadcaster.php
│   ├── NullBroadcaster.php
│   └── PusherBroadcaster.php
├── Data/
│   ├── Contract/
│   │   ├── BroadcastConfigContract.php
│   │   ├── BroadcastLogConfigContract.php
│   │   ├── BroadcastPusherConfigContract.php
│   │   └── MessageContract.php
│   ├── BroadcastConfig.php
│   ├── BroadcastLogConfig.php
│   ├── BroadcastPusherConfig.php
│   └── Message.php
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

### 6. Cache Module (`Cache/`)

_20 files, 12 directories_

```
├── Data/
│   ├── Contract/
│   │   ├── CacheConfigContract.php
│   │   ├── CacheLogConfigContract.php
│   │   ├── CacheNullConfigContract.php
│   │   └── CacheRedisConfigContract.php
│   ├── CacheConfig.php
│   ├── CacheLogConfig.php
│   ├── CacheNullConfig.php
│   └── CacheRedisConfig.php
├── Manager/
│   ├── Contract/
│   │   └── CacheContract.php
│   ├── LogCache.php
│   ├── NullCache.php
│   └── RedisCache.php
├── Provider/
│   ├── CacheComponentProvider.php
│   └── CacheServiceProvider.php
├── Tagger/
│   ├── Contract/
│   │   └── TaggerContract.php
│   └── Tagger.php
├── Throwable/
│   ├── Contract/
│   │   └── CacheThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── CacheInvalidArgumentException.php
│           └── CacheRuntimeException.php
└── README.md
```

### 7. CLI Module (`Cli/`)

_172 files, 84 directories_

```
├── Interaction/
│   ├── Argument/
│   │   ├── Contract/
│   │   │   └── ArgumentContract.php
│   │   ├── Factory/
│   │   │   └── ArgumentFactory.php
│   │   └── Argument.php
│   ├── Data/
│   │   ├── Contract/
│   │   │   └── CliInteractionConfigContract.php
│   │   └── CliInteractionConfig.php
│   ├── Enum/
│   │   ├── BackgroundColor.php
│   │   ├── ExitCode.php
│   │   ├── OptionType.php
│   │   ├── Style.php
│   │   └── TextColor.php
│   ├── Format/
│   │   ├── Contract/
│   │   │   └── FormatContract.php
│   │   ├── BackgroundColorFormat.php
│   │   ├── Format.php
│   │   ├── StyleFormat.php
│   │   └── TextColorFormat.php
│   ├── Formatter/
│   │   ├── Contract/
│   │   │   └── FormatterContract.php
│   │   ├── ErrorFormatter.php
│   │   ├── Formatter.php
│   │   ├── HighlightedTextFormatter.php
│   │   ├── QuestionFormatter.php
│   │   ├── SuccessFormatter.php
│   │   └── WarningFormatter.php
│   ├── Input/
│   │   ├── Contract/
│   │   │   └── InputContract.php
│   │   ├── Factory/
│   │   │   └── InputFactory.php
│   │   └── Input.php
│   ├── Message/
│   │   ├── Contract/
│   │   │   ├── AnswerContract.php
│   │   │   ├── MessageContract.php
│   │   │   ├── ProgressContract.php
│   │   │   └── QuestionContract.php
│   │   ├── Answer.php
│   │   ├── Banner.php
│   │   ├── ErrorMessage.php
│   │   ├── Header.php
│   │   ├── Message.php
│   │   ├── Messages.php
│   │   ├── NewLine.php
│   │   ├── Progress.php
│   │   ├── Question.php
│   │   ├── SuccessMessage.php
│   │   └── WarningMessage.php
│   ├── Option/
│   │   ├── Contract/
│   │   │   └── OptionContract.php
│   │   ├── Factory/
│   │   │   └── OptionFactory.php
│   │   └── Option.php
│   ├── Output/
│   │   ├── Contract/
│   │   │   ├── EmptyOutputContract.php
│   │   │   ├── FileOutputContract.php
│   │   │   ├── OutputContract.php
│   │   │   ├── PlainOutputContract.php
│   │   │   └── StreamOutputContract.php
│   │   ├── Factory/
│   │   │   ├── Contract/
│   │   │   │   └── OutputFactoryContract.php
│   │   │   └── OutputFactory.php
│   │   ├── EmptyOutput.php
│   │   ├── FileOutput.php
│   │   ├── Output.php
│   │   ├── PlainOutput.php
│   │   └── StreamOutput.php
│   ├── Provider/
│   │   ├── CliInteractionComponentProvider.php
│   │   └── CliInteractionServiceProvider.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── CliInteractionThrowable.php
│   │   └── Exception/
│   │       ├── Abstract/
│   │       │   ├── CliInteractionInvalidArgumentException.php
│   │       │   └── CliInteractionRuntimeException.php
│   │       ├── CliInteractionExpectedQuestionOutputException.php
│   │       ├── CliInteractionInvalidEmptyValueException.php
│   │       ├── CliInteractionInvalidNonEmptyValueException.php
│   │       ├── CliInteractionInvalidOptionNameException.php
│   │       ├── CliInteractionNoFormatterException.php
│   │       └── CliInteractionNoValidationCallableException.php
│   └── Writer/
│       ├── Contract/
│       │   └── WriterContract.php
│       └── QuestionWriter.php
├── Middleware/
│   ├── Contract/
│   │   ├── InputReceivedMiddlewareContract.php
│   │   ├── ProcessExitingMiddlewareContract.php
│   │   ├── RouteDispatchedMiddlewareContract.php
│   │   ├── RouteMatchedMiddlewareContract.php
│   │   ├── RouteNotMatchedMiddlewareContract.php
│   │   └── ThrowableCaughtMiddlewareContract.php
│   ├── Handler/
│   │   ├── Abstract/
│   │   │   └── Handler.php
│   │   ├── Contract/
│   │   │   ├── HandlerContract.php
│   │   │   ├── InputReceivedHandlerContract.php
│   │   │   ├── ProcessExitingHandlerContract.php
│   │   │   ├── RouteDispatchedHandlerContract.php
│   │   │   ├── RouteMatchedHandlerContract.php
│   │   │   ├── RouteNotMatchedHandlerContract.php
│   │   │   └── ThrowableCaughtHandlerContract.php
│   │   ├── InputReceivedHandler.php
│   │   ├── ProcessExitingHandler.php
│   │   ├── RouteDispatchedHandler.php
│   │   ├── RouteMatchedHandler.php
│   │   ├── RouteNotMatchedHandler.php
│   │   └── ThrowableCaughtHandler.php
│   ├── Provider/
│   │   ├── CliMiddlewareComponentProvider.php
│   │   └── CliMiddlewareServiceProvider.php
│   └── Throwable/
│       ├── Contract/
│       │   └── CliMiddlewareThrowable.php
│       └── Exception/
│           └── Abstract/
│               ├── CliMiddlewareInvalidArgumentException.php
│               └── CliMiddlewareRuntimeException.php
├── Routing/
│   ├── Attribute/
│   │   ├── Route/
│   │   │   ├── Middleware.php
│   │   │   ├── Name.php
│   │   │   └── RouteHandler.php
│   │   ├── ArgumentParameter.php
│   │   ├── OptionParameter.php
│   │   └── Route.php
│   ├── Collection/
│   │   ├── Contract/
│   │   │   └── RouteCollectionContract.php
│   │   └── RouteCollection.php
│   ├── Collector/
│   │   ├── Contract/
│   │   │   └── RouteCollectorContract.php
│   │   └── AttributeRouteCollector.php
│   ├── Constant/
│   │   ├── OptionName.php
│   │   └── OptionShortName.php
│   ├── Controller/
│   │   └── Controller.php
│   ├── Data/
│   │   ├── Abstract/
│   │   │   └── Parameter.php
│   │   ├── Contract/
│   │   │   ├── ArgumentParameterContract.php
│   │   │   ├── CliRoutingConfigContract.php
│   │   │   ├── OptionParameterContract.php
│   │   │   ├── ParameterContract.php
│   │   │   └── RouteContract.php
│   │   ├── Option/
│   │   │   ├── HelpOptionParameter.php
│   │   │   ├── NoInteractionOptionParameter.php
│   │   │   ├── QuietOptionParameter.php
│   │   │   ├── SilentOptionParameter.php
│   │   │   └── VersionOptionParameter.php
│   │   ├── ArgumentParameter.php
│   │   ├── CliRoutingData.php
│   │   ├── OptionParameter.php
│   │   └── Route.php
│   ├── Dispatcher/
│   │   ├── Contract/
│   │   │   └── RouterContract.php
│   │   └── Router.php
│   ├── Enum/
│   │   ├── ArgumentMode.php
│   │   ├── ArgumentValueMode.php
│   │   ├── OptionMode.php
│   │   └── OptionValueMode.php
│   ├── Provider/
│   │   ├── Contract/
│   │   │   └── CliRouteProviderContract.php
│   │   ├── CliRoutingCliRouteProvider.php
│   │   ├── CliRoutingComponentProvider.php
│   │   └── CliRoutingServiceProvider.php
│   └── Throwable/
│       ├── Contract/
│       │   └── CliRoutingThrowable.php
│       └── Exception/
│           ├── Abstract/
│           │   ├── CliRoutingInvalidArgumentException.php
│           │   └── CliRoutingRuntimeException.php
│           ├── CliRoutingArgumentValuesValidationException.php
│           ├── CliRoutingInvalidArgumentNameException.php
│           ├── CliRoutingInvalidHelpTextCallableException.php
│           ├── CliRoutingInvalidOptionNameException.php
│           ├── CliRoutingInvalidOptionWithValueException.php
│           ├── CliRoutingInvalidRouteNameException.php
│           ├── CliRoutingNoCastException.php
│           ├── CliRoutingNoHelpTextException.php
│           ├── CliRoutingNoOutputDispatchException.php
│           └── CliRoutingOptionValuesValidationException.php
├── Server/
│   ├── Command/
│   │   ├── HelpCommand.php
│   │   ├── ListBashCommand.php
│   │   ├── ListCommand.php
│   │   └── VersionCommand.php
│   ├── Constant/
│   │   └── CommandName.php
│   ├── Data/
│   │   └── Contract/
│   │       ├── CliHelpCommandConfigContract.php
│   │       ├── CliNoInteractionConfigContract.php
│   │       ├── CliQuietInteractionConfigContract.php
│   │       ├── CliSilentInteractionConfigContract.php
│   │       └── CliVersionCommandConfigContract.php
│   ├── Handler/
│   │   ├── Contract/
│   │   │   └── InputHandlerContract.php
│   │   └── InputHandler.php
│   ├── Middleware/
│   │   ├── InputReceived/
│   │   │   ├── CheckForHelpOptionsMiddleware.php
│   │   │   ├── CheckForVersionOptionsMiddleware.php
│   │   │   └── CheckGlobalInteractionOptionsMiddleware.php
│   │   ├── RouteNotMatched/
│   │   │   └── CheckCommandForTypoMiddleware.php
│   │   └── ThrowableCaught/
│   │       ├── LogThrowableCaughtMiddleware.php
│   │       └── OutputThrowableCaughtMiddleware.php
│   ├── Provider/
│   │   ├── CliServerComponentProvider.php
│   │   └── CliServerServiceProvider.php
│   ├── Support/
│   │   └── Exiter.php
│   └── Throwable/
│       ├── Contract/
│       │   └── CliServerThrowable.php
│       └── Exception/
│           └── Abstract/
│               ├── CliServerInvalidArgumentException.php
│               └── CliServerRuntimeException.php
├── Throwable/
│   ├── Contract/
│   │   └── CliThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── CliInvalidArgumentException.php
│           └── CliRuntimeException.php
└── README.md
```

### 8. Container Module (`Container/`)

_16 files, 11 directories_

```
├── Data/
│   └── ContainerData.php
├── Manager/
│   ├── Contract/
│   │   ├── ContainerContract.php
│   │   └── ProvidersAwareContract.php
│   ├── Trait/
│   │   └── ProvidersAware.php
│   ├── ChildContainer.php
│   ├── Container.php
│   └── NativeChildContainer.php
├── Provider/
│   ├── Contract/
│   │   └── ServiceProviderContract.php
│   ├── ContainerComponentProvider.php
│   └── ContainerServiceProvider.php
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

### 9. Crypt Module (`Crypt/`)

_16 files, 10 directories_

```
├── Data/
│   ├── Contract/
│   │   └── CryptConfigContract.php
│   └── CryptConfig.php
├── Manager/
│   ├── Contract/
│   │   └── CryptContract.php
│   ├── NullCrypt.php
│   └── SodiumCrypt.php
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

### 10. Event Module (`Event/`)

_21 files, 17 directories_

```
├── Attribute/
│   ├── Listener.php
│   └── ListenerHandler.php
├── Collection/
│   ├── Contract/
│   │   └── ListenerCollectionContract.php
│   └── ListenerCollection.php
├── Collector/
│   ├── Contract/
│   │   └── ListenerCollectorContract.php
│   └── AttributeListenerCollector.php
├── Contract/
│   ├── ArgumentsCapableEventContract.php
│   └── DispatchCollectableEventContract.php
├── Data/
│   ├── Contract/
│   │   └── ListenerContract.php
│   ├── EventData.php
│   └── Listener.php
├── Dispatcher/
│   ├── Contract/
│   │   └── EventDispatcherContract.php
│   └── EventDispatcher.php
├── Provider/
│   ├── Contract/
│   │   └── ListenerProviderContract.php
│   ├── EventComponentProvider.php
│   └── EventServiceProvider.php
├── Throwable/
│   ├── Contract/
│   │   └── EventThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── EventInvalidArgumentException.php
│       │   └── EventRuntimeException.php
│       └── EventInvalidEventException.php
└── README.md
```

### 11. Filesystem Module (`Filesystem/`)

_25 files, 11 directories_

```
├── Data/
│   ├── Contract/
│   │   ├── FilesystemConfigContract.php
│   │   ├── FilesystemFlysystemConfigContract.php
│   │   ├── FilesystemFlysystemLocalConfigContract.php
│   │   └── FilesystemFlysystemS3ConfigContract.php
│   ├── FilesystemConfig.php
│   ├── FilesystemFlysystemConfig.php
│   ├── FilesystemFlysystemLocalConfig.php
│   ├── FilesystemFlysystemS3Config.php
│   ├── InMemoryFile.php
│   └── InMemoryMetadata.php
├── Enum/
│   └── Visibility.php
├── Manager/
│   ├── Contract/
│   │   └── FilesystemContract.php
│   ├── FlysystemFilesystem.php
│   ├── InMemoryFilesystem.php
│   ├── LocalFlysystemFilesystem.php
│   ├── NullFilesystem.php
│   └── S3FlysystemFilesystem.php
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

### 12. HTTP Module (`Http/`)

_300 files, 161 directories_

```
├── Client/
│   ├── Data/
│   │   ├── Contract/
│   │   │   └── HttpClientConfigContract.php
│   │   └── HttpClientConfig.php
│   ├── Manager/
│   │   ├── Contract/
│   │   │   └── ClientContract.php
│   │   ├── GuzzleClient.php
│   │   ├── LogClient.php
│   │   └── NullClient.php
│   ├── Provider/
│   │   ├── HttpClientComponentProvider.php
│   │   └── HttpClientServiceProvider.php
│   └── Throwable/
│       ├── Contract/
│       │   └── HttpClientThrowable.php
│       └── Exception/
│           └── Abstract/
│               ├── HttpClientInvalidArgumentException.php
│               └── HttpClientRuntimeException.php
├── Message/
│   ├── Constant/
│   │   ├── ContentTypeValue.php
│   │   ├── HeaderName.php
│   │   ├── HeaderValue.php
│   │   └── Port.php
│   ├── Contract/
│   │   └── MessageContract.php
│   ├── Enum/
│   │   ├── ProtocolVersion.php
│   │   ├── RequestMethod.php
│   │   ├── SameSite.php
│   │   ├── StatusCode.php
│   │   └── StatusText.php
│   ├── File/
│   │   ├── Collection/
│   │   │   ├── Contract/
│   │   │   │   └── UploadedFileCollectionContract.php
│   │   │   └── UploadedFileCollection.php
│   │   ├── Constant/
│   │   │   └── UploadErrorExceptionMessage.php
│   │   ├── Contract/
│   │   │   └── UploadedFileContract.php
│   │   ├── Enum/
│   │   │   └── UploadError.php
│   │   ├── Factory/
│   │   │   ├── PsrUploadedFileFactory.php
│   │   │   └── UploadedFileFactory.php
│   │   ├── Psr/
│   │   │   └── UploadedFile.php
│   │   ├── Throwable/
│   │   │   ├── Contract/
│   │   │   │   └── UploadedFileThrowable.php
│   │   │   └── Exception/
│   │   │       ├── Abstract/
│   │   │       │   ├── UploadedFileInvalidArgumentException.php
│   │   │       │   └── UploadedFileRuntimeException.php
│   │   │       ├── UploadedFileAlreadyMovedException.php
│   │   │       ├── UploadedFileInvalidDirectoryException.php
│   │   │       ├── UploadedFileInvalidFilesArrayStructureException.php
│   │   │       ├── UploadedFileInvalidKeyException.php
│   │   │       ├── UploadedFileInvalidParamException.php
│   │   │       ├── UploadedFileInvalidTmpNameException.php
│   │   │       ├── UploadedFileInvalidUploadErrorException.php
│   │   │       ├── UploadedFileInvalidUploadedFileException.php
│   │   │       ├── UploadedFileInvalidValueException.php
│   │   │       ├── UploadedFileMoveFailureException.php
│   │   │       ├── UploadedFileUnableToWriteFileException.php
│   │   │       └── UploadedFileUploadErrorException.php
│   │   └── UploadedFile.php
│   ├── Header/
│   │   ├── Collection/
│   │   │   ├── Contract/
│   │   │   │   └── HeaderCollectionContract.php
│   │   │   └── HeaderCollection.php
│   │   ├── Contract/
│   │   │   └── HeaderContract.php
│   │   ├── Factory/
│   │   │   ├── CookieFactory.php
│   │   │   ├── HeaderFactory.php
│   │   │   └── PsrHeaderFactory.php
│   │   ├── Throwable/
│   │   │   ├── Contract/
│   │   │   │   └── HttpHeaderThrowable.php
│   │   │   └── Exception/
│   │   │       ├── Abstract/
│   │   │       │   ├── HttpHeaderInvalidArgumentException.php
│   │   │       │   └── HttpHeaderRuntimeException.php
│   │   │       ├── HttpHeaderInvalidHeaderNameException.php
│   │   │       ├── HttpHeaderInvalidHeaderParamException.php
│   │   │       ├── HttpHeaderInvalidNameException.php
│   │   │       ├── HttpHeaderInvalidValueException.php
│   │   │       ├── HttpHeaderUnsupportedMethodException.php
│   │   │       ├── HttpHeaderUnsupportedOffsetSetException.php
│   │   │       └── HttpHeaderUnsupportedOffsetUnsetException.php
│   │   ├── Value/
│   │   │   ├── Component/
│   │   │   │   ├── Contract/
│   │   │   │   │   └── ComponentContract.php
│   │   │   │   └── Component.php
│   │   │   ├── Contract/
│   │   │   │   ├── CookieContract.php
│   │   │   │   └── ValueContract.php
│   │   │   ├── Cookie.php
│   │   │   └── Value.php
│   │   ├── ContentType.php
│   │   ├── Header.php
│   │   ├── Location.php
│   │   ├── Referer.php
│   │   └── SetCookie.php
│   ├── Param/
│   │   ├── Abstract/
│   │   │   └── ParamCollection.php
│   │   ├── Contract/
│   │   │   ├── AttributeParamCollectionContract.php
│   │   │   ├── CookieParamCollectionContract.php
│   │   │   ├── ParamCollectionContract.php
│   │   │   ├── ParsedBodyParamCollectionContract.php
│   │   │   ├── ParsedJsonParamCollectionContract.php
│   │   │   ├── QueryParamCollectionContract.php
│   │   │   └── ServerParamCollectionContract.php
│   │   ├── AttributeParamCollection.php
│   │   ├── CookieParamCollection.php
│   │   ├── ParsedBodyParamCollection.php
│   │   ├── ParsedJsonParamCollection.php
│   │   ├── QueryParamCollection.php
│   │   └── ServerParamCollection.php
│   ├── Provider/
│   │   ├── HttpMessageComponentProvider.php
│   │   └── HttpMessageServiceProvider.php
│   ├── Request/
│   │   ├── Contract/
│   │   │   ├── JsonServerRequestContract.php
│   │   │   ├── RequestContract.php
│   │   │   └── ServerRequestContract.php
│   │   ├── Factory/
│   │   │   ├── PsrRequestFactory.php
│   │   │   ├── RequestFactory.php
│   │   │   └── ServerFactory.php
│   │   ├── Psr/
│   │   │   ├── Request.php
│   │   │   └── ServerRequest.php
│   │   ├── Throwable/
│   │   │   ├── Contract/
│   │   │   │   └── HttpRequestThrowable.php
│   │   │   └── Exception/
│   │   │       ├── Abstract/
│   │   │       │   ├── HttpRequestInvalidArgumentException.php
│   │   │       │   └── HttpRequestRuntimeException.php
│   │   │       ├── HttpRequestInvalidMethodException.php
│   │   │       └── HttpRequestInvalidRequestTargetException.php
│   │   ├── JsonServerRequest.php
│   │   ├── Request.php
│   │   └── ServerRequest.php
│   ├── Response/
│   │   ├── Contract/
│   │   │   ├── EmptyResponseContract.php
│   │   │   ├── HtmlResponseContract.php
│   │   │   ├── JsonResponseContract.php
│   │   │   ├── RedirectResponseContract.php
│   │   │   ├── ResponseContract.php
│   │   │   └── TextResponseContract.php
│   │   ├── Factory/
│   │   │   ├── Contract/
│   │   │   │   └── ResponseFactoryContract.php
│   │   │   └── ResponseFactory.php
│   │   ├── Psr/
│   │   │   └── Response.php
│   │   ├── Throwable/
│   │   │   ├── Contract/
│   │   │   │   └── HttpResponseThrowable.php
│   │   │   └── Exception/
│   │   │       ├── Abstract/
│   │   │       │   ├── HttpResponseInvalidArgumentException.php
│   │   │       │   └── HttpResponseRuntimeException.php
│   │   │       ├── HttpRequestInvalidJsonCallbackException.php
│   │   │       └── HttpRequestInvalidRedirectStatusCodeException.php
│   │   ├── EmptyResponse.php
│   │   ├── HtmlResponse.php
│   │   ├── JsonResponse.php
│   │   ├── RedirectResponse.php
│   │   ├── Response.php
│   │   ├── TextResponse.php
│   │   └── XmlResponse.php
│   ├── Stream/
│   │   ├── Contract/
│   │   │   └── StreamContract.php
│   │   ├── Enum/
│   │   │   ├── Mode.php
│   │   │   ├── ModeTranslation.php
│   │   │   └── PhpWrapper.php
│   │   ├── Factory/
│   │   │   ├── PsrStreamFactory.php
│   │   │   └── StreamFactory.php
│   │   ├── Psr/
│   │   │   └── Stream.php
│   │   ├── Throwable/
│   │   │   ├── Contract/
│   │   │   │   └── HttpStreamThrowable.php
│   │   │   └── Exception/
│   │   │       ├── Abstract/
│   │   │       │   ├── HttpStreamInvalidArgumentException.php
│   │   │       │   └── HttpStreamRuntimeException.php
│   │   │       ├── HttpStreamInvalidLengthException.php
│   │   │       ├── HttpStreamInvalidStreamException.php
│   │   │       ├── HttpStreamNoStreamAvailableException.php
│   │   │       ├── HttpStreamStreamReadException.php
│   │   │       ├── HttpStreamStreamSeekException.php
│   │   │       ├── HttpStreamStreamTellException.php
│   │   │       ├── HttpStreamStreamWriteException.php
│   │   │       ├── HttpStreamUnreadableStreamException.php
│   │   │       ├── HttpStreamUnseekableStreamException.php
│   │   │       └── HttpStreamUnwritableStreamException.php
│   │   └── Stream.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── HttpMessageThrowable.php
│   │   └── Exception/
│   │       ├── Abstract/
│   │       │   ├── HttpMessageInvalidArgumentException.php
│   │       │   └── HttpMessageRuntimeException.php
│   │       ├── HttpNotFoundResponseException.php
│   │       ├── HttpRedirectResponseException.php
│   │       └── HttpResponseException.php
│   ├── Trait/
│   │   └── Message.php
│   └── Uri/
│       ├── Constant/
│       │   └── Char.php
│       ├── Contract/
│       │   └── UriContract.php
│       ├── Data/
│       │   └── HostPortAccumulator.php
│       ├── Enum/
│       │   └── Scheme.php
│       ├── Factory/
│       │   ├── MarshalUriFactory.php
│       │   ├── PsrUriFactory.php
│       │   └── UriFactory.php
│       ├── Psr/
│       │   └── Uri.php
│       ├── Throwable/
│       │   ├── Contract/
│       │   │   └── HttpMessageThrowable.php
│       │   └── Exception/
│       │       ├── Abstract/
│       │       │   ├── HttpUriInvalidArgumentException.php
│       │       │   └── HttpUriRuntimeException.php
│       │       ├── HttpUriInvalidFromStringException.php
│       │       ├── HttpUriInvalidPathException.php
│       │       ├── HttpUriInvalidPortException.php
│       │       ├── HttpUriInvalidQueryException.php
│       │       └── NoPortExceptionHttpUri.php
│       ├── Type/
│       │   └── Port.php
│       └── Uri.php
├── Middleware/
│   ├── Contract/
│   │   ├── RequestReceivedMiddlewareContract.php
│   │   ├── ResponseSentMiddlewareContract.php
│   │   ├── RouteDispatchedMiddlewareContract.php
│   │   ├── RouteMatchedMiddlewareContract.php
│   │   ├── RouteNotMatchedMiddlewareContract.php
│   │   ├── SendingResponseMiddlewareContract.php
│   │   └── ThrowableCaughtMiddlewareContract.php
│   ├── Handler/
│   │   ├── Abstract/
│   │   │   └── Handler.php
│   │   ├── Contract/
│   │   │   ├── HandlerContract.php
│   │   │   ├── RequestReceivedHandlerContract.php
│   │   │   ├── ResponseSentHandlerContract.php
│   │   │   ├── RouteDispatchedHandlerContract.php
│   │   │   ├── RouteMatchedHandlerContract.php
│   │   │   ├── RouteNotMatchedHandlerContract.php
│   │   │   ├── SendingResponseHandlerContract.php
│   │   │   └── ThrowableCaughtHandlerContract.php
│   │   ├── RequestReceivedHandler.php
│   │   ├── ResponseSentHandler.php
│   │   ├── RouteDispatchedHandler.php
│   │   ├── RouteMatchedHandler.php
│   │   ├── RouteNotMatchedHandler.php
│   │   ├── SendingResponseHandler.php
│   │   └── ThrowableCaughtHandler.php
│   ├── Provider/
│   │   ├── HttpMiddlewareComponentProvider.php
│   │   └── HttpMiddlewareServiceProvider.php
│   └── Throwable/
│       ├── Contract/
│       │   └── HttpMiddlewareThrowable.php
│       └── Exception/
│           └── Abstract/
│               ├── HttpMiddlewareInvalidArgumentException.php
│               └── HttpMiddlewareRuntimeException.php
├── Routing/
│   ├── Attribute/
│   │   ├── Route/
│   │   │   ├── RequestMethod/
│   │   │   │   ├── Any.php
│   │   │   │   ├── Connect.php
│   │   │   │   ├── Delete.php
│   │   │   │   ├── Get.php
│   │   │   │   ├── Head.php
│   │   │   │   ├── Options.php
│   │   │   │   ├── Patch.php
│   │   │   │   ├── Post.php
│   │   │   │   ├── Put.php
│   │   │   │   └── Trace.php
│   │   │   ├── Middleware.php
│   │   │   ├── Name.php
│   │   │   ├── Path.php
│   │   │   ├── RequestMethod.php
│   │   │   ├── RequestStruct.php
│   │   │   ├── ResponseStruct.php
│   │   │   └── RouteHandler.php
│   │   ├── DynamicRoute.php
│   │   ├── Parameter.php
│   │   └── Route.php
│   ├── Cli/
│   │   └── Command/
│   │       ├── Constant/
│   │       │   └── CommandName.php
│   │       └── ListCommand.php
│   ├── Collection/
│   │   ├── Contract/
│   │   │   └── RouteCollectionContract.php
│   │   └── RouteCollection.php
│   ├── Collector/
│   │   ├── Contract/
│   │   │   └── RouteCollectorContract.php
│   │   └── AttributeRouteCollector.php
│   ├── Constant/
│   │   └── Regex.php
│   ├── Controller/
│   │   ├── ApiController.php
│   │   └── Controller.php
│   ├── Data/
│   │   ├── Contract/
│   │   │   ├── DynamicRouteContract.php
│   │   │   ├── ParameterContract.php
│   │   │   └── RouteContract.php
│   │   ├── DynamicRoute.php
│   │   ├── HttpRoutingData.php
│   │   ├── Parameter.php
│   │   └── Route.php
│   ├── Dispatcher/
│   │   ├── Contract/
│   │   │   └── RouterContract.php
│   │   └── Router.php
│   ├── Factory/
│   │   ├── Contract/
│   │   │   └── RoutingResponseFactoryContract.php
│   │   ├── RouteFactory.php
│   │   └── RoutingResponseFactory.php
│   ├── Matcher/
│   │   ├── Contract/
│   │   │   └── MatcherContract.php
│   │   └── Matcher.php
│   ├── Processor/
│   │   ├── Contract/
│   │   │   └── ProcessorContract.php
│   │   └── Processor.php
│   ├── Provider/
│   │   ├── Contract/
│   │   │   └── HttpRouteProviderContract.php
│   │   ├── HttpRoutingCliComponentProvider.php
│   │   ├── HttpRoutingCliRouteProvider.php
│   │   ├── HttpRoutingCliServiceProvider.php
│   │   ├── HttpRoutingComponentProvider.php
│   │   └── HttpRoutingServiceProvider.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── HttpRoutingThrowable.php
│   │   └── Exception/
│   │       ├── Abstract/
│   │       │   ├── HttpRoutingInvalidArgumentException.php
│   │       │   └── HttpRoutingRuntimeException.php
│   │       ├── HttpRoutingInvalidDynamicRouteNameException.php
│   │       ├── HttpRoutingInvalidMethodTypeException.php
│   │       ├── HttpRoutingInvalidParameterRegexException.php
│   │       ├── HttpRoutingInvalidRouteNameException.php
│   │       ├── HttpRoutingInvalidRouteParameterException.php
│   │       ├── HttpRoutingInvalidRoutePathException.php
│   │       ├── HttpRoutingInvalidRouteRegexException.php
│   │       ├── HttpRoutingNoCastException.php
│   │       ├── HttpRoutingNoRequestStructException.php
│   │       └── HttpRoutingNoResponseStructException.php
│   └── Url/
│       ├── Contract/
│       │   └── UrlContract.php
│       └── Url.php
├── Server/
│   ├── Data/
│   │   ├── Contract/
│   │   │   └── HttpServerConfigContract.php
│   │   └── HttpServerConfig.php
│   ├── Handler/
│   │   ├── Contract/
│   │   │   ├── ExceptionResponseRequestHandlerContract.php
│   │   │   └── RequestHandlerContract.php
│   │   └── RequestHandler.php
│   ├── Middleware/
│   │   ├── RequestReceived/
│   │   │   └── RedirectTrailingSlashMiddleware.php
│   │   ├── RouteMatched/
│   │   │   ├── RequestStructMiddleware.php
│   │   │   └── ResponseStructMiddleware.php
│   │   ├── RouteNotMatched/
│   │   │   └── ViewRouteNotMatchedMiddleware.php
│   │   ├── SendingResponse/
│   │   │   └── NoCacheResponseMiddleware.php
│   │   ├── ThrowableCaught/
│   │   │   ├── LogThrowableCaughtMiddleware.php
│   │   │   └── ViewThrowableCaughtMiddleware.php
│   │   └── CacheResponseMiddleware.php
│   ├── Provider/
│   │   ├── HttpServerComponentProvider.php
│   │   └── HttpServerServiceProvider.php
│   ├── Psr/
│   │   └── RequestHandler.php
│   └── Throwable/
│       ├── Contract/
│       │   └── HttpServerThrowable.php
│       └── Exception/
│           └── Abstract/
│               ├── HttpServerInvalidArgumentException.php
│               └── HttpServerRuntimeException.php
├── Struct/
│   ├── Contract/
│   │   └── StructContract.php
│   ├── Request/
│   │   ├── Contract/
│   │   │   └── RequestStructContract.php
│   │   └── Trait/
│   │       ├── JsonRequestStruct.php
│   │       ├── ParsedBodyRequestStruct.php
│   │       ├── QueryRequestStruct.php
│   │       └── RequestStruct.php
│   ├── Response/
│   │   ├── Contract/
│   │   │   └── ResponseStructContract.php
│   │   └── Trait/
│   │       └── ResponseStruct.php
│   └── Throwable/
│       ├── Contract/
│       │   └── HttpStructThrowable.php
│       └── Exception/
│           ├── Abstract/
│           │   ├── HttpStructInvalidArgumentException.php
│           │   └── HttpStructRuntimeException.php
│           └── HttpStructJsonServerRequestExpectedException.php
├── Throwable/
│   ├── Contract/
│   │   └── HttpThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── HttpInvalidArgumentException.php
│           └── HttpRuntimeException.php
└── README.md
```

### 13. JWT Module (`Jwt/`)

_18 files, 11 directories_

```
├── Data/
│   ├── Contract/
│   │   ├── JwtConfigContract.php
│   │   ├── JwtEdDsaConfigContract.php
│   │   ├── JwtHsConfigContract.php
│   │   └── JwtRsConfigContract.php
│   ├── JwtConfig.php
│   ├── JwtEdDsaConfig.php
│   ├── JwtHsConfig.php
│   └── JwtRsConfig.php
├── Enum/
│   └── Algorithm.php
├── Manager/
│   ├── Contract/
│   │   └── JwtContract.php
│   ├── FirebaseJwt.php
│   └── NullJwt.php
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

### 14. Log Module (`Log/`)

_14 files, 12 directories_

```
├── Data/
│   ├── Contract/
│   │   └── LogConfigContract.php
│   └── LogConfig.php
├── Enum/
│   └── LogLevel.php
├── Logger/
│   ├── Abstract/
│   │   └── Logger.php
│   ├── Contract/
│   │   └── LoggerContract.php
│   ├── NullLogger.php
│   └── PsrLogger.php
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

### 15. Mail Module (`Mail/`)

_23 files, 10 directories_

```
├── Data/
│   ├── Contract/
│   │   ├── AttachmentContract.php
│   │   ├── MailConfigContract.php
│   │   ├── MailMailgunConfigContract.php
│   │   ├── MailPhpMailerConfigContract.php
│   │   ├── MessageContract.php
│   │   └── RecipientContract.php
│   ├── Attachment.php
│   ├── MailConfig.php
│   ├── MailMailgunConfig.php
│   ├── MailPhpMailerConfig.php
│   ├── Message.php
│   └── Recipient.php
├── Mailer/
│   ├── Contract/
│   │   └── MailerContract.php
│   ├── LogMailer.php
│   ├── MailgunMailer.php
│   ├── NullMailer.php
│   └── PhpMailer.php
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

### 16. ORM Module (`Orm/`)

_101 files, 35 directories_

```
├── Constant/
│   ├── DateFormat.php
│   └── Statement.php
├── Data/
│   ├── Contract/
│   │   ├── OrmConfigContract.php
│   │   ├── OrmMysqlConfigContract.php
│   │   ├── OrmPgsqlConfigContract.php
│   │   └── OrmSqliteConfigContract.php
│   ├── Join/
│   │   ├── FullOuterJoin.php
│   │   ├── InnerJoin.php
│   │   ├── LeftJoin.php
│   │   ├── OuterJoin.php
│   │   └── RightJoin.php
│   ├── Where/
│   │   ├── AndNotWhere.php
│   │   ├── AndWhere.php
│   │   ├── NotWhere.php
│   │   ├── OrNotWhere.php
│   │   └── OrWhere.php
│   ├── DatedMetadata.php
│   ├── EntityCast.php
│   ├── EntityMetadata.php
│   ├── Id.php
│   ├── Join.php
│   ├── OrderBy.php
│   ├── OrmConfig.php
│   ├── OrmMysqlConfig.php
│   ├── OrmPgsqlConfig.php
│   ├── OrmSqliteConfig.php
│   ├── SoftDeleteMetadata.php
│   ├── Value.php
│   ├── Where.php
│   └── WhereGroup.php
├── Entity/
│   ├── Abstract/
│   │   ├── DatedEntity.php
│   │   ├── Entity.php
│   │   └── SoftDeleteEntity.php
│   ├── Contract/
│   │   ├── DatedEntityContract.php
│   │   ├── EntityContract.php
│   │   └── SoftDeleteEntityContract.php
│   └── Trait/
│       ├── DatedFields.php
│       └── SoftDeleteFields.php
├── Enum/
│   ├── Comparison.php
│   ├── JoinOperator.php
│   ├── JoinType.php
│   ├── SortOrder.php
│   └── WhereType.php
├── Factory/
│   └── DateFactory.php
├── Manager/
│   ├── Abstract/
│   │   └── PdoManager.php
│   ├── Contract/
│   │   └── ManagerContract.php
│   ├── MysqlManager.php
│   ├── NullManager.php
│   ├── PgsqlManager.php
│   └── SqliteManager.php
├── Middleware/
│   └── EntityRouteMatchedMiddleware.php
├── Provider/
│   ├── OrmComponentProvider.php
│   └── OrmServiceProvider.php
├── QueryBuilder/
│   ├── Abstract/
│   │   └── SqlQueryBuilder.php
│   ├── Contract/
│   │   ├── DeleteQueryBuilderContract.php
│   │   ├── InsertQueryBuilderContract.php
│   │   ├── QueryBuilderContract.php
│   │   ├── SelectQueryBuilderContract.php
│   │   └── UpdateQueryBuilderContract.php
│   ├── Factory/
│   │   ├── Contract/
│   │   │   └── QueryBuilderFactoryContract.php
│   │   └── SqlQueryBuilderFactory.php
│   ├── SqlDeleteQueryBuilder.php
│   ├── SqlInsertQueryBuilder.php
│   ├── SqlSelectQueryBuilder.php
│   └── SqlUpdateQueryBuilder.php
├── Registry/
│   ├── Contract/
│   │   └── EntityMetadataRegistryContract.php
│   └── EntityMetadataRegistry.php
├── Repository/
│   ├── Contract/
│   │   └── RepositoryContract.php
│   └── Repository.php
├── Schema/
│   ├── Abstract/
│   │   ├── Migration.php
│   │   ├── SqlFileMigration.php
│   │   └── TransactionalMigration.php
│   └── Contract/
│       ├── ColumnContract.php
│       ├── ConstraintContract.php
│       ├── IndexContract.php
│       ├── MigrationContract.php
│       ├── SchemaContract.php
│       └── TableContract.php
├── Statement/
│   ├── Contract/
│   │   └── StatementContract.php
│   ├── NullStatement.php
│   └── PdoStatement.php
├── Throwable/
│   ├── Contract/
│   │   └── OrmThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── OrmInvalidArgumentException.php
│       │   └── OrmRuntimeException.php
│       ├── OrmArrayCastingException.php
│       ├── OrmEntityNotFoundException.php
│       ├── OrmExecuteException.php
│       ├── OrmFetchException.php
│       ├── OrmInvalidColumnNumberException.php
│       ├── OrmInvalidEntityException.php
│       ├── OrmInvalidMigrationFileException.php
│       ├── OrmMigrationExecutionException.php
│       ├── OrmNoLastIdException.php
│       ├── OrmNoPgsqlLastIdException.php
│       ├── OrmNotFoundException.php
│       ├── OrmStatementPreparationFailureException.php
│       ├── OrmUnexpectedIdValueException.php
│       ├── OrmUnregisteredEntityException.php
│       ├── OrmUnsupportedCountException.php
│       └── OrmWhereException.php
└── README.md
```

### 17. Reflection Module (`Reflection/`)

_9 files, 8 directories_

```
├── Provider/
│   ├── ReflectionComponentProvider.php
│   └── ReflectionServiceProvider.php
├── Reflector/
│   ├── Contract/
│   │   └── ReflectorContract.php
│   └── Reflector.php
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

### 18. Session Module (`Session/`)

_35 files, 18 directories_

```
├── Data/
│   ├── Contract/
│   │   ├── SessionConfigContract.php
│   │   ├── SessionCookieConfigContract.php
│   │   ├── SessionJwtConfigContract.php
│   │   └── SessionTokenConfigContract.php
│   ├── SessionConfig.php
│   ├── SessionCookieConfig.php
│   ├── SessionJwtConfig.php
│   └── SessionTokenConfig.php
├── Manager/
│   ├── Abstract/
│   │   └── Session.php
│   ├── Contract/
│   │   └── SessionContract.php
│   ├── Cookie/
│   │   ├── CookieSession.php
│   │   └── EncryptedCookieSession.php
│   ├── Jwt/
│   │   ├── Cli/
│   │   │   ├── EncryptedOptionJwtSession.php
│   │   │   └── OptionJwtSession.php
│   │   └── Http/
│   │       ├── EncryptedHeaderJwtSession.php
│   │       └── HeaderJwtSession.php
│   ├── Token/
│   │   ├── Cli/
│   │   │   ├── EncryptedOptionTokenSession.php
│   │   │   └── OptionTokenSession.php
│   │   └── Http/
│   │       ├── EncryptedHeaderTokenSession.php
│   │       └── HeaderTokenSession.php
│   ├── CacheSession.php
│   ├── LogSession.php
│   ├── NullSession.php
│   └── PhpSession.php
├── Provider/
│   ├── SessionComponentProvider.php
│   └── SessionServiceProvider.php
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

### 19. SMS Module (`Sms/`)

_16 files, 10 directories_

```
├── Data/
│   ├── Contract/
│   │   ├── MessageContract.php
│   │   ├── SmsConfigContract.php
│   │   └── SmsVonageConfigContract.php
│   ├── Message.php
│   ├── SmsConfig.php
│   └── SmsVonageConfig.php
├── Messenger/
│   ├── Contract/
│   │   └── MessengerContract.php
│   ├── LogMessenger.php
│   ├── NullMessenger.php
│   └── VonageMessenger.php
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

### 20. Support Module (`Support/`)

_3 files, 2 directories_

```
├── Time/
│   ├── Microtime.php
│   └── Time.php
└── README.md
```

### 21. Throwable Module (`Throwable/`)

_7 files, 7 directories_

```
├── Contract/
│   └── ValkyrjaThrowable.php
├── Exception/
│   └── Abstract/
│       ├── ValkyrjaInvalidArgumentException.php
│       └── ValkyrjaRuntimeException.php
├── Factory/
│   └── ThrowableFactory.php
├── Handler/
│   ├── Contract/
│   │   └── ThrowableHandlerContract.php
│   └── WhoopsThrowableHandler.php
└── README.md
```

### 22. Type Module (`Type/`)

_168 files, 83 directories_

```
├── Abstract/
│   └── Type.php
├── Array/
│   ├── Contract/
│   │   └── ArrayContract.php
│   ├── Factory/
│   │   └── ArrayFactory.php
│   ├── Support/
│   │   └── ArrayOf.php
│   ├── Throwable/
│   │   └── Exception/
│   │       ├── ArrayInvalidEncodedArrayException.php
│   │       ├── ArrayInvalidFalseValueException.php
│   │       ├── ArrayInvalidNonEmptyException.php
│   │       ├── ArrayInvalidNullValueException.php
│   │       ├── ArrayInvalidStringKeysException.php
│   │       └── ArrayInvalidTrueValueException.php
│   ├── ArrayT.php
│   └── NonEmptyArray.php
├── Bool/
│   ├── Contract/
│   │   ├── BoolContract.php
│   │   ├── FalseContract.php
│   │   └── TrueContract.php
│   ├── BoolT.php
│   ├── FalseT.php
│   └── TrueT.php
├── Collection/
│   ├── Contract/
│   │   └── CollectionContract.php
│   └── Collection.php
├── Contract/
│   └── TypeContract.php
├── Data/
│   ├── ArrayCast.php
│   ├── Cast.php
│   ├── OriginalArrayCast.php
│   └── OriginalCast.php
├── Enum/
│   ├── Contract/
│   │   ├── ArrayableContract.php
│   │   ├── BackedEnumContract.php
│   │   ├── EnumContract.php
│   │   └── JsonSerializableContract.php
│   ├── Support/
│   │   └── Enumerable.php
│   ├── Throwable/
│   │   └── Exception/
│   │       ├── EnumCannotModifyException.php
│   │       └── EnumInvalidValueException.php
│   ├── Trait/
│   │   ├── Arrayable.php
│   │   ├── Enumerable.php
│   │   └── JsonSerializable.php
│   ├── CastType.php
│   └── Type.php
├── Float/
│   ├── Contract/
│   │   └── FloatContract.php
│   ├── Throwable/
│   │   └── Exception/
│   │       └── FloatInvalidFromValueException.php
│   └── FloatT.php
├── Id/
│   ├── Contract/
│   │   ├── IdContract.php
│   │   ├── IntIdContract.php
│   │   └── StringIdContract.php
│   ├── Throwable/
│   │   └── Exception/
│   │       └── IdInvalidFromValueException.php
│   ├── Id.php
│   ├── IntId.php
│   └── StringId.php
├── Int/
│   ├── Contract/
│   │   └── IntContract.php
│   ├── Throwable/
│   │   └── Exception/
│   │       └── IntInvalidFromValueException.php
│   └── IntT.php
├── Json/
│   ├── Contract/
│   │   ├── JsonContract.php
│   │   └── JsonObjectContract.php
│   ├── Json.php
│   └── JsonObject.php
├── Model/
│   ├── Abstract/
│   │   ├── CastableModel.php
│   │   ├── IndexedModel.php
│   │   └── Model.php
│   ├── Contract/
│   │   ├── CastableModelContract.php
│   │   ├── ExposableIndexedModelContract.php
│   │   ├── ExposableModelContract.php
│   │   ├── IndexedModelContract.php
│   │   └── ModelContract.php
│   ├── Trait/
│   │   ├── Castable.php
│   │   ├── Exposable.php
│   │   ├── ExposableIndexable.php
│   │   ├── Indexable.php
│   │   ├── ProtectedExposable.php
│   │   └── UnpackForNewInstance.php
│   └── README.md
├── Null/
│   ├── Contract/
│   │   └── NullContract.php
│   └── NullT.php
├── Object/
│   ├── Contract/
│   │   ├── ObjectContract.php
│   │   └── SerializedObjectContract.php
│   ├── Enum/
│   │   └── PropertyVisibilityFilter.php
│   ├── Factory/
│   │   └── ObjectFactory.php
│   ├── Support/
│   │   └── Cls.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── ObjectThrowable.php
│   │   └── Exception/
│   │       ├── InvalidEncodedObjectException.php
│   │       ├── InvalidObjectPropertyProvidedException.php
│   │       ├── InvalidObjectProvidedException.php
│   │       └── InvalidSerializedObjectException.php
│   ├── ObjectT.php
│   └── SerializedObject.php
├── String/
│   ├── Contract/
│   │   └── StringContract.php
│   ├── Factory/
│   │   ├── MbStringFactory.php
│   │   ├── StringCaseFactory.php
│   │   └── StringFactory.php
│   ├── Throwable/
│   │   └── Exception/
│   │       └── StringInvalidEmptyStringException.php
│   ├── NonEmptyString.php
│   └── StringT.php
├── Throwable/
│   ├── Contract/
│   │   └── TypeThrowable.php
│   └── Exception/
│       └── Abstract/
│           ├── TypeInvalidArgumentException.php
│           └── TypeRuntimeException.php
├── Uid/
│   ├── Contract/
│   │   └── UidContract.php
│   ├── Factory/
│   │   └── UidFactory.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── UidThrowable.php
│   │   └── Exception/
│   │       ├── InvalidUidException.php
│   │       └── UidInvalidFromValueException.php
│   └── Uid.php
├── Ulid/
│   ├── Contract/
│   │   └── UlidContract.php
│   ├── Factory/
│   │   └── UlidFactory.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── UlidThrowable.php
│   │   └── Exception/
│   │       ├── InvalidUlidException.php
│   │       ├── UlidInvalidFromValueException.php
│   │       └── UlidRandomBytesFailureException.php
│   ├── README.md
│   └── Ulid.php
├── Uuid/
│   ├── Contract/
│   │   ├── UuidContract.php
│   │   ├── UuidV1Contract.php
│   │   ├── UuidV3Contract.php
│   │   ├── UuidV4Contract.php
│   │   ├── UuidV5Contract.php
│   │   ├── UuidV6Contract.php
│   │   ├── UuidV7Contract.php
│   │   └── UuidV8Contract.php
│   ├── Enum/
│   │   └── Version.php
│   ├── Factory/
│   │   ├── UuidFactory.php
│   │   ├── UuidV1Factory.php
│   │   ├── UuidV3Factory.php
│   │   ├── UuidV4Factory.php
│   │   ├── UuidV5Factory.php
│   │   ├── UuidV6Factory.php
│   │   ├── UuidV7Factory.php
│   │   └── UuidV8Factory.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── UuidThrowable.php
│   │   └── Exception/
│   │       ├── InvalidUuidException.php
│   │       ├── InvalidUuidV1Exception.php
│   │       ├── InvalidUuidV3Exception.php
│   │       ├── InvalidUuidV4Exception.php
│   │       ├── InvalidUuidV5Exception.php
│   │       ├── InvalidUuidV6Exception.php
│   │       ├── InvalidUuidV7Exception.php
│   │       ├── InvalidUuidV8Exception.php
│   │       └── UuidInvalidFromValueException.php
│   ├── README.md
│   ├── Uuid.php
│   ├── UuidV1.php
│   ├── UuidV3.php
│   ├── UuidV4.php
│   ├── UuidV5.php
│   ├── UuidV6.php
│   ├── UuidV7.php
│   └── UuidV8.php
├── Vlid/
│   ├── Contract/
│   │   ├── VlidContract.php
│   │   ├── VlidV1Contract.php
│   │   ├── VlidV2Contract.php
│   │   ├── VlidV3Contract.php
│   │   └── VlidV4Contract.php
│   ├── Enum/
│   │   └── Version.php
│   ├── Factory/
│   │   ├── VlidFactory.php
│   │   ├── VlidV1Factory.php
│   │   ├── VlidV2Factory.php
│   │   ├── VlidV3Factory.php
│   │   └── VlidV4Factory.php
│   ├── Throwable/
│   │   ├── Contract/
│   │   │   └── VlidThrowable.php
│   │   └── Exception/
│   │       ├── InvalidVlidException.php
│   │       ├── InvalidVlidV1Exception.php
│   │       ├── InvalidVlidV2Exception.php
│   │       ├── InvalidVlidV3Exception.php
│   │       ├── InvalidVlidV4Exception.php
│   │       └── VlidInvalidFromValueException.php
│   ├── README.md
│   ├── Vlid.php
│   ├── VlidV1.php
│   ├── VlidV2.php
│   ├── VlidV3.php
│   └── VlidV4.php
└── README.md
```

### 23. Validation Module (`Validation/`)

_33 files, 16 directories_

```
├── Constant/
│   └── ErrorMessage.php
├── Rule/
│   ├── Abstract/
│   │   └── Rule.php
│   ├── Contract/
│   │   └── RuleContract.php
│   ├── Int/
│   │   ├── GreaterThan.php
│   │   └── LessThan.php
│   ├── Is/
│   │   ├── Email.php
│   │   ├── Equal.php
│   │   ├── IsBool.php
│   │   ├── IsEmpty.php
│   │   ├── IsNumeric.php
│   │   ├── IsString.php
│   │   ├── NotEmpty.php
│   │   ├── NotEqual.php
│   │   └── Required.php
│   ├── Orm/
│   │   ├── Abstract/
│   │   │   └── EntityRule.php
│   │   ├── EntityExists.php
│   │   └── EntityNotExists.php
│   └── String/
│       ├── Alpha.php
│       ├── Contains.php
│       ├── EndsWith.php
│       ├── Lowercase.php
│       ├── Max.php
│       ├── Min.php
│       ├── Regex.php
│       ├── StartsWith.php
│       └── Uppercase.php
├── Throwable/
│   ├── Contract/
│   │   └── ValidationThrowable.php
│   └── Exception/
│       ├── Abstract/
│       │   ├── ValidationInvalidArgumentException.php
│       │   └── ValidationRuntimeException.php
│       └── ValidationRuleFailureException.php
├── Validator/
│   ├── Contract/
│   │   └── ValidatorContract.php
│   └── Validator.php
└── README.md
```

### 24. View Module (`View/`)

_68 files, 28 directories_

```
├── Data/
│   ├── Contract/
│   │   ├── ViewConfigContract.php
│   │   ├── ViewOrkaConfigContract.php
│   │   ├── ViewPhpConfigContract.php
│   │   └── ViewTwigConfigContract.php
│   ├── ViewConfig.php
│   ├── ViewOrkaConfig.php
│   ├── ViewPhpConfig.php
│   └── ViewTwigConfig.php
├── Factory/
│   ├── Contract/
│   │   └── ViewResponseFactoryContract.php
│   └── ViewResponseFactory.php
├── Orka/
│   ├── Constant/
│   │   ├── OrkaReplacement.php
│   │   └── OrkaReplacementCollection.php
│   └── Replacement/
│       ├── Block/
│       │   ├── Block.php
│       │   ├── EndBlock.php
│       │   ├── StartBlock.php
│       │   └── TrimBlock.php
│       ├── Comment/
│       │   ├── EndMultiline.php
│       │   ├── SingleLine.php
│       │   └── StartMultiline.php
│       ├── Contract/
│       │   └── ReplacementContract.php
│       ├── Debug/
│       │   └── Dd.php
│       ├── Partial/
│       │   ├── Partial.php
│       │   ├── PartialWithVariables.php
│       │   ├── TrimPartial.php
│       │   └── TrimPartialWithVariables.php
│       ├── Statement/
│       │   ├── Conditional/
│       │   │   ├── Block/
│       │   │   │   ├── ElseHasBlock.php
│       │   │   │   ├── HasBlock.php
│       │   │   │   └── UnlessBlock.php
│       │   │   ├── ElseIf_.php
│       │   │   ├── ElseUnless.php
│       │   │   ├── Else_.php
│       │   │   ├── Empty_.php
│       │   │   ├── EndIf_.php
│       │   │   ├── If_.php
│       │   │   ├── Isset_.php
│       │   │   ├── NotEmpty.php
│       │   │   └── Unless.php
│       │   ├── Iterate/
│       │   │   ├── EndFor_.php
│       │   │   ├── EndForeach_.php
│       │   │   ├── For_.php
│       │   │   └── Foreach_.php
│       │   ├── Switch/
│       │   │   ├── Case_.php
│       │   │   ├── Default_.php
│       │   │   ├── EndSwitch_.php
│       │   │   └── Switch_.php
│       │   └── Break_.php
│       ├── Variable/
│       │   ├── Escaped.php
│       │   ├── SetVariable.php
│       │   ├── SetVariables.php
│       │   └── Unescaped.php
│       └── Layout.php
├── Provider/
│   ├── ViewComponentProvider.php
│   ├── ViewOrkaServiceProvider.php
│   └── ViewServiceProvider.php
├── Renderer/
│   ├── Contract/
│   │   └── RendererContract.php
│   ├── OrkaRenderer.php
│   ├── PhpRenderer.php
│   └── TwigRenderer.php
├── Template/
│   ├── Contract/
│   │   └── TemplateContract.php
│   └── Template.php
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
│       └── ViewRenderFailureException.php
└── README.md
```

---

## Documentation Files at the Base Path

These five files sit directly under `src/Valkyrja`. They belong to no module,
and the 24 module counts above do not count them.

```
├── APPLICATION_STRUCTURE.md
├── GETTING_STARTED.md
├── LIFECYCLE.md
├── README.md
└── VERSIONING_AND_RELEASE_PROCESS.md
```

---

## Key Architectural Patterns

### 1. Manager Pattern

Some modules have `Manager/` subdirectories with:

- One or more implementations (e.g., `RedisCache`, `LogCache`, `NullCache`)
- A contract in `Manager/Contract/`

### 2. Service Providers

Most modules have a `Provider/` directory. A `Provider/` directory holds:

- A `<Module>ComponentProvider` (dependency injection setup)
- A `<Module>ServiceProvider`, except in the Application module (service
  registration)
- A `Contract/` directory in five of them: the Application, Container, and Event
  modules, and the Cli and Http routing subtrees

### 3. Exception Handling

Structured exception hierarchy:

- A `<Module>InvalidArgumentException` and a `<Module>RuntimeException` in most
  modules, in the module's own `Throwable/Exception/Abstract/` directory
- A `ValkyrjaInvalidArgumentException` and a `ValkyrjaRuntimeException` in the
  Throwable module's `Exception/Abstract/` directory, which each module's pair
  extends
- Concrete exceptions directly in `Throwable/Exception/`, in 11 of the 24
  modules (`CryptDecodeFailureException`, `ViewInvalidPathException`)
- A `<Module>Throwable` contract in most modules (`ApiThrowable`,
  `SessionThrowable`)
- A `ValkyrjaThrowable` contract in the Throwable module's `Contract/`
  directory, which every `<Module>Throwable` extends

### 4. Type System

Extensive Type module with:

- Basic types (`Bool`, `Int`, `Float`, `String`, `Null`)
- Collection types (`Array`, `Collection`)
- Unique ID types (`Uid`, `Uuid`, `Ulid`, `Vlid`)
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
