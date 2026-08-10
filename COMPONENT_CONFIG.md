# COMPONENT_CONFIG.md — the component config shape

The **cross-language** convention for a component's configuration. It applies
in every Valkyrja port, and to every contributor, human or agent. PHP spellings
are shown; each Layer-2 guide gives the per-language spelling.

A component gets one `ComponentNameConfigContract` for the settings that apply
to the whole component. The default adapter is the most common such setting.
Each adapter then gets its own `ComponentName<Adapter>ConfigContract`. Every
contract has a default implementation that drops the `Contract` suffix
(`CacheConfig`, `CacheRedisConfig`).

The default implementations live in the component's `Data\` segment. The
contracts live in the component's `Data\Contract\` segment. The component's
service provider publishes each contract as its own container binding.

## The two rules

1. **The component config does not hold the adapter configs.** The container
   resolves an adapter config only when something asks for that adapter. An
   application that uses one cache adapter never constructs the configuration
   for the other cache adapters.
2. **An adapter contract prefixes every property with the adapter name.** One
   application config class can implement several adapter contracts at once.
   Without the prefix, two adapters that both declare a `prefix` property
   collide.

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
contract, the service provider binds the default implementation:

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
