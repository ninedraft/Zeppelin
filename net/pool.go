package net

import (
	"github.com/zeppelinmc/zeppelin/net/packet"
	"github.com/zeppelinmc/zeppelin/net/packet/configuration"
	"github.com/zeppelinmc/zeppelin/net/packet/handshake"
	"github.com/zeppelinmc/zeppelin/net/packet/login"
	"github.com/zeppelinmc/zeppelin/net/packet/play"
	"github.com/zeppelinmc/zeppelin/net/packet/status"
)

var serverboundPool = map[int32]map[int32]func() packet.Packet{
	HandshakingState: {
		0x00: func() packet.Packet { return &handshake.Handshaking{} },
	},
	StatusState: {
		0x00: func() packet.Packet { return &status.StatusRequest{} },
		0x01: func() packet.Packet { return &status.Ping{} },
	},
	LoginState: {
		0x00: func() packet.Packet { return &login.LoginStart{} },
		0x02: func() packet.Packet { return &login.LoginPluginResponse{} },
		0x01: func() packet.Packet { return &login.EncryptionResponse{} },
		0x03: func() packet.Packet { return &login.LoginAcknowledged{} },
		0x04: func() packet.Packet { return &login.CookieResponse{} },
	},
	ConfigurationState: {
		0x00: func() packet.Packet { return &configuration.ClientInformation{} },
		0x01: func() packet.Packet { return &configuration.CookieResponse{} },
		0x02: func() packet.Packet { return &configuration.ServerboundPluginMessage{} },
		0x03: func() packet.Packet { return &configuration.AcknowledgeFinishConfiguration{} },
		0x04: func() packet.Packet { return &configuration.KeepAlive{} },
		0x05: func() packet.Packet { return &configuration.Pong{} },
	},
	PlayState: {
		0x00: func() packet.Packet { return &play.ConfirmTeleportation{} },
		0x04: func() packet.Packet { return &play.ChatCommand{} },
		0x05: func() packet.Packet { return &play.SignedChatCommand{} },
		0x06: func() packet.Packet { return &play.ChatMessage{} },
		0x08: func() packet.Packet { return &play.ChunkBatchReceived{} },
		0x07: func() packet.Packet { return &play.PlayerSession{} },
		0x0A: func() packet.Packet { return &play.ClientInformation{} },
		0x0F: func() packet.Packet { return &play.CloseContainer{} },
		0x12: func() packet.Packet { return &play.ServerboundPluginMessage{} },
		0x18: func() packet.Packet { return &play.ServerboundKeepAlive{} },
		0x1A: func() packet.Packet { return &play.SetPlayerPosition{} },
		0x1B: func() packet.Packet { return &play.SetPlayerPositionAndRotation{} },
		0x1C: func() packet.Packet { return &play.SetPlayerRotation{} },
		0x1D: func() packet.Packet { return &play.SetPlayerOnGround{} },
		0x23: func() packet.Packet { return &play.PlayerAbilitiesServerbound{} },
		0x25: func() packet.Packet { return &play.PlayerCommand{} },
		0x2F: func() packet.Packet { return &play.SetHeldItemServerbound{} },
		0x32: func() packet.Packet { return &play.SetCreativeModeSlot{} },
		0x36: func() packet.Packet { return &play.SwingArm{} },
		0x38: func() packet.Packet { return &play.UseItemOn{} },
	},
}

func OverrideSBPool(state, packetId int32, newPacket func() packet.Packet) {
	serverboundPool[state][packetId] = newPacket
}
