# Sindri — Build Tool

## Overview

`sindri` is a standalone tool that ships with the Valkyrja framework. It generates cache data files for production CGI
and lambda deployments across all language ports. The framework itself has zero AST or build dependency — all source
extraction, analysis, and code generation logic lives exclusively in the build tool.

**sindri** is itself a Valkyrja application. It ships without cache by default since it runs at deploy time rather than
per-request. For environments where build tool startup time is a concern, sindri can generate its own cache and
rebuild — the same two-pass process it performs for any other compiled language application.

The name fits: Sindri was the master smith of Norse mythology who forged Mjölnir and the other legendary artifacts.
Sindri takes raw source files and shapes them into optimised cache artifacts ready for production.

### Namespace and Package Names

Sindri is a peer tool to the Valkyrja framework, not a component within it. Its namespace reflects this — `Sindri`
stands alone rather than nesting under `Valkyrja\*`:

| Language   | Namespace / Package                     |
|------------|-----------------------------------------|
| PHP        | `Sindri`                                |
| Java       | `sindri` (top-level package)            |
| Go         | `sindri` (module: `io/valkyrja/sindri`) |
| Python     | `sindri` (PyPI: `valkyrja-sindri`)      |
| TypeScript | `sindri` (npm: `@valkyrja/sindri`)      |

The package names are namespaced under `valkyrja` at the ecosystem level — clearly Valkyrja's tool. The PHP/Java
namespace and Go/Python/TypeScript module name are just `sindri` — consistent with its identity as a standalone binary
rather than a framework component.

---

## What It Does

The build tool:

1. Reads the application config class to discover top-level component providers
2. Walks the static provider tree via AST to discover all sub-providers
3. Categorizes each leaf provider (container / http route / cli route / event)
4. Extracts all bindings, routes, listeners, and commands across every provider of each type
5. For routes: constructs plain `ValkyrjaRoute` objects, runs `ProcessorContract::route()` to compile regexes,
   pre-builds all dispatcher indexes
6. Resolves all type references to fully-qualified names
7. Writes exactly **four** aggregated data classes for the entire application — one per concern
8. Generated files are compiled/included in the final deployment artifact

The build tool is a **source code generator** — it writes strings that are valid source code. It does not need to
instantiate, extend, or have access to application-defined custom classes at generation time. Class names from AST are
written as text. The compiler/runtime resolves them when the generated files are compiled/loaded alongside the
application.

---

## Application Config — Build Tool Entry Point

The build tool's entry point is the application config class — the same class the developer already uses to configure
the application. No separate `valkyrja.yaml` file is needed. The config class is already the authoritative list of what
the application uses.

```php
// PHP — the config class IS the build tool entry point
$app = Application::create(
    new AppConfig(
        providers: [
            HttpComponentProvider::class,       // framework component providers
            ContainerComponentProvider::class,
            EventComponentProvider::class,
            CliComponentProvider::class,
            App\Providers\AppProvider::class,  // application providers
        ]
    )
);
```

The build tool reads the config class's `providers` list from AST — the same way it reads any other provider list
method. Class references must use `::class` / `.class` / class objects directly — **not constants from a constants class
**. String constant references are not statically resolvable by the build tool.

```
✅ HttpComponentProvider::class     — direct ::class reference, readable
✅ HttpComponentProvider.class      — Java .class reference, readable
✅ HttpComponentProvider            — Python class object, readable

❌ HttpConstants::HTTP_COMPONENT_PROVIDER  — constant reference, not resolvable
❌ getProvider()                           — method call, not resolvable
```

### Component Provider Constants Class — Dropped

A constants class for component provider class references was considered but dropped. It would allow developers to
write:

```php
// this breaks the build tool — constant reference not resolvable from AST
new AppConfig(providers: [HttpConstants::HTTP_COMPONENT_PROVIDER])
```

Since the build tool cannot follow constant references without executing code, and since `::class` is already the
canonical, IDE-supported, autoloader-verified way to reference a class in PHP and Java, the component provider constants
class adds no value and introduces a failure mode. It is not part of the framework.

**Binding key constants files are unaffected** — those exist for Go/TypeScript/Python where no `::class` equivalent
exists, and for cross-language string identity. They are never used in provider lists.

### Build Tool Invocation

All Sindri implementations take the config file path directly — identical interface across all languages:

```bash
# PHP
sindri generate src/Config/AppConfig.php

# Java
sindri generate src/main/java/app/config/AppConfig.java

# Go
sindri generate app/config/app_config.go

# Python
sindri generate app/config/app_config.py

# TypeScript
sindri generate src/config/AppConfig.ts
```

The file path approach is consistent, requires no class name resolution logic, and works identically in all five Sindri
implementations.

The build tool reads provider list methods from AST without executing them. All provider list methods must return simple
list/array/map literals with no conditional logic — this is a hard contract enforced by the build tool.

---

## The Four Output Classes

The build tool always generates exactly four classes regardless of how many providers or routes the application has:

| Output class         | Contains                                                                            |
|----------------------|-------------------------------------------------------------------------------------|
| `AppContainerData`   | All bindings from every `ServiceProvider` across all components                     |
| `AppEventData`       | All listeners from every `ListenerProvider` across all components                   |
| `AppHttpRoutingData` | All HTTP routes from every `HttpRouteProvider` across all components, fully indexed |
| `AppCliRoutingData`  | All CLI routes from every `CliRouteProvider` across all components, fully indexed   |

---

## Build Tool Contract

Any method the build tool reads must return a single flat literal with no logic. Applies to all languages:

```
✅ Simple list/array/slice literal
✅ Simple map/dict/record literal
✅ Class references (::class / .class / ClassName / string constants)
✅ Object instantiations (new Route(...) / Route::get(...))
✅ Method references ([self::class, 'method'] / ClassName::method / p.Method)
✅ Constructor calls with literal arguments

❌ Conditional logic (if / switch / ternary)
❌ Variable references ($var / var / variable)
❌ Method calls other than constructors and static factories
❌ Loops (for / foreach / while)
❌ Variable accumulation (array_push / append / push)
```

Error on violation:

```
Error: UserServiceProvider::publishers() must return a simple map literal.
Conditional logic and dynamic computation are not supported in provider methods.
The build tool requires static analysis of this method's return value.
See: https://valkyrja.io/docs/providers#build-tool-compatibility
```

---

## Handler Method Pointer Convention

Handlers must be **method pointers** — references to static methods on the same class as the provider or controller that
defines the route, listener, or binding. They must not be inline closures or lambdas.

This is the single most important convention for keeping Sindri simple, import-safe, and conflict-free.

### Why Method Pointers

**Import safety** — Sindri reads exactly one file per provider or controller. All imports needed for handler bodies are
in that one file. No cross-file import aggregation, no conflict detection, no registry.

**Conflict elimination** — import conflicts only arise when merging imports across files. Method pointers keep handlers
self-contained in their source file, eliminating the root cause entirely.

**Consistency** — service providers already follow this pattern exactly. `publishers()` returns a map of service id →
method reference on the same class. Route and listener providers now follow the same pattern. The entire framework is
consistent end to end.

### The Pattern

Every handler is a static method on the same class that declares it. The route/listener/binding definition points to
that method by name. Sindri reads the method body from the same file — no cross-file resolution needed.

**Service providers** — already correct:

```php
// publishers() points to methods on the same class
public static function publishers(): array
{
    return [
        RouterContract::class => [self::class, 'publishRouter'],
    ];
}

public static function publishRouter(ContainerContract $c): void
{
    $c->setSingleton(RouterContract::class, new Router($c->make(DispatcherContract::class)));
}
```

**Route providers** — same pattern:

```php
// getRoutes() points to handler methods on the same class
public static function getRoutes(): array
{
    return [
        HttpRoute::get('/users/{id}', [self::class, 'showUser']),
        HttpRoute::post('/users',     [self::class, 'createUser']),
    ];
}

public static function showUser(ContainerContract $c, array $args): ResponseContract
{
    return $c->getSingleton(UserController::class)->show($args['id']);
}

public static function createUser(ContainerContract $c, array $args): ResponseContract
{
    return $c->getSingleton(UserController::class)->create($args);
}
```

**Listener providers** — same pattern:

```php
// getListeners() points to handler methods on the same class
public static function getListeners(): array
{
    return [
        Listener::on(UserCreatedEvent::class, [self::class, 'onUserCreated']),
    ];
}

public static function onUserCreated(ContainerContract $c, array $args): mixed
{
    return $c->getSingleton(UserCreatedListener::class)->handle($args['user_id']);
}
```

### Annotated Controllers — PHP, Java, Python

For annotated controllers, `#[Handler]` / `@Handler` / `@handler` lives on the **implementation method** and carries a *
*callable reference** pointing to the static handler method. The handler may live on the same controller class, the
route provider, or any other class — Sindri follows the callable reference to wherever the handler lives.

```
Annotations live on:    the implementation method (show, store, index etc.)
#[Handler] points to:   a callable (ClassName, methodName) — any class, anywhere
Forge reads:            the handler method body from whichever file the callable resolves to
```

**PHP — handler on same controller:**

```php
class UserController
{
    // Annotations on the implementation method
    #[Route('GET', '/users/{id}')]
    #[Parameter('id', pattern: '[0-9]+')]
    #[Handler([self::class, 'showHandler'])]  // callable — class + method name
    public function show(string $id): ResponseContract
    {
        // actual implementation — not read by Sindri
    }

    // Forge resolves [self::class, 'showHandler'] → this file → reads this method
    public static function showHandler(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(self::class)->show($args['id']);
    }
}
```

**PHP — handler on route provider:**

```php
class UserController
{
    #[Route('GET', '/users/{id}')]
    #[Parameter('id', pattern: '[0-9]+')]
    #[Handler([UserHttpRouteProvider::class, 'showUser'])]  // points elsewhere
    public function show(string $id): ResponseContract { /* ... */ }
}

class UserHttpRouteProvider implements HttpRouteProviderContract
{
    // Forge resolves callable → this file → reads this method + this file's imports
    public static function showUser(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(UserController::class)->show($args['id']);
    }
}
```

**Java — handler on same controller:**

```java
public class UserController {
    @Route(method = "GET", path = "/users/{id}")
    @Parameter(name = "id", pattern = "[0-9]+")
    @Handler(clazz = UserController.class, method = "showHandler")
    public ResponseContract show(String id) { /* actual implementation */ }

    // Forge resolves clazz + method → this file → reads this method
    public static ResponseContract showHandler(ContainerContract c, Map<String, Object> args) {
        return c.getSingleton(UserController.class).show((String) args.get("id"));
    }
}
```

**Python — handler on same controller:**

```python
class UserController:
    @route('GET', '/users/{id}')
    @parameter('id', pattern='[0-9]+')
    @handler((UserController, 'show_handler'))  # callable tuple
    def show(self, id: str) -> ResponseContract:
        pass  # actual implementation — not read by Sindri

    # Forge resolves callable → this file → reads this method
    @staticmethod
    def show_handler(c: ContainerContract, args: dict) -> ResponseContract:
        return c.get_singleton(UserController).show(args['id'])
```

### Why Not Inline Closures

Inline closures in route definitions would require Sindri to parse and rewrite type references from potentially any file
in the application:

```php
// ❌ inline closure — Sindri must resolve ContainerContract, UserController
//    from the imports of this specific file AND know if they conflict
//    with imports from other route providers
HttpRoute::get('/users/{id}', static fn($c, $args) =>
    $c->getSingleton(UserController::class)->show($args['id'])
)

// ✅ method pointer — Sindri reads the method body from this same file
//    all imports are right here, no cross-file resolution needed
HttpRoute::get('/users/{id}', [self::class, 'showUser'])
```

### The Remaining Import Conflict Case

Even with method pointers, the generated output file aggregates handler bodies from multiple provider files. If two
providers import `UserController` from different namespaces, Sindri cannot silently resolve this:

```
Error: Import conflict in generated AppHttpRoutingData.
  App\Http\Controllers\UserController  (from UserHttpRouteProvider)
  App\Api\Controllers\UserController   (from ApiHttpRouteProvider)
Both resolve to the short name 'UserController'. Use the FQN directly in your handler
or rename one of the classes to eliminate the conflict.
```

The developer fixes it in their source — Sindri does not silently rename. Genuine class name collisions across providers
are structural problems in the application that the developer must resolve.

### Build Tool Contract — Handler Methods

**For route providers and listener providers** (explicit `getRoutes()` / `getListeners()`):

```
✅ [self::class, 'methodName']           — PHP method pointer on same class
✅ ClassName::methodName                  — Java static reference
✅ self.method_name                       — Python method reference on same class
✅ p.MethodName                           — Go method reference on same struct
✅ this.methodName                        — TypeScript method reference on same class

✅ Handler method must be static
✅ Handler method must be on the same class as the provider
✅ All type refs in handler body must be imported in the provider file

❌ Inline closures or lambdas in route/listener definitions
```

**For annotated controllers and listeners** (PHP, Java, Python only):

```
✅ #[Handler([ClassName::class, 'methodName'])]   — PHP callable on any class
✅ @Handler(clazz = ClassName.class, method = "m") — Java callable on any class
✅ @handler((ClassName, 'method_name'))            — Python callable on any class

✅ Annotation lives on the implementation method (the instance method)
✅ Handler method must be static — anywhere in the codebase
✅ All type refs in handler body must be imported in the handler's own file

❌ Annotation on the static handler method itself
❌ Handler method that is not static
```

---

## The Build Tool Bootstrap Problem — And Why It Doesn't Matter

The build tool is a Valkyrja application, which means it bootstraps via the provider tree like any other Valkyrja
application. Since it ships without its own cache files, it pays the full bootstrap cost on every run. In practice this
is irrelevant — the build tool runs once at deploy time in a CI pipeline, not per-request. A bootstrap that takes a
second or two is acceptable.

For environments where this matters:

```bash
# pass 1 — build the build tool without cache (slow first run)
sindri build-self --output build-tool-bootstrap

# pass 2 — use the bootstrap to generate the build tool's own cache
./build-tool-bootstrap generate --self

# pass 3 — rebuild the build tool with its own cache (fast)
sindri build-self --with-cache --output sindri-optimized
```

This is the same two-pass compile process the build tool applies to any compiled language application. The build tool
eating its own dog food is a validation that the framework's cache-optional architecture is self-consistent.

---

## Getting an AST From a File Path

The build tool takes the application config file path as its CLI argument — consistent across all languages. No class
name resolution, no guesswork about directory structure, no extra logic. The developer passes the path directly.

```bash
# all languages — same pattern
sindri generate src/config/AppConfig.php
sindri generate src/main/java/app/config/AppConfig.java
sindri generate app/config/app_config.go
sindri generate app/config/app_config.py
sindri generate src/config/AppConfig.ts
```

PHP is the only language with native class-to-file resolution (`ReflectionClass::getFileName()`), but since all other
languages require a file path anyway, Sindri PHP takes a file path too — keeping the CLI interface identical across all
language implementations.

---

### PHP

```php
#!/usr/bin/env php
<?php

require_once 'vendor/autoload.php';

use PhpParser\ParserFactory;
use PhpParser\NodeTraverser;
use PhpParser\PrettyPrinter;
use PhpParser\NodeVisitorAbstract;
use PhpParser\Node;

// entry point — file path passed as CLI argument
// e.g. sindri generate src/config/AppConfig.php
$filepath = $argv[1] ?? throw new InvalidArgumentException('File path required');

if (!file_exists($filepath)) {
    throw new InvalidArgumentException("File not found: {$filepath}");
}

// step 1: parse the file into an AST
$parser = (new ParserFactory())->createForNewestSupportedVersion();
$ast    = $parser->parse(file_get_contents($filepath));

// $ast is an array of PhpParser\Node objects — ready for traversal

// step 2: collect use statements for FQN resolution
$useStatements = [];
$traverser = new NodeTraverser();
$traverser->addVisitor(new class($useStatements) extends NodeVisitorAbstract {
    public function __construct(private array &$map) {}
    public function enterNode(Node $node): void {
        if (!$node instanceof Node\Stmt\Use_) return;
        foreach ($node->uses as $use) {
            $alias       = $use->alias?->name ?? $use->name->getLast();
            $this->map[$alias] = $use->name->toString();
        }
    }
});
$traverser->traverse($ast);

// step 3: walk the AST with a visitor
$traverser = new NodeTraverser();
$traverser->addVisitor(new class extends NodeVisitorAbstract {
    public function enterNode(Node $node): void {
        if ($node instanceof Node\Stmt\Class_) {
            echo "Class: " . $node->name . PHP_EOL;
        }
        if ($node instanceof Node\Stmt\ClassMethod) {
            echo "  Method: " . $node->name . PHP_EOL;
        }
    }
});
$traverser->traverse($ast);

// step 4: print any AST node back to PHP source text
$printer    = new PrettyPrinter\Standard();
$sourceText = $printer->prettyPrint([$ast[0]]);
```

**Key classes:**

- `ParserFactory` — creates the parser targeting the installed PHP version
- `NodeTraverser` + `NodeVisitorAbstract` — visitor pattern for walking the AST
- `PrettyPrinter\Standard` — prints any AST node back to PHP source text

---

### Java

```java
import com.sun.source.tree.*;
import com.sun.source.util.*;

import javax.tools.*;
import java.util.List;

public class ForgeParser {

    /**
     * Parse a Java source file into an AST (CompilationUnitTree).
     * Takes a file path directly — no class name resolution needed.
     *
     * e.g. sindri generate src/main/java/app/config/AppConfig.java
     */
    public static CompilationUnitTree parseFile(String filePath) throws Exception {
        // step 1: get the system Java compiler
        JavaCompiler compiler = ToolProvider.getSystemJavaCompiler();

        // step 2: create a file manager
        StandardJavaFileManager fileManager =
                compiler.getStandardFileManager(null, null, null);

        // step 3: wrap the source file
        Iterable<? extends JavaFileObject> compilationUnits =
                fileManager.getJavaFileObjects(filePath);

        // step 4: create a compilation task — parse only, no output
        JavaCompiler.CompilationTask task = compiler.getTask(
                null,                   // writer (null = stderr)
                fileManager,
                null,                   // diagnostic listener
                List.of("-proc:none"),  // disable annotation processing
                null,
                compilationUnits
        );

        // step 5: parse — no compilation, no class files written
        JavacTask javacTask = (JavacTask) task;
        Iterable<? extends CompilationUnitTree> units = javacTask.parse();

        return units.iterator().next();
        // unit.getTypeDecls() — class/interface declarations
        // unit.getImports()   — import statements for FQN resolution
    }

    /**
     * Walk the AST using TreeScanner.
     */
    public static void walkAST(CompilationUnitTree unit) {
        unit.accept(new TreeScanner<Void, Void>() {
            @Override
            public Void visitClass(ClassTree node, Void p) {
                System.out.println("Class: " + node.getSimpleName());
                return super.visitClass(node, p); // recurse into children
            }

            @Override
            public Void visitMethod(MethodTree node, Void p) {
                System.out.println("  Method: " + node.getName());
                return super.visitMethod(node, p);
            }

            @Override
            public Void visitReturn(ReturnTree node, Void p) {
                System.out.println("    Return: " + node.getExpression());
                return super.visitReturn(node, p);
            }
        }, null);
    }

    public static void main(String[] args) throws Exception {
        // file path passed directly as CLI argument
        String filePath = args[0]; // e.g. "src/main/java/app/config/AppConfig.java"

        CompilationUnitTree unit = parseFile(filePath);
        walkAST(unit);
    }
}
```

**Key classes:**

- `ToolProvider.getSystemJavaCompiler()` — gets the javac compiler instance at runtime
- `JavacTask.parse()` — parses source to AST without compiling or writing class files
- `CompilationUnitTree` — top-level AST node containing imports and class declarations
- `TreeScanner<R, P>` — generic visitor, `super.visitX()` recurses into children

---

### Go

```go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	// file path passed directly as CLI argument
	// e.g. sindri generate app/config/app_config.go
	filepath := os.Args[1]

	// step 1: create a FileSet to track position information
	fset := token.NewFileSet()

	// step 2: parse the file into an AST
	file, err := parser.ParseFile(
		fset,
		filepath,
		nil,                  // src — nil reads from disk
		parser.ParseComments, // include comments in AST
	)
	if err != nil {
		panic(fmt.Sprintf("parse error: %v", err))
	}

	// step 3: walk the AST with ast.Inspect
	// returns true to continue into children, false to stop
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch node := n.(type) {
		case *ast.FuncDecl:
			fmt.Printf("Function: %s at %s\n",
				node.Name.Name,
				fset.Position(node.Pos()),
			)
		case *ast.ReturnStmt:
			for _, result := range node.Results {
				fmt.Printf("  Return: %T\n", result)
			}
		case *ast.CompositeLit:
			// slice [] or map {} literal — provider lists live here
			fmt.Printf("  Composite literal with %d elements\n", len(node.Elts))
			for _, elt := range node.Elts {
				fmt.Printf("    Element: %T = %v\n", elt, elt)
			}
		}

		return true // continue walking into children
	})

	// step 4: print any AST node back to source text
	// go/printer writes to any io.Writer
	// printer.Fprint(os.Stdout, fset, file)
}
```

**Key packages:**

- `go/parser` — `parser.ParseFile()` parses a `.go` file into `*ast.File`
- `go/token` — `token.FileSet` tracks file/line/column positions
- `go/ast` — all AST node types and `ast.Inspect()` walker
- `go/printer` — `printer.Fprint(w, fset, node)` prints any node back to source text

> **Note on `go/packages`:** The simpler `go/parser.ParseFile()` is sufficient for reading provider list literals and
> handler function bodies since these are all simple literals. `golang.org/x/tools/go/packages` is available when full
> type resolution across packages is needed (e.g. resolving an identifier's FQN across module boundaries), but adds the
> requirement of a properly configured Go module environment.

---

### Python

```python
#!/usr/bin/env python3
"""
sindri — Python AST bootstrap.

Usage: sindri generate app/config/app_config.py
"""

import ast
import sys
from pathlib import Path


def parse_file(filepath: str) -> ast.Module:
    """
    Parse a Python source file into an AST.
    Takes a file path directly — no import or class resolution needed.
    """
    source = Path(filepath).read_text(encoding='utf-8')
    return ast.parse(source, filename=filepath)


def walk_ast(tree: ast.Module) -> None:
    """Walk all nodes and print class/function/return structure."""
    for node in ast.walk(tree):
        if isinstance(node, ast.ClassDef):
            print(f"Class: {node.name}")

        elif isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            print(f"  Method: {node.name}")

        elif isinstance(node, ast.Return) and node.value:
            # ast.unparse converts any node back to Python source text
            print(f"    Return: {ast.unparse(node.value)}")

            # if it's a list literal — provider lists and route lists live here
            if isinstance(node.value, ast.List):
                for elt in node.value.elts:
                    print(f"      Element: {ast.unparse(elt)}")


def collect_imports(tree: ast.Module) -> dict[str, str]:
    """
    Build a map of simple name -> fully qualified module path.
    Used for FQN resolution when reading class references.

    e.g. 'from app.http.provider import HttpComponentProvider'
      -> {'HttpComponentProvider': 'app.http.provider.HttpComponentProvider'}
    """
    import_map: dict[str, str] = {}

    for node in ast.walk(tree):
        if isinstance(node, ast.ImportFrom):
            module = node.module or ''
            for alias in node.names:
                name = alias.asname or alias.name
                import_map[name] = f"{module}.{alias.name}"
        elif isinstance(node, ast.Import):
            for alias in node.names:
                name = alias.asname or alias.name
                import_map[name] = alias.name

    return import_map


if __name__ == '__main__':
    # file path passed directly as CLI argument
    filepath = sys.argv[1]  # e.g. 'app/config/app_config.py'

    tree = parse_file(filepath)
    print(f"Parsed: {filepath}")

    walk_ast(tree)

    imports = collect_imports(tree)
    print(f"Imports: {imports}")

    # dump the full AST structure — useful for debugging
    # print(ast.dump(tree, indent=2))
```

**Key modules:**

- `ast.parse(source, filename)` — parses Python source string into `ast.Module`
- `ast.walk(tree)` — generator yielding every node in the tree recursively
- `ast.unparse(node)` — converts any AST node back to Python source text
- `ast.dump(tree, indent=2)` — pretty-prints the full AST structure for debugging

---

### TypeScript

```typescript
import ts from 'typescript'
import * as path from 'path'

/**
 * sindri — TypeScript AST bootstrap.
 *
 * Usage: sindri generate src/config/AppConfig.ts
 */

interface ParsedFile {
    sourceFile: ts.SourceFile
    program: ts.Program
    checker: ts.TypeChecker
}

/**
 * Parse a TypeScript source file into an AST.
 * Takes a file path directly — loaded via tsconfig for full type info.
 */
function parseFile(
    filePath: string,
    tsconfigPath: string = 'tsconfig.json'
): ParsedFile {
    // step 1: read and parse tsconfig.json
    const configFile = ts.readConfigFile(tsconfigPath, ts.sys.readFile)
    if (configFile.error) {
        throw new Error(ts.flattenDiagnosticMessageText(
            configFile.error.messageText, '\n'
        ))
    }

    // step 2: resolve tsconfig options and file list
    const config = ts.parseJsonConfigFileContent(
        configFile.config,
        ts.sys,
        path.dirname(path.resolve(tsconfigPath))
    )

    // step 3: create the compiler program — parses all project files
    const program = ts.createProgram({
        rootNames: config.fileNames,
        options: config.options,
    })

    // step 4: get the type checker
    const checker = program.getTypeChecker()

    // step 5: get the SourceFile (AST) for the given path
    const absolutePath = path.resolve(filePath)
    const sourceFile = program.getSourceFile(absolutePath)

    if (!sourceFile) {
        throw new Error(
            `File not found in program: ${absolutePath}\n` +
            `Ensure the file is included in tsconfig.json`
        )
    }

    return {sourceFile, program, checker}
}

/**
 * Walk all nodes in a source file recursively.
 */
function walkAST(
    node: ts.Node,
    visitor: (node: ts.Node) => void
): void {
    visitor(node)
    ts.forEachChild(node, child => walkAST(child, visitor))
}

/**
 * Get the original source text of any AST node.
 */
function getNodeText(node: ts.Node, sourceFile: ts.SourceFile): string {
    return node.getText(sourceFile)
}

/**
 * Resolve an identifier to its fully qualified name.
 * e.g. 'HttpComponentProvider' -> '"@valkyrja/http/provider".HttpComponentProvider'
 */
function getFQN(node: ts.Node, checker: ts.TypeChecker): string | undefined {
    const symbol = checker.getSymbolAtLocation(node)
    return symbol ? checker.getFullyQualifiedName(symbol) : undefined
}

// entry point — file path passed directly as CLI argument
const filePath = process.argv[2] // e.g. 'src/config/AppConfig.ts'

const {sourceFile, checker} = parseFile(filePath)
console.log(`Parsed: ${sourceFile.fileName}`)

walkAST(sourceFile, (node) => {
    if (ts.isClassDeclaration(node) && node.name) {
        console.log(`Class: ${node.name.text}`)
    }

    if (ts.isMethodDeclaration(node)) {
        const name = (node.name as ts.Identifier).text
        console.log(`  Method: ${name}`)
    }

    if (ts.isReturnStatement(node) && node.expression) {
        console.log(`    Return: ${getNodeText(node.expression, sourceFile)}`)

        // array literal — provider lists and route lists live here
        if (ts.isArrayLiteralExpression(node.expression)) {
            node.expression.elements.forEach(elt => {
                const fqn = getFQN(elt, checker)
                console.log(`      Element FQN: ${fqn}`)
            })
        }
    }
})
```

**Key APIs:**

- `ts.readConfigFile` + `ts.parseJsonConfigFileContent` — reads and resolves `tsconfig.json`
- `ts.createProgram` — creates the compiler with all project files parsed
- `program.getSourceFile(path)` — gets the `ts.SourceFile` (AST) for an absolute path
- `program.getTypeChecker()` — type checker for FQN resolution
- `ts.forEachChild(node, cb)` — iterates immediate children of any node
- `node.getText(sourceFile)` — original source text of any AST node
- `checker.getSymbolAtLocation(node)` + `checker.getFullyQualifiedName(symbol)` — FQN resolution
- Type guards: `ts.isClassDeclaration`, `ts.isMethodDeclaration`, `ts.isReturnStatement`, `ts.isArrayLiteralExpression`,
  etc.

## Constants Files

Every Valkyrja component ships a constants file for all five languages. These are **part of the framework source** — not
generated by Sindri. They define the binding key for every contract and class in that component. Developers always
reference the constant, never the raw string.

The value type differs per language but the abstraction is consistent — callers never write the raw key, they always use
the constant:

**PHP** — constants hold `::class` strings (compile-time FQN strings):

```php
// Valkyrja/Container/ContainerConstants.php — shipped with framework
class ContainerConstants {
    public const CONTAINER  = ContainerContract::class;
    public const DISPATCHER = DispatcherContract::class;
    public const ROUTER     = RouterContract::class;
}
```

**Java** — constants hold `.class` objects (compile-time type tokens):

```java
// io/valkyrja/container/ContainerConstants.java — shipped with framework
public final class ContainerConstants {
    public static final Class<ContainerContract> CONTAINER = ContainerContract.class;
    public static final Class<DispatcherContract> DISPATCHER = DispatcherContract.class;
    public static final Class<RouterContract> ROUTER = RouterContract.class;
}
```

**Go** — constants hold FQN string literals:

```go
// io/valkyrja/container/container_constants.go — shipped with framework
const (
ContainerClass = "io.valkyrja.container.ContainerContract"
DispatcherClass = "io.valkyrja.container.DispatcherContract"
RouterClass = "io.valkyrja.http.routing.RouterContract"
)
```

**Python** — constants hold FQN string literals:

```python
# valkyrja/container/container_constants.py — shipped with framework
class ContainerConstants:
    CONTAINER = "io.valkyrja.container.ContainerContract"
    DISPATCHER = "io.valkyrja.container.DispatcherContract"
    ROUTER = "io.valkyrja.http.routing.RouterContract"
```

**TypeScript** — constants hold FQN string literals:

```typescript
// valkyrja/container/container-constants.ts — shipped with framework
export const ContainerConstants = {
    CONTAINER: 'io.valkyrja.container.ContainerContract',
    DISPATCHER: 'io.valkyrja.container.DispatcherContract',
    ROUTER: 'io.valkyrja.http.routing.RouterContract',
} as const
```

---

### Application Constants

Application constants follow the same pattern as framework constants — one constants file per component, same structure,
same naming convention. Developers write them by hand initially, following the framework pattern.

```
Framework constants  → shipped with framework source, maintained by Valkyrja
Application constants → written by the developer, following the same pattern
```

**Future enhancement:** Forge generating application constants automatically from the provider tree is a natural
extension — walk the application source, discover all contracts and service classes, derive their FQN strings, write the
constants files. This is a quality-of-life feature to implement after the core forge functionality is stable.

---

### How the FQN String Is Derived

For Go, Python, and TypeScript, Sindri derives the FQN string for each class from the language-native module/package
path:

| Language   | FQN format                                | Example                                 |
|------------|-------------------------------------------|-----------------------------------------|
| Go         | `io/valkyrja/container.ContainerContract` | dot-separated package path + type name  |
| Python     | `io.valkyrja.container.ContainerContract` | dot-separated module path + class name  |
| TypeScript | `io.valkyrja.container.ContainerContract` | dot-separated package path + class name |

The constant name is derived from the class name using the language-idiomatic convention — `SCREAMING_SNAKE_CASE` for
Python/TypeScript, `PascalCaseClass` suffix for Go.

---

### The Competitive Advantage

No existing Python framework achieves lazy provider loading at the container level. FastAPI uses function objects as
dependency identifiers with per-request resolution — no caching, no container. Django has no DI container. Third-party
containers like `lagom` use class objects as keys, forcing imports. Valkyrja's combination of string constant keys,
per-component constants files, lambda-wrapped bindings, and Python 3.14 lazy imports is a novel approach. The constants
files are the only visible cost to the developer — framework constants ship with the framework, and application
constants are generated by Sindri.

---

## AST Walk Strategy

Once the config file is parsed, Sindri follows a structured walk strategy. Each step produces data that feeds into the
next, ultimately assembling the four output data classes.

**Imports are collected from every file parsed.** This is the safe, correct default — when Sindri extracts a handler
method body and writes it into generated output, any type reference in that body must be written as a FQN. Without the
full import map Sindri cannot resolve short names to their FQN.

For provider list methods (config providers list, component provider sub-lists, route/listener class lists) Sindri only
needs the class identifier to locate the next file — import collection for those steps adds no cost and ensures no edge
case is missed.

The handler method pointer convention (see "Handler Method Pointer Convention" above) ensures that all imports needed
for any handler body are always present in the same file as the provider or controller that defines the route or
listener. No cross-file import aggregation is needed — each file is self-contained.

---

### Step 1: Walk the Config File

**Goal:** Extract the initial component providers list.

**What to collect:**

```
imports         → map of simple name → FQN (collected as standard practice)
providers list  → list of class references from the providers property/method
                  e.g. [HttpComponentProvider, CliComponentProvider, AppComponentProvider]
```

**Pattern:**

```
1. Parse config file into AST
2. Collect all import statements → import map
3. Find the providers list (array/slice/list literal on the config class)
4. Extract each class reference as an identifier string
5. For each identifier → resolve to file path → Step 1a (dependency resolution)
```

The providers list must be a simple literal — no variables, no method calls, no conditional logic. If it isn't, Sindri
emits an error.

---

### Step 1a: Resolve Component Dependencies

**Goal:** Expand the initial providers list by following each component's `getComponentProviders()` return value,
recursively, until the full transitive dependency set is known. Deduplicate throughout so no component is walked twice.

A component returns the other components it depends on — e.g. `HttpComponentProvider` returns
`[ContainerComponentProvider, EventComponentProvider]`. Those dependencies must be registered before the component that
depends on them. Sindri resolves the full ordered list before beginning the provider walk.

**Algorithm:**

```
resolved = []         // ordered list — final registration order
seen     = {}         // set of already-visited class identifiers

function resolve(identifier):
    if identifier in seen:
        return        // already resolved — skip, deduplication
    seen.add(identifier)

    file = resolve_to_file_path(identifier)
    ast  = parse(file)

    // read getComponentProviders() first — dependencies before self
    deps = extract_list(ast, 'getComponentProviders')
    for dep in deps:
        resolve(dep)  // recurse — depth-first, dependencies load before dependents

    resolved.append(identifier)  // append self after all dependencies

// seed with config providers list
for identifier in config_providers_list:
    resolve(identifier)

// resolved is now the full ordered, deduplicated component list
```

**Example:**

```
Config lists:     [HttpComponentProvider, CliComponentProvider, AppComponentProvider]

HttpComponentProvider.getComponentProviders()
  → [ContainerComponentProvider, EventComponentProvider]

CliComponentProvider.getComponentProviders()
  → [ContainerComponentProvider]          // already seen — skipped

AppComponentProvider.getComponentProviders()
  → [HttpComponentProvider]               // already seen — skipped

Resolved order:
  ContainerComponentProvider              // dependency of Http and Cli
  EventComponentProvider                  // dependency of Http
  HttpComponentProvider
  CliComponentProvider
  AppComponentProvider
```

Circular dependencies emit an error:

```
Error: Circular component dependency detected.
  HttpComponentProvider → AppComponentProvider → HttpComponentProvider
```

---

### Step 2: Walk Each ComponentProvider File

**Goal:** Categorize sub-providers by type so they can be handed to the correct walker.
Uses the fully resolved and ordered list from Step 1a — no further dependency resolution needed here.

**What to collect:**

```
imports             → map of simple name → FQN (collected as standard practice)
container_providers → list of class identifiers (ServiceProvider classes)
event_providers     → list of class identifiers (ListenerProvider classes)
http_providers      → list of class identifiers (HttpRouteProvider classes)
cli_providers       → list of class identifiers (CliRouteProvider classes)
```

**Pattern:**

```
1. Parse ComponentProvider file into AST
2. Collect all import statements → import map
3. Find each of the four sub-provider list methods by name
   (getContainerProviders, getEventProviders, getHttpProviders, getCliProviders)
4. Extract return value — must be a simple list literal
5. Extract each class reference as an identifier
6. For each identifier → resolve to file path → dispatch to Step 3
```

> **Important:** Sindri reads only the return value literal of each method. Any imperative code inside the method body —
> container bindings registered directly, conditional logic, variable assignments — is invisible to Sindri and will not be
> included in the generated cache. This is by design: Sindri is a static AST reader, not a runtime executor. Bindings not
> discoverable via the provider tree are cache-incompatible.

---

### Step 3a: Walk ServiceProvider Files (Container Bindings)

**Goal:** Extract all container bindings — the mapping of binding key to publisher function body.

**What to collect:**

```
imports         → map of simple name → FQN

publishers      → map of binding key → publisher method body source text
                  key format is language-specific:
                    PHP/Java      → ::class / .class  (class name resolved to FQN via imports)
                    Python        → class name        (resolved to FQN via imports)
                    Go/TypeScript → string constant   (string literal value — used as-is)
```

**Pattern:**

```
1. Parse ServiceProvider file into AST
2. Collect all import statements → import map
3. Find the publishers() method
4. Extract the map literal from the return statement
5. For each entry:
   a. Extract the key → resolve to canonical FQN string (language-specific, see table below)
   b. Extract the value → method reference (e.g. self::publishRouter)
   c. Resolve method reference → find that method in the same file
   d. Extract the callable value as a literal → written as-is into generated output
6. Output: map of { canonical key string → callable literal }
```

**Language-specific key resolution:**

| Language   | Key in source                        | Resolution                                           |
|------------|--------------------------------------|------------------------------------------------------|
| PHP        | `RouterContract::class`              | Resolve `RouterContract` via import map → FQN string |
| Java       | `RouterContract.class`               | Resolve `RouterContract` via import map → FQN string |
| Python     | `RouterContract` (class name)        | Resolve via import map → FQN string                  |
| Go         | `"io.valkyrja.http.RouterContract"`  | String literal — use as-is                           |
| TypeScript | `RouterClass` (string constant name) | Resolve constant value via import map → string       |

---

### Step 3b: Walk ListenerProvider Files (Event Listeners)

**Goal:** Extract explicit listeners and the list of annotated classes to scan.

**Imports are needed for explicit listeners** — handler closures and event type references in listener data objects will
be written into generated output and must be fully qualified.

**Imports are not needed for the annotated class list** — those class identifiers are used only to locate files to scan
in Step 4b.

**What to collect:**

```
imports             → map of simple name → FQN
                      needed to rewrite handler closures and event type references

explicit_listeners  → list of listener data objects from getListeners()
                      each listener carries:
                        event type  (class reference → FQN via imports)
                        priority    (integer literal, if present)
                        handler     (closure source text, type refs rewritten to FQN)

annotated_classes   → list of class identifiers from getListenerClasses()
                      (PHP, Java, Python only — Go and TypeScript omit this method)
                      no imports needed — identifiers only, used to locate files
```

**Pattern:**

```
1. Parse ListenerProvider file into AST
2. Collect imports → FQN map  ← needed for explicit_listeners
3. Find getListeners() method:
   a. Extract return list literal
   b. For each listener constructor call:
      - Extract event type class reference → resolve to FQN via imports
      - Extract handler closure source text → rewrite type refs to FQN
      - Extract priority if present
4. Find getListenerClasses() method (PHP/Java/Python only):
   a. Extract return list literal
   b. Extract each class reference as an identifier (no FQN resolution needed)
   c. For each identifier → resolve to file path → Step 4b
5. Output: { explicit_listeners: [...], annotated_class_files: [...] }
```

---

### Step 3c: Walk HttpRouteProvider Files (HTTP Routes)

**Goal:** Extract explicit routes and the list of annotated controller classes to scan.

**What to collect:**

```
imports             → map of simple name → FQN

explicit_routes     → list of route data objects from getRoutes()
                      each route carries:
                        method      (GET, POST, etc. — from factory method name)
                        path        ('/users/{id}' — string literal)
                        parameters  (list of Parameter objects — name + pattern string literals)
                        handler     (method pointer → [self::class, 'methodName'])
                        handler body (static method on same class → extract and rewrite)
                        middleware  (if present)
                        name        (if present)

annotated_classes   → list of class identifiers from getControllerClasses()
                      (PHP, Java, Python only — Go and TypeScript omit this method)
```

**Pattern:**

```
1. Parse HttpRouteProvider file into AST
2. Collect all import statements → import map
3. Find getRoutes() method:
   a. Extract return list literal
   b. For each route constructor call (e.g. HttpRoute::get(...)):
      - Extract HTTP method from the factory method name (get/post/put/delete/patch)
      - Extract path string literal
      - Extract handler callable → [self::class, 'methodName'] — written as-is into output
      - Extract Parameter constructor calls if present:
          name    → string literal argument
          pattern → string literal argument (default '[^/]+' if absent)
      - Extract middleware, name, and other metadata if present
4. Find getControllerClasses() method (PHP/Java/Python only):
   a. Extract return list literal
   b. Extract each class reference as an identifier
   c. For each identifier → resolve to file path → Step 4a
5. Output: { explicit_routes: [...], annotated_class_files: [...] }
```

---

### Step 3d: Walk CliRouteProvider Files (CLI Routes)

Identical to Step 3c with:

- `getRoutes()` → CLI route objects (command path, description, handler)
- `getControllerClasses()` → CLI controller/command classes to scan for annotations
- No HTTP method field — CLI routes use command path only
- Handler returns `OutputContract` instead of `ResponseContract`

---

### Step 4a: Scan Controller/Command Class Files for Route Annotations

**Goal:** Extract routes from annotated/decorated methods. PHP, Java, Python only.

**Input:** file path of a controller class (resolved in Step 3c or 3d).

Sindri reads annotation literals and constructs route data objects — **no method body extraction, no import resolution
of the callable**. The callable from `#[Handler]` is written directly into the generated output as-is, exactly like a
callable in an explicit `getRoutes()` route or a service binding:

```
// service binding — callable literal written as-is
SomeServiceId::class => [SomeProvider::class, 'publishMethod']

// explicit route — callable literal written as-is
new Route('/path', 'name', [SomeClass::class, 'theHandlerMethod'])

// annotated route — same, callable literal written as-is
#[Handler([SomeClass::class, 'theHandlerMethod'])]
```

**What to collect per annotated implementation method:**

```
imports     → map of simple name → FQN
              needed only to resolve the callable class name to FQN for the output literal

per method:
  callable    → (ClassName, methodName) from #[Handler] — written as-is into output
  parameters  → list from #[Parameter] / @Parameter annotations
                each: name (string literal), pattern (string literal)
  path        → from route annotation — string literal
  method      → HTTP method — from route annotation
  middleware  → from middleware annotation if present
  name        → from name annotation if present
```

**Pattern:**

```
1. Parse controller file into AST
2. Collect all import statements → import map
3. Walk all class methods:
   For each implementation method:
     a. Check for @Handler / #[Handler] annotation:
        - Extract callable literal: (ClassName, methodName)
        - Resolve ClassName via imports → FQN (so output contains FQN, not short name)
     b. Check for @Parameter / #[Parameter] annotations (may be multiple):
        - Extract name string literal
        - Extract pattern string literal (default '[^/]+' if absent)
     c. Check for route annotation (HTTP method + path):
        - Extract HTTP method
        - Extract path string literal
     d. If handler found → construct route data from literals → add to output list
4. Output: list of route data objects — same shape as explicit routes from Step 3c
           callable written as FQN literal into generated cache data class
```

---

### Step 4b: Scan Listener Class Files for Listener Annotations

**Goal:** Extract listeners from annotated/decorated methods. PHP, Java, Python only.

`#[Handler]` / `@Handler` / `@handler` lives on the **implementation method** and carries a **callable reference** —
same pattern as annotated controllers. The handler may live on the listener class itself, the listener provider, or any
other class.

**Pattern:**

```
1. Parse listener class file into AST
2. Collect all import statements → import map (for resolving callable class names)
3. Walk all class methods:
   For each method:
     a. Check for @Handler / #[Handler] / @handler annotation:
        - Extract callable: (ClassName, methodName)
        - Resolve ClassName via listener file's imports → FQN
     b. Check for @ListensTo / #[ListensTo] annotation:
        - Extract event type class reference → resolve to FQN via imports
        - Extract priority if present
     c. If handler found:
        - Resolve callable FQN → file path
        - Parse handler file → collect its imports
        - Find methodName static method in handler file
        - Extract method body → rewrite type refs using handler file's imports
        - Add listener data to output list
4. Output: list of listener data objects, same shape as explicit_listeners from Step 3b
```

---

### Step 5: Run ProcessorContract for Routes

**Goal:** Compile route regexes from path patterns and Parameter objects.

For every route collected in Steps 3c, 3d, 4a:

```
1. Construct a plain ValkyrjaRoute from extracted data:
   - path, method, parameters (name + pattern pairs)
2. Pass to ProcessorContract::route() — same processor used at runtime
3. Read back the compiled regex string
4. Store alongside route data for Step 6
```

Sindri has `ProcessorContract` available — it is compiled from the framework source. Regex compilation happens once at
build time and is stored pre-compiled in the cache data class.

---

### Step 6: Assemble and Generate the Four Data Classes

**Goal:** Aggregate everything and write the four output source files.

```
AppContainerData:
  For each ServiceProvider (Step 3a):
    Merge publishers map into global bindings map

AppEventData:
  For each ListenerProvider (Step 3b):
    Merge explicit_listeners
    Merge annotated listeners from Step 4b
  Index by event type

AppHttpRoutingData:
  For each HttpRouteProvider (Step 3c):
    Merge explicit_routes
    Merge annotated routes from Step 4a
  Run all routes through Step 5 (ProcessorContract → compiled regex)
  Build four indexes: routes, paths, dynamicPaths, regexes

AppCliRoutingData:
  Same as AppHttpRoutingData but for CLI routes
```

Sindri writes source files as text — it is a text generator, not a compiler. Application classes referenced in the
output are written by name only. No application classes need to be compiled in at generation time.

---

### Full Walk Summary

```
Config file
  └── imports + providers list
        ├── ComponentProvider files
        │     └── imports + four categorized sub-provider identifier lists
        │           ├── ServiceProvider files
        │           │     imports + publishers map + handler method bodies
        │           │     → Step 3a → bindings
        │           ├── ListenerProvider files
        │           │     imports + explicit listeners + handler method bodies
        │           │     → Step 3b
        │           │     annotated class identifiers → file paths
        │           │           └── Listener class files
        │           │                 imports + annotations
        │           │                 + handler method bodies (same file)
        │           │                 → Step 4b → listeners
        │           ├── HttpRouteProvider files
        │           │     imports + explicit routes + handler method bodies
        │           │     → Step 3c
        │           │     annotated class identifiers → file paths
        │           │           └── Controller class files
        │           │                 imports + annotations
        │           │                 + handler method bodies (same file)
        │           │                 → Step 4a → routes
        │           └── CliRouteProvider files
        │                 imports + explicit routes + handler method bodies
        │                 → Step 3d
        │                 annotated class identifiers → file paths
        │                       └── Command class files
        │                             imports + annotations
        │                             + handler method bodies (same file)
        │                             → Step 4a → routes
        │
        ├── Step 5: ProcessorContract → compile regexes for all routes
        └── Step 6: generate AppContainerData, AppEventData,
                    AppHttpRoutingData, AppCliRoutingData
```

## Language-Specific AST Implementation

### PHP

**Dependencies:** `nikic/php-parser` (dev dependency), standard PHP autoloader

**File resolution:** `ReflectionClass::getFileName()` — resolves any autoloadable class to its source file path. No
environment assumptions beyond the class being autoloadable.

**Provider tree walk:**

```php
// build tool — walk provider tree via AST
$parser   = (new ParserFactory())->createForNewestSupportedVersion();
$printer  = new PrettyPrinter\Standard();

function walkProvider(string $className, Parser $parser): array
{
    // resolve class to file path
    $reflector = new ReflectionClass($className);
    $filepath  = $reflector->getFileName();

    // parse source file
    $ast = $parser->parse(file_get_contents($filepath));

    // collect use statements for FQN resolution
    $useStatements = collectUseStatements($ast);

    // find the provider method and extract its return list
    $visitor = new ProviderListVisitor($useStatements);
    (new NodeTraverser())->addVisitor($visitor)->traverse($ast);

    return $visitor->getProviderClasses();
}

// walk from top-level provider
$componentProviders = walkProvider(HttpComponentProvider::class, $parser);
foreach ($componentProviders as $providerClass) {
    $subProviders = walkProvider($providerClass, $parser);
    // categorize and extract...
}
```

**Annotation extraction (`#[Handler]` on controller methods):**

```php
// find methods with #[Handler] attribute
class HandlerAttributeVisitor extends NodeVisitorAbstract
{
    public function enterNode(Node $node): void
    {
        if (!$node instanceof Node\Stmt\ClassMethod) return;

        foreach ($node->attrGroups as $attrGroup) {
            foreach ($attrGroup->attrs as $attr) {
                if ($attr->name->toString() !== 'Handler') continue;

                // extract closure AST node
                $closureNode = $attr->args[0]->value;

                // pretty print closure back to source string
                $closureSource = $this->printer->prettyPrint([$closureNode]);

                // resolve types to FQN via use statement map
                $resolved = $this->resolveFQN($closureSource, $this->useStatements);

                // extract #[Parameter] annotations from same method
                $parameters = $this->extractParameters($node);

                $this->handlers[] = [
                    'handler'    => $resolved,
                    'parameters' => $parameters,
                ];
            }
        }
    }
}
```

**FQN resolution:**

```php
// collect use statements: 'UserController' => 'App\Http\Controllers\UserController'
function collectUseStatements(array $ast): array
{
    $map     = [];
    $visitor = new UseStatementVisitor();
    (new NodeTraverser())->addVisitor($visitor)->traverse($ast);

    foreach ($visitor->getUseStatements() as $use) {
        $parts = $use->uses[0]->name->parts;
        $alias = $use->uses[0]->alias?->name ?? end($parts);
        $map[$alias] = implode('\\', $parts);
    }

    return $map;
}

// rewrite closure source replacing short names with FQN
function resolveFQN(string $source, array $useMap): string
{
    foreach ($useMap as $alias => $fqn) {
        $source = preg_replace(
            '/\b' . preg_quote($alias) . '\b/',
            '\\' . $fqn,
            $source
        );
    }
    return $source;
}
```

**Route processor:**

```php
// construct ValkyrjaRoute from AST-extracted data, run through processor
$route = (new HttpRoute())
    ->setMethod($method)
    ->setPath($path)
    ->setParameters(array_map(
        fn($p) => new Parameter($p['name'], $p['pattern']),
        $parameters
    ));

$processedRoute = $processor->route($route);
$compiledRegex  = $processedRoute->getCompiledRegex();
```

---

### Java

**Dependencies:** Java annotation processor (`javax.annotation.processing`), Trees API (`com.sun.source.tree`), JavaPoet
for code generation

**File resolution:** Annotation processor has direct access to source files during `javac` via the Trees API — no
external file resolution needed.

**Annotation processor setup:**

```java

@SupportedAnnotationTypes("io.valkyrja.http.routing.Handler")
@SupportedSourceVersion(SourceVersion.RELEASE_21)
public class ValkyrjaAnnotationProcessor extends AbstractProcessor {

    private Trees trees;

    @Override
    public synchronized void init(ProcessingEnvironment env) {
        super.init(env);
        this.trees = Trees.instance(env);
    }

    @Override
    public boolean process(
            Set<? extends TypeElement> annotations,
            RoundEnvironment roundEnv
    ) {
        // collect all @Handler annotated methods
        for (Element element : roundEnv.getElementsAnnotatedWith(Handler.class)) {
            if (element.getKind() != ElementKind.METHOD) continue;
            processHandlerMethod((ExecutableElement) element);
        }
        return true;
    }
}
```

**Lambda source extraction via Trees API:**

```java
private void processHandlerMethod(ExecutableElement method) {
    // get the source tree for this method
    MethodTree methodTree = (MethodTree) trees.getTree(method);

    // find the @Handler annotation and extract lambda source text
    for (AnnotationMirror annotation : method.getAnnotationMirrors()) {
        if (!annotation.getAnnotationType().toString().equals(Handler.class.getName())) continue;

        // get the lambda argument from the annotation
        for (Map.Entry<? extends ExecutableElement, ? extends AnnotationValue> entry
                : annotation.getElementValues().entrySet()) {

            // extract source text of the lambda from annotation value
            String lambdaSource = entry.getValue().toString();

            // resolve all type references to FQN via element utilities
            String resolvedSource = resolveFQN(lambdaSource, method);

            // extract @Parameter annotations from same method
            List<ParameterData> parameters = extractParameters(method);

            handlers.add(new HandlerData(resolvedSource, parameters));
        }
    }
}
```

**FQN resolution via type utilities:**

```java
private String resolveFQN(String source, ExecutableElement method) {
    // get enclosing compilation unit for import resolution
    CompilationUnitTree unit = trees.getPath(method).getCompilationUnit();

    // build import map: simple name → fully qualified name
    Map<String, String> importMap = new HashMap<>();
    for (ImportTree imp : unit.getImports()) {
        String fqn = imp.getQualifiedIdentifier().toString();
        String simpleName = fqn.substring(fqn.lastIndexOf('.') + 1);
        importMap.put(simpleName, fqn);
    }

    // rewrite source replacing simple names with FQN
    for (Map.Entry<String, String> entry : importMap.entrySet()) {
        source = source.replaceAll(
                "\\b" + Pattern.quote(entry.getKey()) + "\\b",
                entry.getValue()
        );
    }
    return source;
}
```

**Cache data class generation via JavaPoet:**

```java
// generate AppHttpRoutingData record using JavaPoet
TypeSpec routingData = TypeSpec.recordBuilder("AppHttpRoutingData")
                .addModifiers(Modifier.PUBLIC, Modifier.FINAL)
                .addSuperinterface(HttpRoutingDataContract.class)
                .addRecordComponent(ParameterSpec.builder(
                        ParameterizedTypeName.get(Map.class, String.class, RouteContract.class),
                        "routes"
                ).build())
                .addRecordComponent(ParameterSpec.builder(
                        ParameterizedTypeName.get(Map.class, String.class,
                                ParameterizedTypeName.get(Map.class, String.class, String.class)),
                        "paths"
                ).build())
                // ... dynamicPaths, regexes
                .addMethod(MethodSpec.methodBuilder("create")
                        .addModifiers(Modifier.PUBLIC, Modifier.STATIC)
                        .returns(ClassName.get("", "AppHttpRoutingData"))
                        .addCode(generateCreateMethodBody(collectedRoutes))
                        .build())
                .build();

JavaFile.

builder("app.cache",routingData)
    .

build()
    .

writeTo(processingEnv.getFiler());
```

---

### Go

**Dependencies:** `go/analysis`, `go/ast`, `go/parser`, `go/token`, `go/types` — all standard library

**File resolution:** Go package paths map directly to directory paths. `go/packages.Load()` resolves package names to
source files.

**Package loading:**

```go
// load packages listed in valkyrja.yaml
cfg := &packages.Config{
Mode: packages.NeedFiles |
packages.NeedSyntax |
packages.NeedTypes |
packages.NeedImports,
}

pkgs, err := packages.Load(cfg, providerPackagePaths...)
if err != nil {
log.Fatalf("failed to load packages: %v", err)
}
```

**Provider method AST walk:**

```go
// walk Register() or GetRoutes() method body
func extractRoutes(pkg *packages.Package, typeName string) []RouteData {
var routes []RouteData

for _, file := range pkg.Syntax {
ast.Inspect(file, func (n ast.Node) bool {
funcDecl, ok := n.(*ast.FuncDecl)
if !ok || funcDecl.Name.Name != "GetRoutes" {
return true
}

// find the return statement
for _, stmt := range funcDecl.Body.List {
retStmt, ok := stmt.(*ast.ReturnStmt)
if !ok { continue }

// extract slice literal elements
compLit, ok := retStmt.Results[0].(*ast.CompositeLit)
if !ok { continue }

for _, elt := range compLit.Elts {
routes = append(routes, extractRouteFromNode(elt, file, pkg))
}
}
return false
})
}
return routes
}
```

**Handler function body extraction:**

```go
// extract function literal source text from AST node
func extractFuncLiteral(node ast.Node, fset *token.FileSet) string {
var buf bytes.Buffer
printer.Fprint(&buf, fset, node)
return buf.String()
}

// resolve imports to fully qualified package paths
func resolveFQN(source string, imports []*ast.ImportSpec) string {
for _, imp := range imports {
path := strings.Trim(imp.Path.Value, `"`)
alias := filepath.Base(path)
if imp.Name != nil {
alias = imp.Name.Name
}
source = strings.ReplaceAll(source, alias+".", path+"/")
}
return source
}
```

**Code generation:**

```go
// generate AppHttpRoutingData using go/format and text/template
const routingDataTemplate = `
package cache

import (
    "io/valkyrja/http/routing/data"
    "io/valkyrja/http/routing/data/contract"
)

// AppHttpRoutingData — generated by sindri, do not edit
var AppHttpRoutingData = data.HttpRoutingData{
    Routes: map[string]contract.RouteContract{
        {{- range .Routes}}
        "{{.Name}}": {{.RouteSource}},
        {{- end}}
    },
    Paths: map[string]map[string]string{
        {{- range $method, $paths := .Paths}}
        "{{$method}}": {
            {{- range $path, $key := $paths}}
            "{{$path}}": "{{$key}}",
            {{- end}}
        },
        {{- end}}
    },
    // ... dynamicPaths, regexes
}
`

tmpl := template.Must(template.New("routing").Parse(routingDataTemplate))
var buf bytes.Buffer
tmpl.Execute(&buf, collectedData)

// format the generated Go source
formatted, _ := format.Source(buf.Bytes())
os.WriteFile("app/cache/http_routing_data.go", formatted, 0644)
```

---

### Python

**Dependencies:** `ast` (standard library), `inspect` (standard library)

**File resolution:** `inspect.getfile(ClassName)` — equivalent of PHP's `ReflectionClass::getFileName()`. Works for any
importable class including framework classes.

**Provider tree walk:**

```python
import ast
import inspect


def walk_provider(provider_class: type) -> dict:
    """Walk a provider class and extract its sub-providers via AST."""
    filepath = inspect.getfile(provider_class)
    source = open(filepath).read()
    tree = ast.parse(source)

    # collect import map for FQN resolution
    import_map = collect_imports(tree)

    # find the provider method and extract its return list
    return extract_provider_list(tree, 'get_http_providers', import_map)


def collect_imports(tree: ast.Module) -> dict[str, str]:
    """Build a map of simple name → fully qualified module path."""
    import_map = {}
    for node in ast.walk(tree):
        if isinstance(node, ast.ImportFrom):
            module = node.module or ''
            for alias in node.names:
                name = alias.asname or alias.name
                import_map[name] = f"{module}.{alias.name}"
        elif isinstance(node, ast.Import):
            for alias in node.names:
                name = alias.asname or alias.name
                import_map[name] = alias.name
    return import_map


def extract_provider_list(
        tree: ast.Module,
        method_name: str,
        import_map: dict
) -> list[str]:
    """Extract the return value of a provider list method as FQN strings."""
    for node in ast.walk(tree):
        if not isinstance(node, ast.FunctionDef): continue
        if node.name != method_name: continue

        # find the return statement — must be a simple list literal
        for stmt in node.body:
            if not isinstance(stmt, ast.Return): continue
            if not isinstance(stmt.value, ast.List): continue

            classes = []
            for elt in stmt.value.elts:
                if isinstance(elt, ast.Name):
                    # resolve simple name to FQN
                    classes.append(import_map.get(elt.id, elt.id))
                elif isinstance(elt, ast.Attribute):
                    # already qualified
                    classes.append(ast.unparse(elt))
            return classes

    return []
```

**Decorator extraction (`@handler` on controller methods):**

```python
def extract_handlers(controller_class: type) -> list[dict]:
    """Extract @handler decorated methods from a controller class via AST."""
    filepath = inspect.getfile(controller_class)
    source = open(filepath).read()
    tree = ast.parse(source)
    imports = collect_imports(tree)
    handlers = []

    for node in ast.walk(tree):
        if not isinstance(node, ast.FunctionDef): continue

        for decorator in node.decorator_list:
            # find @handler(...) decorator
            if not isinstance(decorator, ast.Call): continue
            if not isinstance(decorator.func, ast.Name): continue
            if decorator.func.id != 'handler': continue

            # extract the lambda/closure argument
            if not decorator.args: continue
            handler_source = ast.unparse(decorator.args[0])
            handler_fqn = resolve_fqn(handler_source, imports)

            # extract @parameter decorators from same method
            parameters = extract_parameters(node, imports)

            handlers.append({
                'handler': handler_fqn,
                'parameters': parameters,
            })

    return handlers


def resolve_fqn(source: str, import_map: dict[str, str]) -> str:
    """Replace simple names with fully qualified names in source text."""
    import re
    for alias, fqn in import_map.items():
        source = re.sub(r'\b' + re.escape(alias) + r'\b', fqn, source)
    return source
```

**Code generation:**

```python
from string import Template

ROUTING_DATA_TEMPLATE = '''
# generated by sindri — do not edit
from valkyrja.http.routing.data import HttpRoutingData
$imports

APP_HTTP_ROUTING_DATA = HttpRoutingData(
    routes={
$routes
    },
    paths={
$paths
    },
    dynamic_paths={
$dynamic_paths
    },
    regexes={
$regexes
    },
)
'''


def generate_routing_data(collected: dict) -> str:
    routes_str = '\n'.join(
        f"        '{name}': {route_source},"
        for name, route_source in collected['routes'].items()
    )
    # ... build paths, dynamic_paths, regexes strings

    return Template(ROUTING_DATA_TEMPLATE).substitute(
        imports='\n'.join(collected['imports']),
        routes=routes_str,
        paths=build_paths_str(collected['paths']),
        dynamic_paths=build_paths_str(collected['dynamic_paths']),
        regexes=build_paths_str(collected['regexes']),
    )


output = generate_routing_data(collected_data)
open('app/cache/app_http_routing_data.py', 'w').write(output)
```

---

### TypeScript

**Dependencies:** `typescript` npm package (compiler API), standard Node.js `fs`

**File resolution:** TypeScript compiler API resolves module references to source files via `tsconfig.json` — no
separate file resolution step needed.

**Program setup:**

```typescript
import ts from 'typescript'
import * as fs from 'fs'

// load tsconfig and create compiler program
const configFile = ts.readConfigFile('tsconfig.json', ts.sys.readFile)
const config = ts.parseJsonConfigFileContent(
    configFile.config,
    ts.sys,
    process.cwd()
)

const program = ts.createProgram(config.fileNames, config.options)
const checker = program.getTypeChecker()
```

**Provider tree walk:**

```typescript
function extractProviderList(
    className: string,
    methodName: string,
    program: ts.Program
): string[] {
    const checker = program.getTypeChecker()
    const classes: string[] = []

    for (const sourceFile of program.getSourceFiles()) {
        if (sourceFile.isDeclarationFile) continue

        ts.forEachChild(sourceFile, function visit(node) {
            if (!ts.isClassDeclaration(node)) {
                ts.forEachChild(node, visit)
                return
            }
            if (node.name?.text !== className) {
                ts.forEachChild(node, visit)
                return
            }

            // find the provider list method
            for (const member of node.members) {
                if (!ts.isMethodDeclaration(member)) continue
                if ((member.name as ts.Identifier).text !== methodName) continue

                // find the return statement — must be a simple array literal
                const body = member.body
                if (!body) continue

                for (const stmt of body.statements) {
                    if (!ts.isReturnStatement(stmt)) continue
                    if (!stmt.expression) continue
                    if (!ts.isArrayLiteralExpression(stmt.expression)) continue

                    for (const element of stmt.expression.elements) {
                        if (ts.isIdentifier(element)) {
                            // resolve to fully qualified module path via type checker
                            const symbol = checker.getSymbolAtLocation(element)
                            const fqn = checker.getFullyQualifiedName(symbol!)
                            classes.push(fqn)
                        }
                    }
                }
            }
        })
    }
    return classes
}
```

**Handler method body extraction:**

```typescript
function extractPublisherMethod(
    className: string,
    methodName: string,
    program: ts.Program,
    sourceFile: ts.SourceFile
): string {
    const checker = program.getTypeChecker()
    let methodSource = ''

    ts.forEachChild(sourceFile, function visit(node) {
        if (!ts.isClassDeclaration(node) || node.name?.text !== className) {
            ts.forEachChild(node, visit)
            return
        }

        for (const member of node.members) {
            if (!ts.isMethodDeclaration(member)) continue
            if ((member.name as ts.Identifier).text !== methodName) continue

            // extract method body source text
            const body = member.body!
            const rawSource = sourceFile.text.slice(body.pos, body.end)

            // resolve all type references to fully qualified paths
            methodSource = resolveFQNTypes(rawSource, body, checker)
        }
    })

    return methodSource
}

function resolveFQNTypes(
    source: string,
    node: ts.Node,
    checker: ts.TypeChecker
): string {
    // walk identifiers in the node and replace with FQN via type checker
    ts.forEachChild(node, function visit(child) {
        if (ts.isIdentifier(child)) {
            const symbol = checker.getSymbolAtLocation(child)
            if (symbol) {
                const fqn = checker.getFullyQualifiedName(symbol)
                source = source.replace(child.text, fqn)
            }
        }
        ts.forEachChild(child, visit)
    })
    return source
}
```

**Code generation:**

```typescript
function generateRoutingData(collected: CollectedRouteData): string {
    const routesEntries = Object.entries(collected.routes)
        .map(([name, source]) => `    '${name}': ${source},`)
        .join('\n')

    const pathsEntries = Object.entries(collected.paths)
        .map(([method, paths]) => {
            const inner = Object.entries(paths)
                .map(([path, key]) => `        '${path}': '${key}',`)
                .join('\n')
            return `    '${method}': {\n${inner}\n    },`
        })
        .join('\n')

    return `// generated by sindri — do not edit
import { HttpRoutingData } from '@valkyrja/http/routing/data'
${collected.imports.join('\n')}

export const AppHttpRoutingData = new HttpRoutingData({
    routes: {
${routesEntries}
    },
    paths: {
${pathsEntries}
    },
    dynamicPaths: { /* ... */ },
    regexes: { /* ... */ },
})
`
}

fs.writeFileSync('app/cache/AppHttpRoutingData.ts', generateRoutingData(collected))
```

---

## Deployment Workflows

### Interpreted Languages (PHP, Python)

```bash
# development — no cache needed, full provider tree at runtime
php -S localhost:8000 public/index.php

# production CGI/lambda — cache required
sindri generate
deploy

# production worker — cache optional
deploy
```

### Compiled Languages (Java, Go, TypeScript)

```bash
# Java
sindri generate    # generates cache source files
mvn compile                # single pass — compiles with generated files
deploy target/app.jar

# Go
sindri generate    # generates cache source files
go build -o app            # compiles with generated files
deploy app

# TypeScript
sindri generate    # generates cache source files
tsc                        # compiles with generated files
deploy dist/
```

### CI/CD Pipeline

```yaml
steps:
  - name: Install dependencies
    run: # language-specific install

  - name: Generate cache data files
    run: sindri generate

  - name: Build (compiled languages only)
    run: # mvn compile / go build / tsc

  - name: Deploy
    run: deploy
```

---

## The Build Tool Bootstrapping Itself

`sindri` is a Valkyrja application subject to the same rules as any other Valkyrja application. It ships without its own
cache files because it runs at deploy time rather than per-request. For environments where build tool startup time is a
concern, it can generate its own cache and rebuild:

```bash
# pass 1 — build without cache (slow, one-time)
sindri build-self --output build-tool-bootstrap

# pass 2 — generate the build tool's own cache
./build-tool-bootstrap generate --self

# pass 3 — rebuild with cache (fast all subsequent runs)
sindri build-self --with-cache --output sindri
```

This is the same two-pass process the build tool applies to compiled language applications. The build tool bootstrapping
itself is a validation that the framework's cache-optional architecture is self-consistent — no special cases, no
exemptions.

---

## Current Implementation Status

| Language   | Without cache | Cache generation                                            | Notes                                                 |
|------------|---------------|-------------------------------------------------------------|-------------------------------------------------------|
| PHP        | ✅ works       | ⚠️ CLI command exists — will break when handler logic ships | Migrate to sindri before handler logic implementation |
| Java       | ✅ works       | ❌ not yet built                                             | sindri Java AST implementation pending                |
| Go         | ✅ works       | ❌ not yet built                                             | sindri Go AST implementation pending                  |
| Python     | ✅ works       | ❌ not yet built                                             | sindri Python AST implementation pending              |
| TypeScript | ✅ works       | ❌ not yet built                                             | sindri TypeScript compiler API implementation pending |

The PHP CLI command is the most pressing TODO. It will stop working correctly once closure-based handler logic replaces
the current dispatch-based routing — the existing serialization mechanism cannot handle closures. The migration to
`sindri` and `#[Handler]` annotation extraction needs to happen before handler logic ships in PHP.

---

## Framework Source Shipping Policy

The build tool requires access to provider source files to extract bindings and handlers. Each language must ship source
accordingly:

| Language   | Source shipping  | Requirement                                                              |
|------------|------------------|--------------------------------------------------------------------------|
| PHP        | Composer package | Always present on disk — no special requirement                          |
| Java       | Maven / Gradle   | Must publish `-sources.jar` as a **required** build dependency           |
| Go         | Go modules       | Full source downloaded via `go mod download` — always present            |
| Python     | pip package      | Always present on disk — no special requirement                          |
| TypeScript | npm package      | Must ship `.ts` source files alongside compiled `.js` — not just `.d.ts` |

Third-party packages built on Valkyrja must follow the same policy to support full cache generation for their bindings.

---

## PHP Build Tool — valkyrja/sindri (formerly Bin)

The PHP implementation of the build tool ships as its own standalone repository and Composer package — separate from the
Valkyrja framework itself. It was previously part of the framework as the `Bin` component.

**Package:** `sindri` (or `valkyrja/sindri`)
**Repository:** separate from `valkyrja/framework`
**Composer dependency:** `nikic/php-parser`

---

### Why Bin Leaves the Framework

The `Bin` component was always conceptually a development and build-time concern — file generation, project scaffolding,
cache generation. None of this is a runtime concern. Keeping it in the framework meant:

- `nikic/php-parser` was a framework dependency, pulled into every application even in production
- File generation scaffolding lived in runtime code where it had no business being
- The framework carried tooling weight that only mattered at build/dev time
- No clean equivalent existed for other language ports — each language would need its own bin component inside the
  framework

Moving it out:

- `nikic/php-parser` becomes a `sindri` dependency, never present in production applications
- The framework base code is simplified — no file generation, no scaffolding, no AST tooling
- Each language port gets its own build tool repository following the same pattern
- The separation of runtime vs build-time concerns is clean and explicit

---

### What valkyrja/sindri Provides

**Cache generation**

- Reads the application `AppConfig` class
- Walks provider tree via AST
- Extracts `#[Handler]` annotations and explicit route definitions
- Runs `ProcessorContract::route()` for regex compilation
- Generates `AppContainerData`, `AppEventData`, `AppHttpRoutingData`, `AppCliRoutingData`

**Project scaffolding**

- `valkyrja new project-name` — creates a blank Valkyrja application with the correct directory structure
- `valkyrja make:provider ProviderName` — generates a blank service provider
- `valkyrja make:controller ControllerName` — generates a blank controller with example `#[Handler]`
- `valkyrja make:listener ListenerName` — generates a blank event listener
- `valkyrja make:command CommandName` — generates a blank CLI command

**All file generation that was previously in Bin**

- Anything that writes files to disk lives here, not in the framework

---

### Installation

```bash
# install as a dev dependency — never needed in production
composer require --dev sindri
```

```json
// composer.json
{
  "require"    : {
    "valkyrja/framework": "^26.0"
  },
  "require-dev": {
    "sindri": "^1.0"
  }
}
```

---

### Usage

```bash
# generate cache data files for production
./vendor/bin/valkyrja generate

# create a new project
composer create-project valkyrja/project my-app

# scaffold files
./vendor/bin/valkyrja make:controller UserController
./vendor/bin/valkyrja make:provider UserServiceProvider
./vendor/bin/valkyrja make:listener UserCreatedListener
./vendor/bin/valkyrja make:command SendEmailCommand
```

---

### The Build Tool Ecosystem

Each language port ships its own build tool as a separate package following the same pattern:

| Language   | Build tool package                        | AST dependency                   |
|------------|-------------------------------------------|----------------------------------|
| PHP        | `sindri` (separate repo, formerly Bin)    | `nikic/php-parser`               |
| Java       | `io.valkyrja:sindri` (separate artifact)  | Trees API (built into javac)     |
| Go         | `io/valkyrja/sindri` (separate module)    | `go/analysis`, `go/ast` (stdlib) |
| Python     | `sindri` (separate PyPI package)          | `ast`, `inspect` (stdlib)        |
| TypeScript | `@valkyrja/sindri` (separate npm package) | `typescript` compiler API        |

In all cases the build tool is a dev/build dependency only — never present in production runtime. The framework packages
have zero AST or build tooling dependencies.

---

## Discussion Summary

The build tool's design emerged from the need to generate cache data files for production CGI and lambda deployments
across all five language ports. Several approaches were considered and rejected before arriving at the current design.

The first approach considered was the two-pass compile — build a bootstrap binary, run it to generate cache, compile
again with the generated files. This works but adds a compile step and requires a bootstrap binary to exist before the
cache can be generated.

The key insight that eliminated the two-pass compile was recognizing that the application config class + AST analysis
gives the build tool everything it needs without running the application. The config class already lists top-level
providers — the same class the developer uses to create the application. The AST walker reads provider list method
return values as static data — no execution required. The build tool can discover the complete provider tree, extract
all handlers and bindings, run the framework's own `ProcessorContract` for regex compilation, and generate fully
resolved cache data files in a single pass before the first compile.

A separate `valkyrja.yaml` was considered as the build tool entry point but superseded by the config class approach —
eliminating a duplicate source of truth. The component provider constants class was also dropped as it would allow
constant references in provider lists that the build tool cannot resolve statically.

The build tool as a text generator insight resolved the custom route child class problem. The build tool doesn't need
application classes compiled in to reference them by name in generated source — exactly like a developer writing
`extends AuthenticatedRoute` in an editor. Class names from AST are strings. Generated files are strings. The compiler
resolves them later.

The four-output-class design emerged from recognizing that one class per provider would require the framework to merge N
structures at boot. One class per concern for the entire application means the framework loads exactly four objects and
the data is immediately ready to use. The build tool does the aggregation work once at generation time.

The self-bootstrapping property — the build tool running sindri on itself to generate its own cache — was identified as
a validation of the architecture's self-consistency rather than a practical requirement. It demonstrates that no special
cases exist in the framework design.

The PHP CLI command breaking change was identified as the most pressing near-term issue. It is the only currently
working cache generation mechanism, and it will stop working when closure-based handler logic replaces dispatch-based
routing. This migration is documented as a TODO that must happen before handler logic ships in PHP.

The separation of `Bin` from the framework into its own `sindri` repository was decided when it became clear that the
build tool needed `nikic/php-parser` as a dependency. Keeping this in the framework would mean a parser library — a
purely development and build-time concern — would be a production dependency of every Valkyrja application. Moving it
out keeps the framework clean and the separation of runtime vs build-time concerns explicit.

The move also revealed that `Bin`'s file generation and scaffolding features were always development tooling, not
framework concerns. Every file generation feature that lived in the framework now lives in the build tool where it
belongs. This simplifies the framework base code and makes the runtime package leaner.

The pattern established for PHP — separate build tool repository, dev-only dependency, no AST tooling in the framework —
is replicated across all five language ports. In all cases the framework ships with zero build or AST dependencies. The
build tool for each language is an optional dev dependency that applications install during development but never ship
to production.
