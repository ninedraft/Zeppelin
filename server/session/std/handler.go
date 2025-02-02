package std

import (
	"time"

	"github.com/zeppelinmc/zeppelin/log"
	"github.com/zeppelinmc/zeppelin/net"
	"github.com/zeppelinmc/zeppelin/net/io"
	"github.com/zeppelinmc/zeppelin/net/packet"
	"github.com/zeppelinmc/zeppelin/net/packet/configuration"
	"github.com/zeppelinmc/zeppelin/net/packet/play"
	"github.com/zeppelinmc/zeppelin/text"
)

type handler func(*StandardSession, packet.Packet)

var handlers = make(map[[2]int32]handler)

func RegisterHandler(state, id int32, handler handler) {
	handlers[[2]int32{state, id}] = handler
}

func (session *StandardSession) inConfiguration() bool {
	return session.conn.State() == net.ConfigurationState
}

func (session *StandardSession) handlePackets() {
	keepAlive := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-keepAlive.C:
			session.conn.WritePacket(&play.ClientboundKeepAlive{KeepAliveID: time.Now().UnixMilli()})
		default:
			if lastKeepAlive := session.lastKeepalive.Get(); lastKeepAlive != 0 && time.Now().Unix()-lastKeepAlive > 21 {
				session.Disconnect(text.TextComponent{Text: "Timed out"})
			}
			p, err := session.conn.ReadPacket()
			if err != nil {
				session.broadcast.RemovePlayer(session)
				return
			}

			handler, ok := handlers[[2]int32{session.conn.State(), p.ID()}]
			if !ok {
				switch pk := p.(type) {
				case *play.ChunkBatchReceived:
					session.awaitingChunkBatchAcknowledgement.Set(false)
				case *play.ConfirmTeleportation:
					session.AwaitingTeleportAcknowledgement.Set(false)
				case *play.ServerboundKeepAlive:
					session.lastKeepalive.Set(time.Now().Unix())
				case *play.SetHeldItemServerbound:
					if pk.Slot < 0 || pk.Slot > 8 {
						session.Disconnect(text.TextComponent{Text: "Invalid slot"})
					}
					session.player.SetSelectedItemSlot(int32(pk.Slot))
				case *play.PlayerSession:
					session.hasSessionData.Set(true)
					session.sessionData.Set(*pk)

					session.broadcast.UpdateSession(session)
				case *configuration.ServerboundPluginMessage:
					if pk.Channel == "minecraft:brand" {
						_, data, _ := io.ReadVarInt(pk.Data)
						session.clientName = string(data)
					}
				case *configuration.AcknowledgeFinishConfiguration:
					session.conn.SetState(net.PlayState)
					session.login()
				default:
					log.Printlnf("Unknown packet 0x%02x %T", p.ID(), p)
				}
				continue
			}
			handler(session, p)
		}
	}
}
