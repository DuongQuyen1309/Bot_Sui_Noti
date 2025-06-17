package router

import (
	"github.com/DuongQuyen1309/suibottele/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/received-amount", handler.GetReceivedAmountOfACoinType)
	router.GET("/sent-amount", handler.GetSentAmountOfACoinType)
	router.GET("/transaction/:hash/balance-change-events", handler.GetBalanceChangeEventsByTransactionHash)
	router.GET("/balance-change-events", handler.ListEventsInRange)
	return router
}
