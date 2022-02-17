package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/gyu-young-park/simplebank/db/sqlc"
)

//Server serves HTTP requests for out banking service.
type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	//HandleFunc을 여러개를 넣을 수 있는데 마지막이 진짜 핸들러고 중간은 미들웨어이다.
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

//start runs the http server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// gin.H는 map[string]interface의 shortcut이다.
func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
