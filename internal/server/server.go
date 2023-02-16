package server

import (
	"flag"
	"os"
	"time"
	"videochat/internal/handlers" //----coustom packages

	w "videochat/pkg/webrtc"

	"github.com/gofiber/fiber/v2" //----- web routing framework.
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html" //------for frontend part.
	"github.com/gofiber/websocket/v2"
)

var (
	addr = flag.String("addr ", ":"+os.Getenv("PORT"), "")
	cert = flag.String("cert", "", "")
	key  = flag.String("key", "", "")
)

func Run() error {
	flag.Parse()

	if *addr == ":" {
		*addr = ":8080"
	}

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})
	app.Use(logger.New()) //------middlewares
	app.Use(cors.New())   //------middlewares

	//--------------setting the routes for http request and response .
	app.Get("/", handlers.Welcome)
	app.Get("/room/:uuid", handlers.Room)
	app.Get("/room/create", handlers.RoomCreate)
	app.Get("/room/:uuid/websocket", websocket.New(handlers.RoomWebsocket, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	}))
	app.Get("/room/:uuid/chat", handlers.RoomChat)
	app.Get("/room/:uuid/chat/websocket", websocket.New(handlers.RoomChatWebsocket))
	app.Get("/room/:uuid/viewer/websocket", websocket.New(handlers.RoomViewerWebsocket))

	//--------for stream websocket .
	app.Get("/stream/:ssuid", handlers.Stream)
	app.Get("/stream/:ssuid/websocket", websocket.New(handlers.StreamWebsocket, websocket.Config{
		HandshakeTimeout: 10 * time.Second}))
	app.Get("/stream/:ssuid/chat/websocket", websocket.New(handlers.StreamChatWebsocket))
	app.Get("/stream/:ssuid/viewer/websocket", websocket.New(handlers.StreamViewerWebsocket))

	app.Static("/", "./assets")

	w.Rooms = make(map[string]*w.Room) //------here 'w' stand for webrtc
	w.Stream = make(map[string]*w.Room)

	go dispatchKeyFrames()
	if *cert != "" {
		return app.ListenTLS(*addr, *cert, *key)
	}
	return app.Listen(*addr)

}

func dispatchKeyFrames() {
	for range time.NewTicker(time.Second * 3).c {
		for _, room := range w.Rooms {
			room.Peers.DispatchKeyFrame()
		}
	}
}
