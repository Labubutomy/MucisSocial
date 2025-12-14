package com.musicsocial.session.controller

import com.musicsocial.session.domain.model.RoomState
import com.musicsocial.session.service.RoomService
import com.musicsocial.session.service.SyncService
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*

@RestController
@RequestMapping("/api/rooms")
class RoomController(
    private val roomService: RoomService,
    private val syncService: SyncService
) {
    
    @GetMapping("/health")
    fun health(): ResponseEntity<Map<String, String>> {
        return ResponseEntity.ok(mapOf("status" to "healthy"))
    }
    
    @GetMapping("/{roomId}")
    fun getRoom(@PathVariable roomId: String): ResponseEntity<RoomState> {
        val room = roomService.getRoom(roomId)
        return if (room != null) {
            ResponseEntity.ok(room)
        } else {
            ResponseEntity.notFound().build()
        }
    }
    
    @PostMapping("/{roomId}")
    fun createRoom(@PathVariable roomId: String): ResponseEntity<RoomState> {
        val room = roomService.createRoom(roomId)
        return ResponseEntity.ok(room)
    }
    
    @PostMapping("/{roomId}/participants")
    fun addParticipant(
        @PathVariable roomId: String,
        @RequestParam userId: String,
        @RequestParam username: String
    ): ResponseEntity<RoomState> {
        val room = roomService.addParticipant(roomId, userId, username)
        // Send sync message to notify all participants about the new participant
        syncService.sendSyncMessage(room, userId)
        return ResponseEntity.ok(room)
    }
    
    @DeleteMapping("/{roomId}/participants/{userId}")
    fun removeParticipant(
        @PathVariable roomId: String,
        @PathVariable userId: String
    ): ResponseEntity<RoomState> {
        val room = roomService.removeParticipant(roomId, userId)
        return if (room != null) {
            // Send sync message to notify all participants about the participant leaving
            syncService.sendSyncMessage(room, userId)
            ResponseEntity.ok(room)
        } else {
            ResponseEntity.notFound().build()
        }
    }
    
    @DeleteMapping("/{roomId}")
    fun deleteRoom(@PathVariable roomId: String): ResponseEntity<Void> {
        roomService.deleteRoom(roomId)
        return ResponseEntity.noContent().build()
    }
}

