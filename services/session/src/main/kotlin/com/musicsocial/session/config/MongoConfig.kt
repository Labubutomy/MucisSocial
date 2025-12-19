package com.musicsocial.session.config

import com.mongodb.client.MongoClient
import com.mongodb.client.MongoClients
import org.slf4j.LoggerFactory
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.Primary
import org.springframework.data.mongodb.MongoDatabaseFactory
import org.springframework.data.mongodb.core.MongoTemplate
import org.springframework.data.mongodb.core.SimpleMongoClientDatabaseFactory

@Configuration
class MongoConfig {
    
    private val logger = LoggerFactory.getLogger(javaClass)
    
    private fun getMongoUri(): String {
        // Try SPRING_DATA_MONGODB_URI first
        val uriFromEnv = System.getenv("SPRING_DATA_MONGODB_URI")
        if (!uriFromEnv.isNullOrBlank()) {
            logger.info("Using MongoDB URI from SPRING_DATA_MONGODB_URI environment variable")
            return uriFromEnv
        }
        
        // Otherwise construct from components
        val host = System.getenv("MONGODB_HOST") ?: "mongodb"
        val port = System.getenv("MONGODB_PORT") ?: "27017"
        val database = System.getenv("MONGODB_DATABASE") ?: "music_sessions"
        
        val uri = "mongodb://$host:$port/$database"
        logger.info("Constructed MongoDB URI from components: $uri")
        return uri
    }
    
    @Bean
    @Primary
    fun mongoClient(): MongoClient {
        val uri = getMongoUri()
        logger.info("Creating MongoClient with URI: $uri")
        return MongoClients.create(uri)
    }
    
    @Bean
    @Primary
    fun mongoDatabaseFactory(mongoClient: MongoClient): MongoDatabaseFactory {
        val uri = getMongoUri()
        val database = uri.substringAfterLast("/").substringBefore("?")
        logger.info("Creating MongoDatabaseFactory for database: $database")
        return SimpleMongoClientDatabaseFactory(mongoClient, database)
    }
    
    @Bean
    @Primary
    fun mongoTemplate(mongoDatabaseFactory: MongoDatabaseFactory): MongoTemplate {
        return MongoTemplate(mongoDatabaseFactory)
    }
}

