# COMPONENT_CONFIG.md — the component config shape

The **cross-language** convention for a component's configuration. It applies
in every Valkyrja port, and to every contributor, human or agent. PHP spellings
are shown; each Layer-2 guide gives the per-language spelling.

A component gets one `ComponentNameConfigContract` for the settings that apply
to the whole component. The default adapter is the most common such setting.
Each adapter then gets its own `ComponentName<Adapter>ConfigContract`. A
contract whose settings have a usable default gets a default implementation
that drops the `Contract` suffix (`CacheConfig`, `CacheRedisConfig`). A
contract whose settings have no usable default gets none, and the service
provider throws instead of binding one.

A component whose subcomponents configure separately gets one contract for each
of them, named `ComponentName<SubComponent>ConfigContract`. An adapter of that
subcomponent carries an adapter name as well, so the extra name is what tells
the two apart.

Queue is the case.
`Valkyrja\Queue\Client\Data\Contract\QueueClientConfigContract` holds
`$defaultQueueClient`, and
`Valkyrja\Queue\Client\Data\Contract\QueueRedisClientConfigContract` holds
the settings of the Redis client. A host application holds the first to name
its default client. A host never holds
`Valkyrja\Application\Data\Contract\QueueConfigContract`, which configures a
queue application instead.

```php
// Right — the name says what it configures. `QueueClientConfig` names the
// component and the subcomponent, so it configures the subcomponent.
// `QueueRedisClientConfig` names an adapter too, so it configures one adapter
// of that subcomponent.
use Valkyrja\Queue\Client\Data\Contract\QueueClientConfigContract;
use Valkyrja\Queue\Client\Data\Contract\QueueRedisClientConfigContract;
```

A contract lives in the `Data\Contract\` segment of whatever it configures, and
a default implementation lives in the `Data\` segment beside it. A component's
contracts therefore sit in the component, and a subcomponent's sit in the
subcomponent, that subcomponent's adapter contracts included. The component's
service provider publishes each contract as its own container binding.

## The rules

1. **The component config does not hold the adapter configs.** The container
   resolves an adapter config only when something asks for that adapter. An
   application that uses one cache adapter never constructs the configuration
   for the other cache adapters.
2. **An adapter contract prefixes every property with the adapter name.** One
   application config class can implement several adapter contracts at once.
   Without the prefix, two adapters that both declare a `prefix` property
   collide.
3. **A component that configures an adapter outside that adapter's own
   domain carries its component name in every property.** The component
   whose domain the adapter belongs to keeps the bare adapter prefix. A PSR
   logger is a logging thing, so `LogPsrConfigContract` holds `$psrName`,
   and every other component logs through the log adapter under
   `$cacheLogLogger` or `$mailLogLogger`. Redis is a cache, so
   `CacheRedisConfigContract` holds `$redisHost`, and the Queue client that
   runs jobs through Redis holds `$queueRedisClientHost`. An adapter that one
   component configures alone needs no component name, because nothing can
   collide with it: `$mailgunDomain` exists only in Mail.
4. **A subcomponent contract carries the subcomponent name in every property.**
   `QueueClientConfigContract` holds `$defaultQueueClient`, where the `default`
   property names the component as well, as a component config's own
   `$defaultCache` does. This governs the subcomponent's own contract. An
   adapter of that subcomponent carries the subcomponent name as well, in the
   order its own contract name spells it: `QueueRedisClientConfigContract`
   holds `$queueRedisClientHost`, and `HttpClientLogConfigContract` holds
   `$httpClientLogLogger`. The property prefix is the contract name without
   its `ConfigContract` suffix, which satisfies rules 2 and 3 at once.

```php
// Wrong — the component config holds every adapter config. An application that
// uses only the null cache still constructs the redis and the log configuration.
interface CacheConfigContract
{
    public string $defaultCache { get; }
    public CacheRedisConfig $redisCache { get; }
    public CacheLogConfig $logCache { get; }
    public CacheNullConfig $nullCache { get; }
}
```

```php
// Right — the component config holds the component-wide setting only.
interface CacheConfigContract
{
    /** @var class-string<CacheContract> */
    public string $defaultCache { get; }
}
```

```php
// Right — each adapter has its own contract, and each property carries the
// adapter prefix, so one class can implement several contracts at once.
interface CacheRedisConfigContract
{
    public string $redisHost { get; }
    public int $redisPort { get; }
    public string $redisPrefix { get; }
}

interface CacheNullConfigContract
{
    public string $nullPrefix { get; }
}
```

```php
// Right — the log adapter of each component names its own logger property, so
// one application config class sets a different logger for each component.
interface CacheLogConfigContract
{
    /** @var class-string<LoggerContract> */
    public string $cacheLogLogger { get; }
}

interface MailLogConfigContract
{
    /** @var class-string<LoggerContract> */
    public string $mailLogLogger { get; }
}
```

## The application implements only what it uses

One application config class implements the component contract, and it adds an
adapter contract only when the application uses that adapter:

```php
final class AppConfig extends Config implements CacheConfigContract, CacheRedisConfigContract
{
    public function __construct(
        public string $defaultCache = RedisCache::class,
        public string $redisHost = 'cache.internal',
        public int $redisPort = 6379,
        public string $redisPrefix = 'app:',
    ) {
        parent::__construct();
    }
}
```

## The service provider binds by contract

The service provider binds the application config when the application config
implements the contract. When the application config does not implement the
contract, the service provider binds the default implementation, or throws when
the contract has none:

```php
public static function publishRedisConfig(ContainerContract $container): void
{
    $config = $container->getSingleton(ConfigContract::class);

    if ($config instanceof CacheRedisConfigContract) {
        $container->setSingleton(CacheRedisConfigContract::class, $config);

        return;
    }

    $container->setSingleton(CacheRedisConfigContract::class, new CacheRedisConfig());
}
```
