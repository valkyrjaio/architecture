package io.valkyrja.volundr.analysis;

import com.github.javaparser.StaticJavaParser;
import com.github.javaparser.ast.CompilationUnit;
import com.github.javaparser.ast.body.ClassOrInterfaceDeclaration;
import com.github.javaparser.ast.body.MethodDeclaration;
import com.github.javaparser.ast.expr.ArrayInitializerExpr;
import com.github.javaparser.ast.expr.ClassExpr;
import com.github.javaparser.ast.expr.Expression;
import com.github.javaparser.ast.expr.MethodCallExpr;
import com.github.javaparser.ast.stmt.ReturnStmt;
import com.github.javaparser.ast.type.ClassOrInterfaceType;
import com.github.javaparser.resolution.UnsolvedSymbolException;
import com.github.javaparser.symbolsolver.JavaSymbolSolver;
import com.github.javaparser.symbolsolver.resolution.typesolvers.CombinedTypeSolver;
import com.github.javaparser.symbolsolver.resolution.typesolvers.JavaParserTypeSolver;
import com.github.javaparser.symbolsolver.resolution.typesolvers.ReflectionTypeSolver;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;

/**
 * Walks a ComponentProviderContract implementation and extracts the class
 * references returned from each get*Providers method, resolving each to a
 * source file on disk.
 *
 * Assumes the Java port keeps the PHP invariant: each get*Providers method
 * returns a literal array (or List.of(...)) of class literals.
 */
public final class ProviderWalker {

    private static final Map<String, String> METHODS = Map.of(
        "componentProviders", "getComponentProviders",
        "containerProviders", "getContainerProviders",
        "eventProviders",     "getEventProviders",
        "cliProviders",       "getCliProviders",
        "httpProviders",      "getHttpProviders"
    );

    private final List<Path> sourceRoots;

    public ProviderWalker(List<Path> sourceRoots) {
        this.sourceRoots = List.copyOf(sourceRoots);

        // A CombinedTypeSolver lets JavaParser resolve type references to
        // their FQNs using the same roots we'll use to locate files.
        CombinedTypeSolver typeSolver = new CombinedTypeSolver();
        typeSolver.add(new ReflectionTypeSolver());
        for (Path root : this.sourceRoots) {
            typeSolver.add(new JavaParserTypeSolver(root.toFile()));
        }
        StaticJavaParser.getParserConfiguration()
            .setSymbolResolver(new JavaSymbolSolver(typeSolver));
    }

    public Result analyze(Path file) throws IOException {
        CompilationUnit cu = StaticJavaParser.parse(file);
        Result result = new Result();

        Optional<ClassOrInterfaceDeclaration> decl = cu.findFirst(ClassOrInterfaceDeclaration.class);
        decl.ifPresent(d -> result.className = d.getFullyQualifiedName().orElse(null));
        if (decl.isEmpty()) {
            return result;
        }

        for (Map.Entry<String, String> entry : METHODS.entrySet()) {
            decl.get().getMethodsByName(entry.getValue()).stream().findFirst().ifPresent(method -> {
                List<ProviderRef> refs = extractClassLiterals(method).stream()
                    .map(fqn -> new ProviderRef(fqn, resolveToFile(fqn)))
                    .toList();
                result.providers.put(entry.getKey(), refs);
            });
            result.providers.putIfAbsent(entry.getKey(), List.of());
        }

        return result;
    }

    /**
     * Pull every `SomeClass.class` expression out of return statements in the
     * method and return their FQNs. Handles both `return new Class[] { A.class }`
     * and `return List.of(A.class, B.class)` patterns.
     */
    private List<String> extractClassLiterals(MethodDeclaration method) {
        List<String> fqns = new ArrayList<>();
        for (ReturnStmt ret : method.findAll(ReturnStmt.class)) {
            ret.getExpression().ifPresent(expr -> collectClassExprs(expr, fqns));
        }
        return fqns;
    }

    private void collectClassExprs(Expression expr, List<String> out) {
        // new Class[] { A.class, B.class } or { A.class, B.class }
        if (expr instanceof ArrayInitializerExpr arr) {
            arr.getValues().forEach(v -> collectClassExprs(v, out));
            return;
        }
        // List.of(A.class, ...), Arrays.asList(A.class, ...), etc.
        if (expr instanceof MethodCallExpr call) {
            call.getArguments().forEach(a -> collectClassExprs(a, out));
            return;
        }
        if (expr instanceof ClassExpr classExpr && classExpr.getType() instanceof ClassOrInterfaceType type) {
            // Try resolution first for accuracy, fall back to textual name.
            try {
                out.add(type.resolve().asReferenceType().getQualifiedName());
            } catch (UnsolvedSymbolException | RuntimeException e) {
                out.add(type.getNameWithScope());
            }
        }
    }

    /**
     * Java's package-to-path mapping is deterministic: `a.b.C` lives at
     * `<root>/a/b/C.java`. We check each configured source root.
     */
    private Path resolveToFile(String fqn) {
        String relative = fqn.replace('.', '/') + ".java";
        for (Path root : sourceRoots) {
            Path candidate = root.resolve(relative);
            if (Files.isRegularFile(candidate)) {
                return candidate.toAbsolutePath().normalize();
            }
        }
        return null;
    }

    public static final class Result {
        public String className;
        public final Map<String, List<ProviderRef>> providers = new LinkedHashMap<>();
    }

    public record ProviderRef(String fqn, Path file) {}
}
