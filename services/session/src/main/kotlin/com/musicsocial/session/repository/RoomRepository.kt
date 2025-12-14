package com.musicsocial.session.repository

import com.musicsocial.session.domain.model.RoomState
import org.springframework.data.mongodb.repository.MongoRepository
import org.springframework.stereotype.Repository
import java.util.Optional

@Repository
interface RoomRepository : MongoRepository<RoomState, String> {
    fun findByRoomId(roomId: String): Optional<RoomState>
    fun existsByRoomId(roomId: String): Boolean
}

