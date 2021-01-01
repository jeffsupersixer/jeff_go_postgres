package api

import (
	db "github.com/jeffsupersixer/jeff_go_postgres/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves http request
type Server struct {
	store *db.Store
	router *gin.Engine
}

// NewServer create a new Http server
func NewServer(store *db.Store) *Server {
	server := &Server{
		store: store,
	}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)

	server.router = router
	return server
}
