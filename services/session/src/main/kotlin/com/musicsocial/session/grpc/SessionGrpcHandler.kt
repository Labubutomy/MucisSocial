package com.musicsocial.session.grpc

import com.musicsocial.session.domain.model.RoomState as DomainRoomState
import com.musicsocial.session.domain.model.TrackInfo as DomainTrackInfo
import com.musicsocial.session.domain.model.Participant as DomainParticipant
import com.musicsocial.session.domain.model.Action as DomainAction
import com.musicsocial.session.domain.model.ActionType as DomainActionType
import com.musicsocial.session.service.RoomService
import com.musicsocial.session.service.SyncService
import com.musicsocial.session.proto.v1.SessionServiceGrpc
import com.musicsocial.session.proto.v1.SessionProto.*
import com.google.protobuf.Timestamp
import io.grpc.stub.StreamObserver
import net.devh.boot.grpc.server.service.GrpcService
import org.springframework.beans.factory.annotation.Autowired
import java.time.Instant

@GrpcService
class SessionGrpcHandler @Autowired constructor(
    private val roomService: RoomService,
    private val syncService: SyncService
) : SessionServiceGrpc.SessionServiceImplBase() {
    
    override fun createRoom(
        request: CreateRoomRequest,
        responseObserver: StreamObserver<CreateRoomResponse>
    ) {
        try {
            val room = roomService.createRoom(request.roomId)
            val response = CreateRoomResponse.newBuilder()
                .setRoom(convertToProto(room))
                .build()
            responseObserver.onNext(response)
            responseObserver.onCompleted()
        } catch (e: Exception) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                .withDescription("Failed to create room: ${e.message}")
                .asRuntimeException())
        }
    }
    
    override fun getRoom(
        request: GetRoomRequest,
        responseObserver: StreamObserver<GetRoomResponse>
    ) {
        try {
            val room = roomService.getRoom(request.roomId)
            if (room == null) {
                responseObserver.onError(io.grpc.Status.NOT_FOUND
                    .withDescription("Room not found: ${request.roomId}")
                    .asRuntimeException())
                return
            }
            val response = GetRoomResponse.newBuilder()
                .setRoom(convertToProto(room))
                .build()
            responseObserver.onNext(response)
            responseObserver.onCompleted()
        } catch (e: Exception) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                .withDescription("Failed to get room: ${e.message}")
                .asRuntimeException())
        }
    }
    
    override fun deleteRoom(
        request: DeleteRoomRequest,
        responseObserver: StreamObserver<DeleteRoomResponse>
    ) {
        try {
            roomService.deleteRoom(request.roomId)
            val response = DeleteRoomResponse.newBuilder()
                .setSuccess(true)
                .build()
            responseObserver.onNext(response)
            responseObserver.onCompleted()
        } catch (e: Exception) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                .withDescription("Failed to delete room: ${e.message}")
                .asRuntimeException())
        }
    }
    
    override fun addParticipant(
        request: AddParticipantRequest,
        responseObserver: StreamObserver<AddParticipantResponse>
    ) {
        try {
            val room = roomService.addParticipant(
                request.roomId,
                request.userId,
                request.username
            )
            // Send sync message to notify all participants about the new participant
            syncService.sendSyncMessage(room, request.userId)
            
            val response = AddParticipantResponse.newBuilder()
                .setRoom(convertToProto(room))
                .build()
            responseObserver.onNext(response)
            responseObserver.onCompleted()
        } catch (e: Exception) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                .withDescription("Failed to add participant: ${e.message}")
                .asRuntimeException())
        }
    }
    
    override fun removeParticipant(
        request: RemoveParticipantRequest,
        responseObserver: StreamObserver<RemoveParticipantResponse>
    ) {
        try {
            val room = roomService.removeParticipant(request.roomId, request.userId)
            if (room == null) {
                responseObserver.onError(io.grpc.Status.NOT_FOUND
                    .withDescription("Room or participant not found")
                    .asRuntimeException())
                return
            }
            // Send sync message to notify all participants about the participant leaving
            syncService.sendSyncMessage(room, request.userId)
            
            val response = RemoveParticipantResponse.newBuilder()
                .setRoom(convertToProto(room))
                .build()
            responseObserver.onNext(response)
            responseObserver.onCompleted()
        } catch (e: Exception) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                .withDescription("Failed to remove participant: ${e.message}")
                .asRuntimeException())
        }
    }
    
    override fun healthCheck(
        request: HealthCheckRequest,
        responseObserver: StreamObserver<HealthCheckResponse>
    ) {
        val response = HealthCheckResponse.newBuilder()
            .setStatus("healthy")
            .build()
        responseObserver.onNext(response)
        responseObserver.onCompleted()
    }
    
    private fun convertToProto(room: DomainRoomState): RoomState {
        val builder = RoomState.newBuilder()
            .setRoomId(room.roomId)
            .setPosition(room.position)
            .setIsPlaying(room.isPlaying)
            .setCreatedAt(toTimestamp(room.createdAt))
            .setUpdatedAt(toTimestamp(room.updatedAt))
        
        room.currentTrack?.let {
            builder.setCurrentTrack(convertTrackToProto(it))
        }
        
        room.participants.forEach { participant ->
            builder.addParticipants(convertParticipantToProto(participant))
        }
        
        room.queue.forEach { track ->
            builder.addQueue(convertTrackToProto(track))
        }
        
        room.lastAction?.let {
            builder.setLastAction(convertActionToProto(it))
        }
        
        return builder.build()
    }
    
    private fun convertTrackToProto(track: DomainTrackInfo): TrackInfo {
        return TrackInfo.newBuilder()
            .setTrackId(track.trackId)
            .setTitle(track.title)
            .setArtist(track.artist)
            .setDuration(track.duration)
            .setCdnUrl(track.cdnUrl)
            .build()
    }
    
    private fun convertParticipantToProto(participant: DomainParticipant): Participant {
        return Participant.newBuilder()
            .setUserId(participant.userId)
            .setUsername(participant.username)
            .setIsOnline(participant.isOnline)
            .setJoinedAt(toTimestamp(participant.joinedAt))
            .build()
    }
    
    private fun convertActionToProto(action: DomainAction): Action {
        val builder = Action.newBuilder()
            .setActionId(action.actionId)
            .setType(convertActionTypeToProto(action.type))
            .setUserId(action.userId)
            .setTimestamp(toTimestamp(action.timestamp))
        
        action.payload?.forEach { (key, value) ->
            builder.putPayload(key, value.toString())
        }
        
        return builder.build()
    }
    
    private fun convertActionTypeToProto(type: DomainActionType): ActionType {
        return when (type) {
            DomainActionType.PLAY -> ActionType.ACTION_TYPE_PLAY
            DomainActionType.PAUSE -> ActionType.ACTION_TYPE_PAUSE
            DomainActionType.SEEK -> ActionType.ACTION_TYPE_SEEK
            DomainActionType.CHANGE_TRACK -> ActionType.ACTION_TYPE_CHANGE_TRACK
            DomainActionType.JOIN_ROOM -> ActionType.ACTION_TYPE_JOIN_ROOM
            DomainActionType.LEAVE_ROOM -> ActionType.ACTION_TYPE_LEAVE_ROOM
        }
    }
    
    private fun toTimestamp(instant: Instant): Timestamp {
        return Timestamp.newBuilder()
            .setSeconds(instant.epochSecond)
            .setNanos(instant.nano)
            .build()
    }
}
