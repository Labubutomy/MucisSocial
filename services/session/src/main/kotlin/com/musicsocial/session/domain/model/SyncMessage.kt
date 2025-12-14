package com.musicsocial.session.domain.model

import java.time.Instant

data class SyncMessage(
    val syncId: String,
    val roomId: String,
    val state: RoomState,
    val triggeredBy: String,
    val timestamp: Instant
)

