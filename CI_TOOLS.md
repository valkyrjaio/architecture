# CI Tools

CI toolchain for all five Valkyrja language ports. Tools are grouped by role so equivalents across languages are easy to
identify. Gaps are noted where a language lacks a strong equivalent for a role that other languages cover well.

---

## Role Categories

| Role                         | Description                                                                              |
|------------------------------|------------------------------------------------------------------------------------------|
| **Architecture enforcement** | Validates structural rules — dependency direction, layer boundaries, package constraints |
| **Static analysis**          | Bug detection, type inference, code correctness without execution                        |
| **Security analysis**        | Vulnerability scanning, OWASP patterns, taint analysis                                   |
| **Formatting**               | Code style and formatting — automated, enforced in CI                                    |
| **Automated migration**      | Automated refactoring, version upgrades, deprecation removal                             |
| **Dead code detection**      | Unused exports, files, dependencies, variables                                           |
| **Testing**                  | Unit and integration test runner                                                         |

---

## PHP

| Tool         | Role                       | Notes                                                            |
|--------------|----------------------------|------------------------------------------------------------------|
| PHPArkitect  | Architecture enforcement   | Enforces layer boundaries, dependency direction, naming rules    |
| PHPStan      | Static analysis            | Type inference, null safety, dead code, level 0–9 strictness     |
| Psalm        | Static analysis + security | Type safety, taint analysis for security vulnerabilities         |
| PHP-CS-Fixer | Formatting                 | Enforces PSR-12 and custom rules, auto-fixes                     |
| Rector       | Automated migration        | PHP version upgrades, deprecation removal, automated refactoring |
| PHPUnit      | Testing                    | Standard PHP test runner                                         |

PHP has the most complete toolchain of all five languages. PHPStan and Psalm overlap but complement — PHPStan is
stronger on type inference, Psalm is stronger on security taint analysis.

---

## Java

| Tool                          | Role                       | Notes                                                                                          |
|-------------------------------|----------------------------|------------------------------------------------------------------------------------------------|
| ArchUnit                      | Architecture enforcement   | Enforces package dependencies, layer rules, naming conventions as JUnit tests                  |
| ErrorProne + NullAway         | Static analysis            | Google's compiler plugin — 400+ bug patterns caught at compile time. NullAway adds null safety |
| SpotBugs + FindSecBugs        | Static analysis + security | Bytecode analysis post-compilation. FindSecBugs adds OWASP/CWE security checks                 |
| Spotless (Google Java Format) | Formatting                 | Enforces Google Java Format via Gradle/Maven plugin                                            |
| OpenRewrite                   | Automated migration        | Java version upgrades, framework migrations, automated refactoring. Rector equivalent          |
| JUnit 5                       | Testing                    | Standard Java test runner                                                                      |

**ErrorProne vs SpotBugs:** <br>ErrorProne runs during compilation (catches errors before bytecode exists). SpotBugs
runs on compiled bytecode (finds different patterns, particularly concurrency and null dereference). The overlap is
partial — both are worth running. FindSecBugs is a SpotBugs plugin, not a separate tool.

**OpenRewrite** fills the Rector gap in the original Java toolchain. It handles Java LTS migrations, Spring Boot
upgrades, dependency updates, and automated refactoring with a recipe-based system.

---

## Go

| Tool              | Role                      | Notes                                                                                                                                              |
|-------------------|---------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| golangci-lint     | Everything except testing | Meta-linter bundling 50+ tools: staticcheck, go vet, errcheck, gosec, revive, go-cleanarch and more. One binary, one config file (`.golangci.yml`) |
| go test           | Testing                   | Built into the language — no separate tool needed                                                                                                  |
| gofmt / goimports | Formatting                | Built into the language — gofmt is the standard, goimports adds import management                                                                  |

Go's toolchain is the simplest of all five languages. golangci-lint is the de facto standard used by Kubernetes,
Prometheus, and Terraform. It covers every role except testing in a single tool:

- **Architecture:** `go-cleanarch` linter (dependency rules, clean architecture validation)
- **Static analysis:** `staticcheck` (150+ checks), `go vet` (compiler-level), `errcheck` (unchecked errors)
- **Security:** `gosec` (OWASP patterns, injection risks, hardcoded credentials)
- **Dead code:** `unused` linter
- **Formatting:** `gofmt` is enforced via `gofmt` linter

No Rector/OpenRewrite equivalent — Go doesn't accumulate the same version migration debt. The language and stdlib are
stable and backwards-compatible.

```yaml
# .golangci.yml — recommended baseline
linters:
  enable:
    - staticcheck
    - govet
    - errcheck
    - gosec
    - revive
    - go-cleanarch
    - unused
    - misspell
    - gofmt
    - goimports
```

---

## Python

| Tool          | Role                         | Notes                                                                                                                            |
|---------------|------------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| import-linter | Architecture enforcement     | Enforces import boundaries and layer contracts. Less powerful than PHPArkitect/ArchUnit — documented gap                         |
| Ruff          | Static analysis + formatting | Replaces flake8, black, isort, and many flake8 plugins. Written in Rust — 10-100x faster. Single tool for linting and formatting |
| mypy          | Type checking                | Static type analysis against PEP 484 type hints. Validates type correctness across the entire codebase                           |
| Bandit        | Security analysis            | AST-based security scanner — hardcoded credentials, injection risks, insecure function use                                       |
| pytest        | Testing                      | Standard Python test runner                                                                                                      |

**Architecture enforcement gap:** No Python equivalent of PHPArkitect or ArchUnit exists with the same power.
`import-linter` enforces import contracts but is limited to import graph analysis. This is a known gap — enforce
architectural rules through code review and provider tree conventions instead.

**No automated migration tool:** No Rector/OpenRewrite equivalent for Python. Migration scripts are written manually.
Not a significant gap for Valkyrja since the framework requires Python 3.14 minimum and won't accumulate legacy
migration debt.

**Ruff vs mypy:** These are complementary, not competing. Ruff handles style, lint rules, and formatting. mypy handles
type correctness. Both are required.

```toml
# pyproject.toml — recommended baseline
[tool.ruff]
select = ["E", "F", "I", "N", "B", "S", "UP", "RUF"]
line-length = 120

[tool.mypy]
strict = true
python_version = "3.14"

[tool.pytest.ini_options]
testpaths = ["tests"]
```

---

## TypeScript

| Tool                       | Role                | Notes                                                                                                                                 |
|----------------------------|---------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| `tsc --noEmit`             | Type checking       | The TypeScript compiler itself — full type-aware analysis across all files                                                            |
| ESLint + typescript-eslint | Static analysis     | Type-aware linting rules: no-floating-promises, no-unsafe-*, exhaustive checks. The backbone of production-grade TypeScript CI        |
| Biome                      | Formatting          | Prettier replacement written in Rust — significantly faster, single binary for format + basic lint                                    |
| Knip                       | Dead code detection | Finds unused exports, files, dependencies. Helps Vercel delete ~300k lines of unused code. No equivalent in other language toolchains |
| Vitest                     | Testing             | Modern Vite-native test runner. Jest is the alternative for non-Vite projects                                                         |

**Architecture enforcement gap:** No strong equivalent of PHPArkitect or ArchUnit for TypeScript. `eslint-plugin-import`
can enforce some import boundary rules but is not as expressive. This is a documented gap — enforce architectural rules
through code review and TypeScript's module system.

**Security gap:** No dedicated security scanner equivalent to Psalm taint analysis or Bandit for TypeScript. ESLint
rules catch some patterns (e.g. `no-eval`) but are not a security tool. Semgrep can be added for deeper security
analysis if needed.

**ESLint vs Biome:** These are complementary. Biome handles formatting and basic lint rules with superior speed.
ESLint + typescript-eslint handles type-aware rules that Biome cannot yet replicate (type information is required). Both
are needed.

```json
// tsconfig.json — strict baseline
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "exactOptionalPropertyTypes": true
  }
}
```

---

## Gaps Summary

| Role                     | PHP               | Java                    | Go                       | Python                     | TypeScript                        |
|--------------------------|-------------------|-------------------------|--------------------------|----------------------------|-----------------------------------|
| Architecture enforcement | ✅ PHPArkitect     | ✅ ArchUnit              | ✅ go-cleanarch           | ⚠️ import-linter (limited) | ⚠️ eslint-plugin-import (limited) |
| Static analysis          | ✅ PHPStan + Psalm | ✅ ErrorProne + SpotBugs | ✅ golangci-lint          | ✅ mypy + Ruff              | ✅ tsc + typescript-eslint         |
| Security                 | ✅ Psalm taint     | ✅ FindSecBugs           | ✅ gosec                  | ✅ Bandit                   | ⚠️ no dedicated tool              |
| Formatting               | ✅ PHP-CS-Fixer    | ✅ Spotless              | ✅ gofmt (built-in)       | ✅ Ruff                     | ✅ Biome                           |
| Automated migration      | ✅ Rector          | ✅ OpenRewrite           | —                        | —                          | —                                 |
| Dead code                | —                 | —                       | ✅ unused (golangci-lint) | —                          | ✅ Knip                            |
| Testing                  | ✅ PHPUnit         | ✅ JUnit 5               | ✅ go test (built-in)     | ✅ pytest                   | ✅ Vitest                          |

**Key gaps:**

- **Architecture enforcement** is weak in Python and TypeScript. No tool matches the expressiveness of PHPArkitect or
  ArchUnit. Enforce via code review and framework conventions.
- **Security scanning** is absent from TypeScript. Bandit, gosec, and FindSecBugs have no direct equivalent. Add Semgrep
  for deeper TypeScript security analysis if required.
- **Automated migration** is only covered by PHP (Rector) and Java (OpenRewrite). Go, Python, and TypeScript have no
  equivalent — less of an issue since these ports start fresh on current language versions.
- **Dead code detection** is only strong in Go (golangci-lint) and TypeScript (Knip). PHP, Java, and Python lack a
  dedicated tool — IDEs and some lint rules partially cover this.

---

## Per-Language Toolchain Summary

```
PHP         — PHPArkitect + PHPStan + Psalm + PHP-CS-Fixer + Rector + PHPUnit
Java        — ArchUnit + ErrorProne/NullAway + SpotBugs/FindSecBugs + Spotless + OpenRewrite + JUnit 5
Go          — golangci-lint (all-in-one) + go test + gofmt
Python      — import-linter + Ruff + mypy + Bandit + pytest
TypeScript  — tsc + ESLint/typescript-eslint + Biome + Knip + Vitest
```
