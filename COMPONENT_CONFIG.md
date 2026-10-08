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

Queue's client subcomponent shows both shapes.
`Valkyrja\Queue\Client\Data\Contract\QueueClientConfigContract` configures
the subcomponent, and
`Valkyrja\Queue\Client\Data\Contract\QueueRedisClientConfigContract`
configures one adapter of it. An application names its default client through
the subcomponent contract, and
`Valkyrja\Application\Data\Contract\QueueConfigContract` is a separate
thing: it configures a queue application rather than naming a client.

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
3. **When two components would declare the same property for the same
   adapter, the component whose domain the adapter belongs to keeps the bare
   name.** Every other one carries its own component name. A PSR logger is a
   logging thing, so `LogPsrConfigContract` holds `$psrName`. Every component
   that logs borrows the log adapter instead, and no component owns it, so
   each one carries its own name: `$cacheLogLogger`, `$mailLogLogger`. Redis
   is a cache, so `CacheRedisConfigContract` keeps `$redisHost` and a queue
   that runs jobs through redis carries its own name.

   An adapter that one component configures alone never reaches this rule,
   because nothing can collide with it: `$mailgunDomain` exists only in Mail,
   and `$amqpHost` only in Queue. Two components can also configure one
   adapter and still collide on nothing, as `ViewPhpConfigContract`'s
   `$phpPath` and `SessionPhpConfigContract`'s `$phpCookiePath` do.

4. **A subcomponent name counts as the adapter name for rule 3.** `Client` is
   a subcomponent of Http and of Queue alike, so an application config that
   holds both reaches rule 3 and each property carries its component name.
   An adapter of a subcomponent carries both names, in the order its own
   contract spells them: `HttpClientLogConfigContract` holds
   `$httpClientLogLogger`.

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
