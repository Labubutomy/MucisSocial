package com.musicsocial.session.domain.model

import org.springframework.data.annotation.Id
import org.springframework.data.mongodb.core.index.Indexed
import org.springframework.data.mongodb.core.mapping.Document
import java.time.Instant

@Document(collection = "rooms")
data class RoomState(
    @Id
    val id: String? = null,
    @Indexed(unique = true)
    val roomId: String,
    val currentTrack: TrackInfo?,
    val position: Double, // позиция в секундах
    val isPlaying: Boolean,
    val participants: List<Participant>,
    val queue: List<TrackInfo>,
    val lastAction: Action?,
    val createdAt: Instant,
    val updatedAt: Instant
)

data class TrackInfo(
    val trackId: String,
    val title: String,
    val artist: String,
    val duration: Double,
    val cdnUrl: String
)

data class Participant(
    val userId: String,
    val username: String,
    val isOnline: Boolean,
    val joinedAt: Instant
)

data class Action(
    val actionId: String,
    val type: ActionType,
    val userId: String,
    val timestamp: Instant,
    val payload: Map<String, Any>?
)

enum class ActionType {
    PLAY,
    PAUSE,
    SEEK,
    CHANGE_TRACK,
    JOIN_ROOM,
    LEAVE_ROOM
}

