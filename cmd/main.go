package main

import (
	"context"
	"fmt"

	"github.com/DuongQuyen1309/suibottele/internal/datastore"
	"github.com/DuongQuyen1309/suibottele/internal/db"

	// "github.com/DuongQuyen1309/suibottele/internal/router"
	"github.com/DuongQuyen1309/suibottele/internal/service"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file", err)
	}
}
func main() {
	ctx := context.Background()
	db.ConnectDB()
	datastore.CreateTransactionsTable(ctx)
	err := service.SUITeleNoti(ctx)
	if err != nil {
		return
	}
	// router := router.SetupRouter()
	// router.Run(":8080")
}
