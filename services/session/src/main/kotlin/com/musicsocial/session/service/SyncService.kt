package com.musicsocial.session.service

import com.musicsocial.session.domain.model.RoomState
import com.musicsocial.session.domain.model.SyncMessage
import org.slf4j.LoggerFactory
import org.springframework.beans.factory.annotation.Value
import org.springframework.kafka.core.KafkaTemplate
import org.springframework.stereotype.Service
import java.time.Instant
import java.util.*

@Service
class SyncService(
    private val kafkaTemplate: KafkaTemplate<String, Any>,
    @Value("\${kafka.topics.sync}")
    private val syncTopic: String
) {
    
    private val logger = LoggerFactory.getLogger(SyncService::class.java)
    
    fun sendSyncMessage(roomState: RoomState, triggeredBy: String) {
        try {
            val syncMessage = SyncMessage(
                syncId = UUID.randomUUID().toString(),
                roomId = roomState.roomId,
                state = roomState,
                triggeredBy = triggeredBy,
                timestamp = Instant.now()
            )
            
            logger.info("[SyncService] ========== SENDING SYNC MESSAGE ==========")
            logger.info("[SyncService] Room: ${roomState.roomId}")
            logger.info("[SyncService] isPlaying: ${roomState.isPlaying}")
            logger.info("[SyncService] currentTrack: ${roomState.currentTrack?.trackId} - ${roomState.currentTrack?.title}")
            logger.info("[SyncService] position: ${roomState.position}")
            logger.info("[SyncService] lastAction: ${roomState.lastAction?.type}")
            logger.info("[SyncService] Topic: $syncTopic")
            logger.info("[SyncService] Triggered by: $triggeredBy")
            
            val future = kafkaTemplate.send(syncTopic, roomState.roomId, syncMessage)
            future.whenComplete { result, exception ->
                if (exception != null) {
                    logger.error("[SyncService] Error sending sync message: ${exception.message}", exception)
                } else {
                    logger.info("[SyncService] Sync message sent successfully to topic $syncTopic, partition ${result?.recordMetadata?.partition()}, offset ${result?.recordMetadata?.offset()}")
                }
            }
        } catch (e: Exception) {
            logger.error("[SyncService] Error sending sync message: ${e.message}", e)
            e.printStackTrace()
        }
    }
}

