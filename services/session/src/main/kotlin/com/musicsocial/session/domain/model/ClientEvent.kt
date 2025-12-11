package com.musicsocial.session.domain.model

import java.time.Instant

data class ClientEvent(
    val eventId: String,
    val type: EventType,
    val roomId: String,
    val userId: String,
    val action: PlayerAction,
    val payload: Map<String, Any>,
    val clientTimestamp: Instant,
    val serverTimestamp: Instant
)

enum class EventType {
    PLAYER_ACTION
}

enum class PlayerAction {
    PLAY,
    PAUSE,
    SEEK,
    CHANGE_TRACK
}

