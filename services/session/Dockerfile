# Not Change This Version Of Gradle
FROM gradle:9.2.1-jdk21-jammy AS build

WORKDIR /app

# Copy build files
COPY build.gradle.kts settings.gradle.kts ./
COPY gradle ./gradle

# Copy proto files for generation (before source code)
COPY proto ./proto

# Download dependencies
RUN gradle dependencies --no-daemon

# Copy source code
COPY src ./src

# Build application
RUN gradle build --no-daemon -x test

# Runtime stage
FROM eclipse-temurin:21-jre-jammy

WORKDIR /app

# Copy built JAR
COPY --from=build /app/build/libs/*.jar app.jar

# Expose ports (REST and gRPC)
EXPOSE 8080 50056

# Run application
ENTRYPOINT ["java", "-jar", "app.jar"]

