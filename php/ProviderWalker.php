<?php

declare(strict_types=1);

/*
 * This file is part of the Valkyrja Framework package.
 *
 * (c) Melech Mizrachi <melechmizrachi@gmail.com>
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

namespace Valkyrja\Volundr\Analysis;

use Composer\Autoload\ClassLoader;
use PhpParser\Node;
use PhpParser\Node\Expr\Array_;
use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\Node\Name;
use PhpParser\Node\Stmt\Class_;
use PhpParser\Node\Stmt\ClassMethod;
use PhpParser\Node\Stmt\Interface_;
use PhpParser\NodeFinder;
use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;
use RuntimeException;

/**
 * Walks a ComponentProviderContract implementation and extracts the FQCNs it
 * returns from each get*Providers method, resolving them to file paths via
 * Composer's autoloader.
 */
final class ProviderWalker
{
    private const METHODS = [
        'componentProviders' => 'getComponentProviders',
        'containerProviders' => 'getContainerProviders',
        'eventProviders'     => 'getEventProviders',
        'cliProviders'       => 'getCliProviders',
        'httpProviders'      => 'getHttpProviders',
    ];

    public function __construct(
        private readonly ClassLoader $loader,
    ) {
    }

    /**
     * @return array{
     *   class: string|null,
     *   componentProviders: list<array{fqcn: string, file: ?string}>,
     *   containerProviders: list<array{fqcn: string, file: ?string}>,
     *   eventProviders:     list<array{fqcn: string, file: ?string}>,
     *   cliProviders:       list<array{fqcn: string, file: ?string}>,
     *   httpProviders:      list<array{fqcn: string, file: ?string}>,
     * }
     */
    public function analyze(string $filePath): array
    {
        $code = file_get_contents($filePath);

        if ($code === false) {
            throw new RuntimeException("Cannot read $filePath");
        }

        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $ast    = $parser->parse($code) ?? [];

        // Run the NameResolver so `use` statements are applied and every Name
        // node has a resolvedName attribute holding its FQCN.
        $traverser = new NodeTraverser();
        $traverser->addVisitor(new NameResolver());
        $ast = $traverser->traverse($ast);

        $finder = new NodeFinder();

        // Find the class or interface declaration in the file. For a provider
        // we expect exactly one top-level class; interfaces are included so
        // this tool also works against the contract itself for testing.
        /** @var Class_|Interface_|null $decl */
        $decl = $finder->findFirst($ast, static fn (Node $n) => $n instanceof Class_ || $n instanceof Interface_);

        $result = ['class' => $decl?->namespacedName?->toString()];

        foreach (self::METHODS as $key => $_) {
            $result[$key] = [];
        }

        if ($decl === null) {
            return $result;
        }

        foreach (self::METHODS as $key => $methodName) {
            $method = $this->findMethod($decl, $methodName);

            if ($method === null) {
                continue;
            }

            foreach ($this->extractClassStrings($method) as $fqcn) {
                $result[$key][] = [
                    'fqcn' => $fqcn,
                    'file' => $this->resolveToFile($fqcn),
                ];
            }
        }

        return $result;
    }

    private function findMethod(Class_|Interface_ $decl, string $name): ClassMethod|null
    {
        foreach ($decl->getMethods() as $method) {
            if (strcasecmp($method->name->toString(), $name) === 0) {
                return $method;
            }
        }

        return null;
    }

    /**
     * Extract every `Foo::class` reference from the return statements of a
     * method. We only care about literal return arrays — anything dynamic is
     * intentionally not supported, which keeps the invariant clean.
     *
     * @return list<string>
     */
    private function extractClassStrings(ClassMethod $method): array
    {
        $fqcns  = [];
        $finder = new NodeFinder();

        /** @var list<Node\Stmt\Return_> $returns */
        $returns = $finder->findInstanceOf($method->stmts ?? [], Node\Stmt\Return_::class);

        foreach ($returns as $return) {
            if (!$return->expr instanceof Array_) {
                continue;
            }

            foreach ($return->expr->items as $item) {
                if ($item === null || !$item->value instanceof ClassConstFetch) {
                    continue;
                }

                // We only care about `SomeClass::class` (not $var::class).
                if (!$item->value->class instanceof Name) {
                    continue;
                }

                if (!$item->value->name instanceof Node\Identifier) {
                    continue;
                }

                if ($item->value->name->toLowerString() !== 'class') {
                    continue;
                }

                // NameResolver has already attached the fully-qualified name.
                $resolved = $item->value->class->getAttribute('resolvedName');

                if ($resolved instanceof Name) {
                    $fqcns[] = $resolved->toString();
                } else {
                    $fqcns[] = $item->value->class->toString();
                }
            }
        }

        return $fqcns;
    }

    /**
     * Delegate to Composer — it already knows how to resolve PSR-4, PSR-0,
     * and classmap entries, so we don't have to reinvent any of it.
     */
    private function resolveToFile(string $fqcn): string|null
    {
        $path = $this->loader->findFile($fqcn);

        return $path === false ? null : realpath($path) ?: $path;
    }
}
