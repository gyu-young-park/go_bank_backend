package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/gyu-young-park/simplebank/db/sqlc"
	"github.com/gyu-young-park/simplebank/token"
	"github.com/gyu-young-park/simplebank/util"
)

//Server serves HTTP requests for out banking service.
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.TokenMaker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	toekenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		store:      store,
		tokenMaker: toekenMaker,
		config:     config,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	//HandleFunc을 여러개를 넣을 수 있는데 마지막이 진짜 핸들러고 중간은 미들웨어이다.
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	router.POST("/transfers", server.createTransfer)
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	server.router = router
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
