/**
 * Walks a TypeScript implementation of ComponentProviderContract and
 * extracts the classes returned from each get*Providers method, resolving
 * each reference back to a source file via the TS module resolver.
 *
 * The TS port is assumed to use:
 *
 *   import { FooProvider } from "./foo";
 *   export class MyComponentProvider implements ComponentProviderContract {
 *     static getComponentProviders(app: Application): Array<ComponentProviderCtor> {
 *       return [FooProvider, BarProvider];
 *     }
 *     // ...
 *   }
 *
 * We extract the identifiers from the return array and use the TS compiler's
 * symbol table to find their declaration file.
 */
import * as path from "node:path";
import * as ts from "typescript";

const METHODS = {
    componentProviders: "getComponentProviders",
    containerProviders: "getContainerProviders",
    eventProviders:     "getEventProviders",
    cliProviders:       "getCliProviders",
    httpProviders:      "getHttpProviders",
} as const;

type MethodKey = keyof typeof METHODS;

export interface ProviderRef {
    name: string;       // local name as written in source
    file: string | null;
}

export type AnalysisResult = {
    className: string | null;
} & Record<MethodKey, ProviderRef[]>;

export class ProviderWalker {
    private readonly program: ts.Program;
    private readonly checker: ts.TypeChecker;

    /**
     * Build a Program from the project's tsconfig.json. The Program handles
     * path mappings, baseUrl, node_modules, and .d.ts resolution for us.
     */
    constructor(tsconfigPath: string) {
        const configFile = ts.readConfigFile(tsconfigPath, ts.sys.readFile);
        if (configFile.error) {
            throw new Error(ts.flattenDiagnosticMessageText(configFile.error.messageText, "\n"));
        }
        const parsed = ts.parseJsonConfigFileContent(
            configFile.config,
            ts.sys,
            path.dirname(tsconfigPath),
        );
        this.program = ts.createProgram({
            rootNames: parsed.fileNames,
            options:   parsed.options,
        });
        this.checker = this.program.getTypeChecker();
    }

    analyze(filePath: string): AnalysisResult {
        const sourceFile = this.program.getSourceFile(filePath);
        if (!sourceFile) {
            throw new Error(`Source file not in program: ${filePath}`);
        }

        const result: AnalysisResult = {
            className:          null,
            componentProviders: [],
            containerProviders: [],
            eventProviders:     [],
            cliProviders:       [],
            httpProviders:      [],
        };

        // Find the (first) class declaration in the file.
        const classDecl = sourceFile.statements.find(
            (s): s is ts.ClassDeclaration => ts.isClassDeclaration(s) && !!s.name,
        );
        if (!classDecl?.name) {
            return result;
        }
        result.className = classDecl.name.text;

        for (const [key, methodName] of Object.entries(METHODS) as [MethodKey, string][]) {
            const method = classDecl.members.find(
                (m): m is ts.MethodDeclaration =>
                    ts.isMethodDeclaration(m) &&
                    ts.isIdentifier(m.name) &&
                    m.name.text === methodName,
            );
            if (!method?.body) continue;

            result[key] = this.extractRefs(method);
        }

        return result;
    }

    /**
     * Find all `return [Foo, Bar]` statements in the method and collect each
     * array element identifier, resolving each to its source file.
     */
    private extractRefs(method: ts.MethodDeclaration): ProviderRef[] {
        const refs: ProviderRef[] = [];

        const visit = (node: ts.Node): void => {
            if (ts.isReturnStatement(node) && node.expression && ts.isArrayLiteralExpression(node.expression)) {
                for (const el of node.expression.elements) {
                    // We only care about bare identifiers (class references) —
                    // computed expressions break the static invariant and are
                    // intentionally ignored.
                    if (ts.isIdentifier(el)) {
                        refs.push({
                            name: el.text,
                            file: this.resolveIdentifier(el),
                        });
                    }
                }
                return;
            }
            ts.forEachChild(node, visit);
        };
        visit(method.body!);
        return refs;
    }

    /**
     * Use the type checker to find the symbol's declaration. The checker has
     * already done all the module resolution work during program creation.
     */
    private resolveIdentifier(id: ts.Identifier): string | null {
        const symbol = this.checker.getSymbolAtLocation(id);
        if (!symbol) return null;

        // For imported symbols, getAliasedSymbol gives us the original declaration.
        const resolved = (symbol.flags & ts.SymbolFlags.Alias)
            ? this.checker.getAliasedSymbol(symbol)
            : symbol;

        const decl = resolved.declarations?.[0];
        if (!decl) return null;

        return path.resolve(decl.getSourceFile().fileName);
    }
}
