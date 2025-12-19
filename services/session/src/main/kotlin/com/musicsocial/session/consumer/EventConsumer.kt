package com.musicsocial.session.consumer

import com.musicsocial.session.domain.model.ClientEvent
import com.musicsocial.session.service.EventProcessor
import com.musicsocial.session.service.SyncService
import org.slf4j.LoggerFactory
import org.springframework.kafka.annotation.KafkaListener
import org.springframework.kafka.support.Acknowledgment
import org.springframework.kafka.support.KafkaHeaders
import org.springframework.messaging.handler.annotation.Header
import org.springframework.messaging.handler.annotation.Payload
import org.springframework.stereotype.Component
import java.time.Instant

@Component
class EventConsumer(
    private val eventProcessor: EventProcessor,
    private val syncService: SyncService
) {
    
    private val logger = LoggerFactory.getLogger(EventConsumer::class.java)
    
    init {
        logger.info("[EventConsumer] EventConsumer initialized and ready to consume events")
    }
    
    @KafkaListener(
        topics = ["\${kafka.topics.events}"],
        groupId = "session-service-group"
    )
    fun consumeEvent(
        @Payload event: Map<String, Any>,
        @Header(KafkaHeaders.RECEIVED_KEY) roomId: String?,
        acknowledgment: Acknowledgment
    ) {
        try {
            logger.info("[EventConsumer] ========== RECEIVED EVENT FROM KAFKA ==========")
            logger.info("[EventConsumer] Received event: $event")
            logger.info("[EventConsumer] Event payload: ${event["payload"]}")
            
            val actionStr = (event["action"] as? String ?: "PLAY").uppercase()
            logger.info("[EventConsumer] Parsing action: $actionStr")
            
            val clientEvent = ClientEvent(
                eventId = event["event_id"] as? String ?: "",
                type = com.musicsocial.session.domain.model.EventType.valueOf(
                    (event["type"] as? String ?: "PLAYER_ACTION").uppercase()
                ),
                roomId = event["room_id"] as? String ?: roomId ?: "",
                userId = event["user_id"] as? String ?: "",
                action = com.musicsocial.session.domain.model.PlayerAction.valueOf(
                    actionStr
                ),
                payload = (event["payload"] as? Map<*, *>)?.mapKeys { it.key.toString() }
                    ?.mapValues { it.value } as? Map<String, Any> ?: emptyMap(),
                clientTimestamp = Instant.ofEpochSecond(
                    (event["client_timestamp"] as? Number)?.toLong() ?: 0L
                ),
                serverTimestamp = Instant.ofEpochSecond(
                    (event["server_timestamp"] as? Number)?.toLong() ?: Instant.now().epochSecond
                )
            )
            
            logger.info("[EventConsumer] Parsed ClientEvent: action=${clientEvent.action}, roomId=${clientEvent.roomId}, userId=${clientEvent.userId}")
            logger.info("[EventConsumer] ClientEvent payload: ${clientEvent.payload}")
            
            val updatedRoom = eventProcessor.processEvent(clientEvent)
            
            logger.info("[EventConsumer] Event processed, updatedRoom: ${updatedRoom != null}")
            if (updatedRoom != null) {
                logger.info("[EventConsumer] ========== BEFORE SENDING SYNC ==========")
                logger.info("[EventConsumer] Updated room state:")
                logger.info("[EventConsumer]   - roomId: ${updatedRoom.roomId}")
                logger.info("[EventConsumer]   - isPlaying: ${updatedRoom.isPlaying}")
                logger.info("[EventConsumer]   - currentTrack: ${updatedRoom.currentTrack?.trackId} - ${updatedRoom.currentTrack?.title}")
                logger.info("[EventConsumer]   - position: ${updatedRoom.position}")
                logger.info("[EventConsumer]   - lastAction: ${updatedRoom.lastAction?.type}")
                logger.info("[EventConsumer] Calling syncService.sendSyncMessage...")
                syncService.sendSyncMessage(updatedRoom, clientEvent.userId)
                logger.info("[EventConsumer] syncService.sendSyncMessage called")
            } else {
                logger.warn("[EventConsumer] Event processing returned null room state")
            }
            
            acknowledgment.acknowledge()
        } catch (e: Exception) {
            logger.error("[EventConsumer] Error processing event: ${e.message}", e)
            e.printStackTrace()
            // В продакшене здесь должна быть обработка ошибок и retry логика
            acknowledgment.acknowledge()
        }
    }
}

