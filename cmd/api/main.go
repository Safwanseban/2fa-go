package main

import (
	"log"

	"github.com/Safwanseban/2fa-go/internal/db"
	"github.com/Safwanseban/2fa-go/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	db := db.InitDB()

	handlers.NewApiHandler(engine, db)
	log.Fatal(engine.Run(":3000"))

}
