package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexpantyukhin/btceapi"
)

func checkStringParameterNotEmpty(flag string, value string) {
	if len(value) == 0 {
		showErrorAndExit(fmt.Sprintf("%s parameter can not be empty.", flag))
	}
}

func showErrorAndExit(str string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s", str)
	os.Exit(1)
}

func checkError(err error) {
	if err != nil {
		showErrorAndExit(fmt.Sprintf("%v\n", err))
	}
}

func logMessage(str string) {
	fmt.Printf("Info: " + str)
}

func getTradeHistoryByOrder(btceAPI btceapi.BtceAPI, orderID int, since time.Time) (btceapi.Trade, bool) {
	tradeHistory, err := btceAPI.GetTradeHistory(btceapi.FilterParams{Since: since, End: time.Now().UTC()})
	if err != nil {
		panic("Error: can't get trade history")
	}

	for _, tradeHistory := range tradeHistory {
		if tradeHistory.OrderID == orderID {
			return tradeHistory, true
		}
	}

	return btceapi.Trade{}, false
}

func trade(btceAPI btceapi.BtceAPI, pair string, enterBound, exitBound, startAmount float64) {
	amount := startAmount
	start := time.Now().UTC()
	tradeAnswer, err := btceAPI.Trade(pair, "sell", enterBound, amount)
	checkError(err)
	logMessage(fmt.Sprintf("SELL order pair \"%s\", amount \"%f\", price \"%f\"", pair, amount, enterBound))
	orderID := tradeAnswer.OrderID
	sell := false

	for {
		time.Sleep(time.Second)

		tradeHistory, tradeFound := getTradeHistoryByOrder(btceAPI, orderID, start)
		if tradeFound {
			start = time.Unix(tradeHistory.Timestamp, 0)
			var tradeType string
			var price float64

			if sell {
				tradeAnswer, err = btceAPI.Trade(pair, "sell", enterBound, amount)
				tradeType = "SELL"
				price = enterBound
			} else {
				amount = tradeHistory.Amount * enterBound / exitBound
				tradeAnswer, err = btceAPI.Trade(pair, "buy", exitBound, amount)
				tradeType = "BUY"
				price = exitBound
			}

			checkError(err)
			logMessage(fmt.Sprintf("%s order pair \"%s\", amount \"%f\", price \"%f\"", tradeType, pair, amount, price))
			orderID = tradeAnswer.OrderID
			sell = !sell
		}
	}
}

var key string
var secret string
var pair string
var enter float64
var exit float64
var amount float64

func parseFloatOrPanic(val string, paramName string) float64 {
	param, err := strconv.ParseFloat(val, 64)
	checkError(err)

	return param
}

func initAndCheckFlags() {
	var enterBoundString string
	var exitBoundString string
	var amountString string

	flag.StringVar(&key, "key", "", "Your API key")
	flag.StringVar(&secret, "secret", "", "Your API secret key")
	flag.StringVar(&pair, "pair", "", "Pair which you want to trade.")
	flag.StringVar(&enterBoundString, "enter", "", "Price which you want to .")
	flag.StringVar(&exitBoundString, "exit", "", "Price which you want to .")
	flag.StringVar(&amountString, "amount", "", "Your amount")

	flag.Parse()

	checkStringParameterNotEmpty("key", key)
	checkStringParameterNotEmpty("secret", secret)
	checkStringParameterNotEmpty("pair", pair)
	checkStringParameterNotEmpty("enter", enterBoundString)
	checkStringParameterNotEmpty("exit", exitBoundString)
	checkStringParameterNotEmpty("amount", amountString)

	amount = parseFloatOrPanic(amountString, "amount")
	enter = parseFloatOrPanic(enterBoundString, "enter")
	exit = parseFloatOrPanic(exitBoundString, "exit")
}

func main() {
	btceapi.ApiURL = "https://wex.nz"

	initAndCheckFlags()

	btceAPI := btceapi.BtceAPI{Key: key, Secret: secret}
	info, err := btceAPI.GetInfo()

	if err != nil {
		showErrorAndExit("Can not connect to exchange: " + err.Error())
	}

	if info.Rights.Trade == 0 || info.Rights.Info == 0 {
		showErrorAndExit("Not enough rights for trading. Please update the key privileges.")
	}

	firstCurrency := strings.Split(pair, "_")[0]
	firstCurrencyBalance := info.Funds[firstCurrency]

	if firstCurrencyBalance < amount {
		amount = firstCurrencyBalance
	}

	trade(btceAPI, pair, enter, exit, amount)
}
