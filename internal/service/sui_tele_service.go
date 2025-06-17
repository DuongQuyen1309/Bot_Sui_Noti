package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/DuongQuyen1309/suibottele/internal/config"
	"github.com/DuongQuyen1309/suibottele/internal/datastore"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	coinDecimals  = make(map[string]int)
	bot           *tgbotapi.BotAPI
	configuration *config.Config
)

const (
	limitOfItemPerPage = 50
)

func SUITeleNoti(ctx context.Context) error {
	var client = sui.NewSuiClient(constant.SuiMainnetEndpoint)
	var err error
	configuration, err = config.LoadCofig()
	if err != nil {
		fmt.Println("error load config", err)
		return err
	}
	bot, err = CreateBot()
	if err != nil {
		fmt.Println("error create bot", err)
		return err
	}

	for _, token := range configuration.Wallet.Token {
		req := models.SuiXGetCoinMetadataRequest{
			CoinType: token.Address,
		}
		coinMetaData, err := client.SuiXGetCoinMetadata(ctx, req)
		if err != nil {
			fmt.Println("error getting coin metadata", err)
			return err
		}
		coinDecimals[token.Address] = coinMetaData.Decimals
	}

	latestCheckPointNumber, err := client.SuiGetLatestCheckpointSequenceNumber(ctx)
	if err != nil {
		fmt.Println("error getting latest checkpoint", err)
		return err
	}
	newestcheckpoint := latestCheckPointNumber
	generalContext, cancel := context.WithCancel(ctx)
	errChan := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err = FilterInRealtime(client, generalContext, int(newestcheckpoint))
		if err != nil {
			errChan <- err
			fmt.Println("error filtering realtime", err)
			return
		}
	}()
	go func() {
		defer wg.Done()
		pastTime, cancel := context.WithCancel(generalContext)
		err = FilterInPast(pastTime, cancel, client)
		if err != nil {
			errChan <- err
			fmt.Println("error filtering in past", err)
			return
		}
	}()
	select {
	case <-generalContext.Done():
		cancel()
	case err := <-errChan:
		fmt.Println("error in tracking event", err)
		cancel()
		return err
	}
	wg.Wait()
	return nil
}

func FilterInPast(pastTime context.Context, cancel context.CancelFunc, client sui.ISuiAPI) error {
	errChan := make(chan error, 2)
	var wg sync.WaitGroup
	completeChan := make(chan bool, 2)
	complete := 0
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := FilterEventReceivedMoneyInPast(pastTime, client)
		if err != nil {
			errChan <- err
			fmt.Println("error filtering received transactions in past", err)
			return
		}
		completeChan <- true
	}()
	go func() {
		defer wg.Done()
		err := FilterEventSentMoneyInPast(pastTime, client)
		if err != nil {
			errChan <- err
			fmt.Println("error filtering sent transactions in past", err)
			return
		}
		completeChan <- true
	}()
	for {
		select {
		case <-pastTime.Done():
			cancel()
			return pastTime.Err()
		case err := <-errChan:
			cancel()
			return err
		case <-completeChan:
			complete++
		}
		if complete == 2 {
			break
		}
	}
	wg.Wait()
	return nil
}
func FilterEventReceivedMoneyInPast(pastTime context.Context, client sui.ISuiAPI) error {
	var currentCursor *string
	for {
		req := models.SuiXQueryTransactionBlocksRequest{
			SuiTransactionBlockResponseQuery: models.SuiTransactionBlockResponseQuery{
				TransactionFilter: models.TransactionFilter{
					"ToAddress": configuration.Wallet.AddressId,
				},
				Options: models.SuiTransactionBlockOptions{
					ShowBalanceChanges: true,
				},
			},
			Cursor:          currentCursor,
			Limit:           limitOfItemPerPage,
			DescendingOrder: true,
		}
		select {
		case <-pastTime.Done():
			return pastTime.Err()
		default:
			resp, err := QueryTransactionBlocks(client, pastTime, req)
			if err != nil {
				return err
			}
			currentCursor = &resp.NextCursor
			time.Sleep(200 * time.Millisecond)
			continue
		}
	}
}
func QueryTransactionBlocks(client sui.ISuiAPI, ctx context.Context, req models.SuiXQueryTransactionBlocksRequest) (*models.SuiXQueryTransactionBlocksResponse, error) {
	resp, err := client.SuiXQueryTransactionBlocks(ctx, req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if err := ProcessTransactionBlock(resp, ctx); err != nil {
		return nil, err
	}
	return &resp, nil
}
func FilterEventSentMoneyInPast(pastTime context.Context, client sui.ISuiAPI) error {
	var currentCurson *string
	for {
		req := models.SuiXQueryTransactionBlocksRequest{
			SuiTransactionBlockResponseQuery: models.SuiTransactionBlockResponseQuery{
				TransactionFilter: models.TransactionFilter{
					"FromAddress": configuration.Wallet.AddressId,
				},
				Options: models.SuiTransactionBlockOptions{
					ShowBalanceChanges: true,
				},
			},
			Cursor:          currentCurson,
			Limit:           limitOfItemPerPage,
			DescendingOrder: true,
		}
		select {
		case <-pastTime.Done():
			return pastTime.Err()
		default:
			resp, err := QueryTransactionBlocks(client, pastTime, req)
			if err != nil {
				return err
			}
			currentCurson = &resp.NextCursor
			time.Sleep(200 * time.Millisecond)
			continue
		}
	}
}

func FilterInRealtime(client sui.ISuiAPI, ctx context.Context, newestCheckpoint int) error {
	for {
		req := models.SuiGetCheckpointRequest{
			CheckpointID: strconv.Itoa(int(newestCheckpoint)),
		}
		_, err := client.SuiGetCheckpoint(ctx, req)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if err := HandleCheckpoint(strconv.Itoa(int(newestCheckpoint)), ctx, client); err != nil {
			fmt.Println("error check point :", newestCheckpoint, err)
			return err
		}
		newestCheckpoint++
	}
}

func HandleCheckpoint(currentCheckpoint string, ctx context.Context, client sui.ISuiAPI) error {
	var currentCurson *string
	for {
		req := models.SuiXQueryTransactionBlocksRequest{
			SuiTransactionBlockResponseQuery: models.SuiTransactionBlockResponseQuery{
				TransactionFilter: models.TransactionFilter{
					"Checkpoint": currentCheckpoint,
				},
				Options: models.SuiTransactionBlockOptions{
					ShowInput:          true,
					ShowEffects:        true,
					ShowBalanceChanges: true,
				},
			},
			Cursor:          currentCurson,
			Limit:           limitOfItemPerPage,
			DescendingOrder: true,
		}
		resp, err := client.SuiXQueryTransactionBlocks(ctx, req)
		if err != nil {
			fmt.Println(err)
			return err
		}
		for _, tx := range resp.Data {
			err := HandleBalanceChangeOfTransactionBlock(tx, ctx)
			if err != nil {
				return err
			}
		}
		currentCurson = &resp.NextCursor
		time.Sleep(200 * time.Millisecond)
	}
}

func HandleBalanceChangeOfTransactionBlock(tx models.SuiTransactionBlockResponse, ctx context.Context) error {
	digest := tx.Digest
	stringTimestamp, err := strconv.Atoi(tx.TimestampMs)
	if err != nil {
		fmt.Println("error convert timestamp", err)
		return err
	}
	timestamp := time.UnixMilli(int64(stringTimestamp))
	for _, change := range tx.BalanceChanges {
		rawAmountBigInt := new(big.Int)
		rawAmountBigInt, ok := rawAmountBigInt.SetString(change.Amount, 10)
		if !ok {
			fmt.Println("error convert amount", err)
			return err
		}
		coinType := change.CoinType
		var addressOwner AddressOwner
		if err := json.Unmarshal(change.Owner, &addressOwner); err != nil {
			fmt.Println("error unmarshal", err)
			return err
		}
		walletAddress := addressOwner.AddressOwner
		if walletAddress == configuration.Wallet.AddressId {
			amount := new(big.Float).Quo(new(big.Float).SetInt(rawAmountBigInt), new(big.Float).SetFloat64(math.Pow(10, float64(coinDecimals[coinType]))))
			amountFloat64, _ := amount.Float64()
			if err := datastore.InsertDB(walletAddress, amountFloat64, change.Amount, digest, coinType, timestamp, ctx); err != nil {
				fmt.Println("error insert db", err)
				return err
			}
			if err := SendNotification(walletAddress, amount, coinType, timestamp); err != nil {
				fmt.Println("error send notification", err)
				return err
			}
		}
	}
	return nil
}
func ProcessTransactionBlock(resp models.SuiXQueryTransactionBlocksResponse, ctx context.Context) error {
	for _, tx := range resp.Data {
		err := HandleBalanceChangeOfTransactionBlock(tx, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type AddressOwner struct {
	AddressOwner string `json:"AddressOwner"`
}

func CreateBot() (*tgbotapi.BotAPI, error) {
	botApi := os.Getenv("BOT_API")
	bot, err := tgbotapi.NewBotAPI(botApi)
	if err != nil {
		return nil, err
	}
	return bot, nil
}
func AddSign(amount *big.Float) string {
	if amount.Cmp(big.NewFloat(0)) > 0 {
		return fmt.Sprintf("+%v", amount)
	}
	return fmt.Sprintf("%v", amount)
}
func SendNotification(wallet string, amount *big.Float, coinType string, timestamp time.Time) error {
	var msg tgbotapi.MessageConfig
	teleId, err := strconv.ParseInt(os.Getenv("TELE_ID"), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TELE_ID: %v", err)
	}
	msg = tgbotapi.NewMessage(teleId, fmt.Sprintf("Wallet: %s\nBalance Change: %v\nCoin Type: %s\nAt :%v", wallet, AddSign(amount), coinType, timestamp))
	_, err = bot.Send(msg)
	if err != nil {
		fmt.Println("error sending message", err)
		return err
	}
	return nil
}
