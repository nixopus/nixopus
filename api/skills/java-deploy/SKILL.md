---
name: java-deploy
description: Build and deploy Java applications — Maven, Gradle, Spring Boot, version detection, and Dockerfile patterns. Use when deploying a Java project, or when pom.xml or build.gradle is detected.
metadata:
  version: "1.0"
---

# Java Deployment

## Detection

Project is Java if:

- **Maven**: `pom.xml` exists at the root
- **Gradle**: `gradlew` or `build.gradle` / `build.gradle.kts` exists at the root

## Versions

### Java / JDK

1. `pom.xml` → `maven.compiler.source` or `maven.compiler.target`
2. `build.gradle` → `sourceCompatibility` or `JavaLanguageVersion`
3. `.sdkmanrc` or `.java-version` file
4. Defaults to **21**

### Gradle

Read from `gradle/wrapper/gradle-wrapper.properties` → `distributionUrl`.

## Build

### Maven

- `mvn -B package -DskipTests`
- Output: `target/<artifact-id>-<version>.jar`
- Spring Boot: `target/*.jar` (fat JAR)
- Offline deps: `mvn -B dependency:go-offline` for layer caching

### Gradle

- `./gradlew build -x test` (prefer wrapper when `gradlew` exists)
- Use `--no-daemon` in Docker
- Output: `build/libs/<name>-<version>.jar`
- Spring Boot: `build/libs/*.jar` (fat JAR)
- Warm deps: `./gradlew dependencies --no-daemon` for layer caching

### Start Command

- `java -jar <jar-file>`
- Spring Boot: `java -jar app.jar`

## Port Detection

1. `server.port` in `application.properties` or `application.yml`
2. `PORT` or `SERVER_PORT` environment
3. Spring Boot default: 8080
4. Default: **8080**

## Framework Detection

| Dependency / config | Framework |
|---|---|
| `spring-boot-starter-web` | Spring Boot |
| `spring-boot-starter-webflux` | Spring Boot (reactive) |
| `quarkus` | Quarkus |
| `micronaut` | Micronaut |
| `jakarta.servlet` | Jakarta EE / Servlet |

### Spring Boot

- Fat JAR includes embedded Tomcat
- Config: `application.properties`, `application.yml`, `application-*.yml`

## Install Stage Optimization

- Maven: `pom.xml`, `pom.xml` children (multi-module)
- Gradle: `build.gradle`, `build.gradle.kts`, `settings.gradle`, `gradlew`, `gradle/wrapper/*`
- Then: `src/`

## Caching

Use BuildKit cache mounts:

**Gradle:**
```dockerfile
RUN --mount=type=cache,target=/root/.gradle \
    ./gradlew build -x test --no-daemon
```

**Maven:**
```dockerfile
RUN --mount=type=cache,target=/root/.m2/repository \
    mvn -B package -DskipTests
```

## Base Images

| Stage | Image |
|---|---|
| Build (Maven) | `eclipse-temurin:21-jdk-alpine` or `maven:3-eclipse-temurin-21` |
| Build (Gradle) | `eclipse-temurin:21-jdk-alpine` |
| Runtime | `eclipse-temurin:21-jre-alpine` or `amazoncorretto:21-alpine` |

Use JRE (not JDK) for runtime.

## Dockerfile Patterns

### Maven + Spring Boot

```dockerfile
FROM eclipse-temurin:21-jdk-alpine AS build
WORKDIR /app
COPY pom.xml ./
RUN mvn -B dependency:go-offline
COPY src ./src
RUN mvn -B package -DskipTests

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

### Gradle + Spring Boot

```dockerfile
FROM eclipse-temurin:21-jdk-alpine AS build
WORKDIR /app
COPY gradlew ./
COPY gradle gradle
COPY build.gradle settings.gradle ./
RUN ./gradlew dependencies --no-daemon
COPY src ./src
RUN ./gradlew build -x test --no-daemon

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=build /app/build/libs/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

### Simple Maven (no Spring Boot)

```dockerfile
FROM eclipse-temurin:21-jdk-alpine AS build
WORKDIR /app
COPY pom.xml ./
RUN mvn -B dependency:go-offline
COPY src ./src
RUN mvn -B package -DskipTests

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]
```

## Gotchas

- Gradle wrapper `gradlew` must have execute permission — add `RUN chmod +x gradlew` in the Dockerfile if builds fail with permission denied
- `--no-daemon` is essential for Gradle in Docker — the daemon persists between builds, wastes memory, and can hang
- Spring Boot fat JAR glob `target/*.jar` may match multiple JARs (sources, javadoc) — use `maven-jar-plugin` with a fixed `finalName` or be specific with the path
- Use JRE (not JDK) for the runtime stage — JDK adds ~200MB of unnecessary build tools
- `JAVA_TOOL_OPTIONS` affects all stages including runtime — unset it after the build stage if used for build-only JVM flags
