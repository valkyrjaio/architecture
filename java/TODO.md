# Java

```
./.github/ci/spotless/gradlew spotlessApply   # auto-format
./.github/ci/spotless/gradlew    # check without modifying

./gradlew spotlessCheck   # check formatting
./gradlew spotlessApply   # apply formatting
./gradlew archunit        # architecture tests
./gradlew errorprone      # static analysis
./gradlew spotbugs        # bug detection
./gradlew junit           # unit testing
./gradlew ci              # run all checks
```

## TODOs

### Publishing to Maven Central

Publishing must happen in order: **framework first**, then the entry modules (they depend on it).

#### 1. One-time setup

1. Register at https://central.sonatype.com and claim the `io.valkyrja` namespace. DONE
2. Generate a GPG key pair and upload the public key to a keyserver: How?
   ```
   gpg --gen-key
   gpg --keyserver keyserver.ubuntu.com --send-keys <KEY_ID>
   ```
3. Export credentials to `~/.gradle/gradle.properties` (never commit these):
   ```properties
   mavenCentralUsername=<sonatype-token-username>
   mavenCentralPassword=<sonatype-token-password>
   signing.keyId=<last-8-chars-of-key-id>
   signing.password=<gpg-passphrase>
   signing.secretKeyRingFile=/path/to/secring.gpg
   ```

#### 2. Add publishing config to each `build.gradle.kts`

Add these plugins and blocks to `framework/build.gradle.kts` and each `entry/*/build.gradle.kts`:

```kotlin
plugins {
    java
    `maven-publish`
    signing
}

java {
    withJavadocJar()
    withSourcesJar()
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            from(components["java"])
            pom {
                name.set("Valkyrja <Module>")          // e.g. "Valkyrja Tomcat"
                description.set("<short description>")
                url.set("https://github.com/valkyrjaio/valkyrja-java")
                licenses {
                    license {
                        name.set("MIT License")
                        url.set("https://opensource.org/licenses/MIT")
                    }
                }
                developers {
                    developer {
                        id.set("melechmizrachi")
                        name.set("Melech Mizrachi")
                        email.set("melechmizrachi@gmail.com")
                    }
                }
                scm {
                    connection.set("scm:git:git://github.com/valkyrjaio/valkyrja-java.git")
                    developerConnection.set("scm:git:ssh://github.com/valkyrjaio/valkyrja-java.git")
                    url.set("https://github.com/valkyrjaio/valkyrja-java")
                }
            }
        }
    }
    repositories {
        maven {
            name = "MavenCentral"
            url = uri("https://central.sonatype.com/api/v1/publisher/upload")
        }
    }
}

signing {
    sign(publishing.publications["maven"])
}
```

Artifact coordinates per module:
- `framework`       → `io.valkyrja:valkyrja:<version>`
- `entry/tomcat`    → `io.valkyrja:valkyrja-tomcat:<version>`
- `entry/netty`     → `io.valkyrja:valkyrja-netty:<version>`
- `entry/jetty`     → `io.valkyrja:valkyrja-jetty:<version>`

#### 3. Publish

```
# Publish framework first
cd framework && ./gradlew publishToMavenCentral

# Then entry modules (after framework is live and resolvable)
cd entry/tomcat && ./gradlew publishToMavenCentral
cd entry/netty  && ./gradlew publishToMavenCentral
cd entry/jetty  && ./gradlew publishToMavenCentral
```

---

- **`entry/*/settings.gradle.kts` — replace `includeBuild` with a real Maven coordinate once published**
  Each entry module (`tomcat`, `netty`, `jetty`) currently references the framework via
  `includeBuild("../../framework")`, which only works in this local monorepo layout. Once
  `io.valkyrja:valkyrja` is published to Maven Central (or a private registry), the
  `includeBuild` line must be removed and the framework resolved purely as a versioned
  dependency — exactly as `application` does in production via
  `implementation("io.valkyrja:valkyrja:26.0.0")`.
