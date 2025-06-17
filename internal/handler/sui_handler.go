package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DuongQuyen1309/suibottele/internal/datastore"
	"github.com/gin-gonic/gin"
)

const (
	DATE_PATTERN = "2006-01-02"
)

func GetReceivedAmountOfACoinType(c *gin.Context) {
	coinType := c.Query("coin-type")
	if coinType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coin-type is required"})
		return
	}
	totalAmount, err := datastore.CalculaterReceivedAmount(coinType, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"totalAmount": totalAmount})
}
func GetSentAmountOfACoinType(c *gin.Context) {
	coinType := c.Query("coin-type")
	if coinType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coin-type is required"})
		return
	}
	totalAmount, err := datastore.CalculaterSentAmount(coinType, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"totalAmount": totalAmount})
}

func GetBalanceChangeEventsByTransactionHash(c *gin.Context) {
	hash := c.Param("hash")
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page is required and must be a number"})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit is required and must be a number"})
		return
	}
	if page <= 0 || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page and limit must be greater than 0"})
		return
	}
	offset := (page - 1) * limit
	transaction, err := datastore.GetBalanceChangeEvents(hash, offset, limit, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func ListEventsInRange(c *gin.Context) {
	fromDateInput := c.Query("from-date")
	toDateInput := c.Query("do-date")
	if fromDateInput == "" || toDateInput == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from-date and to-date is required"})
		return
	}
	fromDate, err := time.Parse(DATE_PATTERN, fromDateInput)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from-date format is invalid"})
		return
	}
	toDate, err := time.Parse(DATE_PATTERN, toDateInput)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from-date format is invalid"})
		return
	}
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page is required and must be a number"})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit is required and must be a number"})
		return
	}
	if page <= 0 || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page and limit must be greater than 0"})
		return
	}
	offset := (page - 1) * limit
	transactions, err := datastore.GetBalanceChangeEventInRange(fromDate, toDate, offset, limit, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, transactions)
}
