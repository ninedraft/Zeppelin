package server

import (
	"fmt"
	"runtime"
	"time"

	"github.com/zeppelinmc/zeppelin/net/packet/configuration"
	"github.com/zeppelinmc/zeppelin/net/packet/status"
	"github.com/zeppelinmc/zeppelin/server/command"
	"github.com/zeppelinmc/zeppelin/server/config"
	"github.com/zeppelinmc/zeppelin/server/player"
	"github.com/zeppelinmc/zeppelin/server/session/std"
	_ "github.com/zeppelinmc/zeppelin/server/session/std/handler"
	"github.com/zeppelinmc/zeppelin/text"
	"github.com/zeppelinmc/zeppelin/util"

	"github.com/zeppelinmc/zeppelin/server/world"
	_ "github.com/zeppelinmc/zeppelin/server/world/terrain"

	"github.com/zeppelinmc/zeppelin/log"
	"github.com/zeppelinmc/zeppelin/net"
	"github.com/zeppelinmc/zeppelin/server/session"
)

// Creates a new server instance using the specified config, returns an error if unable to bind listener
func New(cfg config.ServerConfig, world *world.World) (*Server, error) {
	var statusProvider = net.Status(status.StatusResponseData{
		Version: status.StatusVersion{
			Name:     "1.21",
			Protocol: net.ProtocolVersion,
		},
		Description: text.Unmarshal(cfg.MOTD, cfg.Chat.Formatter.Rune()),
		Players: status.StatusPlayers{
			Max: 20,
		},
		EnforcesSecureChat: true,
	})
	lcfg := net.Config{
		Status: statusProvider,

		IP:                   cfg.Net.ServerIP,
		Port:                 cfg.Net.ServerPort,
		CompressionThreshold: cfg.Net.CompressionThreshold,
		Encrypt:              cfg.Net.EncryptionMode == config.EncryptionYes || cfg.Net.EncryptionMode == config.EncryptionOnline,
		Authenticate:         cfg.Net.EncryptionMode == config.EncryptionOnline,
	}
	if cfg.Net.EncryptionMode != config.EncryptionOnline {
		log.Warnln("Server is running in offline mode. The server will let anyone log as any username and potentially harm the server. Proceed with caution")
	}

	if cfg.Chat.ChatMode == "secure" && cfg.Net.EncryptionMode != config.EncryptionOnline {
		log.Warnln("You can't use secure chat without encryption mode set to online! Using disguised chat mode instead.")
		cfg.Chat.ChatMode = "disguised"
	}
	if cfg.Chat.ChatMode == "secure" && cfg.Chat.DisableWarning {
		log.Warnln("Enabling Chat.DisableWarning is redundant when using secure chat mode.")
		cfg.Chat.DisableWarning = false
	}
	if cfg.Chat.DisableWarning {
		log.Warnln("Using Chat.DisableWarning violates the MUG (Minecraft Usage Guidelines) and could get your server banned. Proceed with caution")
	}

	listener, err := lcfg.New()
	server := &Server{
		listener: listener,
		cfg:      cfg,
		World:    world,
		Players:  player.NewPlayerManager(),
		stopLoop: make(chan int),
	}
	server.Console = &Console{Server: server}
	server.Broadcast = session.NewBroadcast(server.Console)

	compstr := "compress everything"
	if cfg.Net.CompressionThreshold > 0 {
		compstr = fmt.Sprintf("compress everything over %d bytes", cfg.Net.CompressionThreshold)
	} else if cfg.Net.CompressionThreshold < 0 {
		compstr = "no compression"
	}

	log.Infof("Compression threshold is %d (%s)\n", cfg.Net.CompressionThreshold, compstr)
	server.createTicker()
	return server, err
}

type Server struct {
	cfg      config.ServerConfig
	listener *net.Listener
	ticker   Ticker

	World *world.World

	Console *Console

	Broadcast *session.Broadcast

	CommandManager *command.Manager

	closed   bool
	stopLoop chan int

	Players *player.PlayerManager
}

func (srv *Server) SetStatusProvider(sp net.StatusProvider) {
	srv.listener.SetStatusProvider(sp)
}

func (srv *Server) Config() config.ServerConfig {
	return srv.cfg
}

func (srv *Server) Start(ts time.Time) {
	if !util.HasArgument("--no-plugins") {
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" || runtime.GOOS == "freebsd" {
			log.Infoln("Loading plugins")
			srv.loadPlugins()
		}
	}
	srv.ticker.Start()
	log.Infolnf("Started server ticker (%d TPS)", srv.cfg.Net.TPS)
	log.Infolnf("Done! (%s)", time.Since(ts))
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			if !srv.closed {
				log.Errorln("Server error: ", err)
			}
			<-srv.stopLoop
			return
		}
		srv.handleNewConnection(conn)
	}
}

func (srv *Server) handleNewConnection(conn *net.Conn) {
	log.Infolnf("[%s] Player attempting to connect: %s (%s)", conn.RemoteAddr(), conn.Username(), conn.UUID())
	if _, ok := srv.Broadcast.SessionByUsername(conn.Username()); ok {
		conn.WritePacket(&configuration.Disconnect{
			Reason: text.TextComponent{Text: "You are already connected to the server from another session. Please disconnect then try again"},
		})
		return
	}
	playerData, err := srv.World.PlayerData(conn.UUID().String())
	if err != nil {
		playerData = srv.World.NewPlayerData(conn.UUID())
	}

	player := srv.Players.New(playerData)
	std.NewStandardSession(conn, player, srv.World, srv.Broadcast, srv.cfg, func() net.StatusProvider {
		return srv.listener.StatusProvider()
	}, srv.CommandManager).Configure()
}

func (srv *Server) Stop() {
	log.InfolnClean("Stopping server")
	srv.closed = true
	srv.listener.Close()
	srv.Broadcast.DisconnectAll(text.TextComponent{Text: "Server closed"})

	log.InfolnClean("Saving player data")
	srv.Players.SaveAll()
	srv.stopLoop <- 0
}
