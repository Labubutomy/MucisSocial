package com.musicsocial.session.service

import com.musicsocial.session.domain.model.*
import org.springframework.stereotype.Service
import java.time.Instant
import java.util.*

@Service
class EventProcessor(
    private val roomService: RoomService
) {
    
    fun processEvent(event: ClientEvent): RoomState? {
        val room = roomService.getRoom(event.roomId) ?: return null
        
        val action = Action(
            actionId = UUID.randomUUID().toString(),
            type = when (event.action) {
                PlayerAction.PLAY -> ActionType.PLAY
                PlayerAction.PAUSE -> ActionType.PAUSE
                PlayerAction.SEEK -> ActionType.SEEK
                PlayerAction.CHANGE_TRACK -> ActionType.CHANGE_TRACK
            },
            userId = event.userId,
            timestamp = event.serverTimestamp,
            payload = event.payload
        )
        
        val updatedRoom = when (event.action) {
            PlayerAction.PLAY -> {
                println("[EventProcessor] Processing PLAY action - setting isPlaying = true")
                room.copy(
                    id = room.id, // Preserve the MongoDB id
                    isPlaying = true,
                    lastAction = action,
                    updatedAt = Instant.now()
                )
            }
            PlayerAction.PAUSE -> {
                // Update position from payload if provided, otherwise keep current position
                val position = (event.payload["position"] as? Number)?.toDouble() ?: room.position
                room.copy(
                    id = room.id, // Preserve the MongoDB id
                    isPlaying = false,
                    position = position.coerceAtLeast(0.0),
                    lastAction = action,
                    updatedAt = Instant.now()
                )
            }
            PlayerAction.SEEK -> {
                val position = (event.payload["position"] as? Number)?.toDouble() ?: room.position
                room.copy(
                    id = room.id, // Preserve the MongoDB id
                    position = position.coerceAtLeast(0.0),
                    lastAction = action,
                    updatedAt = Instant.now()
                )
            }
            PlayerAction.CHANGE_TRACK -> {
                val trackId = event.payload["track_id"] as? String
                println("[EventProcessor] Processing CHANGE_TRACK, trackId: $trackId")
                println("[EventProcessor] Payload: ${event.payload}")
                
                if (trackId != null) {
                    // В реальном приложении здесь должна быть валидация трека через CDN API
                    val trackInfo = TrackInfo(
                        trackId = trackId,
                        title = event.payload["title"] as? String ?: "Unknown",
                        artist = event.payload["artist"] as? String ?: "Unknown",
                        duration = (event.payload["duration"] as? Number)?.toDouble() ?: 0.0,
                        cdnUrl = event.payload["cdn_url"] as? String ?: ""
                    )
                    println("[EventProcessor] Created TrackInfo: $trackInfo")
                    room.copy(
                        id = room.id, // Preserve the MongoDB id
                        currentTrack = trackInfo,
                        position = 0.0,
                        isPlaying = true,
                        lastAction = action,
                        updatedAt = Instant.now()
                    )
                } else {
                    println("[EventProcessor] WARNING: track_id is null in payload!")
                    room.copy(
                        id = room.id, // Preserve the MongoDB id
                        lastAction = action,
                        updatedAt = Instant.now()
                    )
                }
            }
        }
        
        // Update room state, preserving the id
        return roomService.updateRoomState(event.roomId) { _ -> updatedRoom }
    }
}

