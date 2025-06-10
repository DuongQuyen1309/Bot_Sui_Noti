package router

import (
	"github.com/DuongQuyen1309/suibottele/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/received-amount", handler.GetReceivedAmountOfACoinType)
	router.GET("/sent-amount", handler.GetSentAmountOfACoinType)
	// vì trong 1 transaction block, một ví có thể vừa tăng vừa giảm, có thể chứa nhiều hơn 1 bản ghi, nên em để transactions
	router.GET("/transactions/:hash", handler.DetailTransactionByHash)
	router.GET("/transactions", handler.ListTransactionsInRange)
	return router
}
