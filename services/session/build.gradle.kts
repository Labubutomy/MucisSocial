plugins {
	kotlin("jvm") version "2.2.21"
	kotlin("plugin.spring") version "2.2.21"
	id("org.springframework.boot") version "4.0.0"
	id("io.spring.dependency-management") version "1.1.7"
	id("com.google.protobuf") version "0.9.4"
}

sourceSets {
	main {
		proto {
			// Use only proto/ directory, exclude default src/main/proto
			setSrcDirs(listOf("proto"))
		}
		java {
			srcDirs(
				"src/main/kotlin",
				"build/generated/source/proto/main/java",
				"build/generated/source/proto/main/grpc"
			)
		}
	}
}

protobuf {
	protoc {
		artifact = "com.google.protobuf:protoc:3.25.3"
	}
	plugins {
		create("grpc") {
			artifact = "io.grpc:protoc-gen-grpc-java:1.62.2"
		}
	}
	generateProtoTasks {
		ofSourceSet("main").forEach { task ->
			task.plugins {
				create("grpc")
			}
		}
	}
}


group = "com.musicsocial"
version = "0.0.1-SNAPSHOT"
description = "Demo project for Spring Boot"

java {
	toolchain {
		languageVersion = JavaLanguageVersion.of(21)
	}
}

repositories {
	mavenCentral()
}

dependencies {
	implementation("org.springframework.boot:spring-boot-starter-web")
	implementation("org.springframework.boot:spring-boot-starter-data-redis")
	implementation("org.springframework.kafka:spring-kafka")
	implementation("org.springframework.boot:spring-boot-starter-data-mongodb")
	implementation("org.springframework.boot:spring-boot-starter-validation")
	implementation("com.fasterxml.jackson.module:jackson-module-kotlin")
	implementation("com.fasterxml.jackson.datatype:jackson-datatype-jsr310")
	implementation("org.jetbrains.kotlin:kotlin-reflect")
	implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
	
	// gRPC dependencies
	implementation("net.devh:grpc-spring-boot-starter:2.15.0.RELEASE")
	implementation("net.devh:grpc-client-spring-boot-starter:2.15.0.RELEASE")
	
	// Protobuf and gRPC
	implementation("com.google.protobuf:protobuf-java:3.25.3")
	implementation("com.google.protobuf:protobuf-java-util:3.25.3")
	implementation("io.grpc:grpc-stub:1.62.2")
	implementation("io.grpc:grpc-protobuf:1.62.2")
	implementation("io.grpc:grpc-netty-shaded:1.62.2")
	
	// Required for generated gRPC code (javax.annotation.Generated)
	compileOnly("javax.annotation:javax.annotation-api:1.3.2")
	
	testImplementation("org.springframework.boot:spring-boot-starter-test")
	testImplementation("org.springframework.kafka:spring-kafka-test")
}

kotlin {
	compilerOptions {
		freeCompilerArgs.addAll("-Xjsr305=strict", "-Xannotation-default-target=param-property")
	}
}

tasks.withType<Test> {
	useJUnitPlatform()
}
