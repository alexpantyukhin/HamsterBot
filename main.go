package main

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/alexpantyukhin/btceapi"
)

func checkStringParameterNotEmpty(flag string, value string) {
	if len(value) == 0 {
		panic("ERROR:" + flag + "parameter can not be empty.")
	}
}

func checkIntParameterFloatEmpty(flag string, value float64) {
	if value == 0 {
		panic("ERROR:" + flag + "parameter can not be empty.")
	}
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

func checkError(err error) {
	if err != nil {
		panic("ERROR: " + err.Error())
	}
}

func trade(btceAPI btceapi.BtceAPI, pair string, enterBound, exitBound, startAmount float64) {
	amount := startAmount
	start := time.Now().UTC()
	tradeAnswer, err := btceAPI.Trade(pair, "sell", enterBound, amount)
	checkError(err)
	orderID := tradeAnswer.OrderID
	sell := false

	for {
		time.Sleep(time.Second)

		tradeHistory, tradeFound := getTradeHistoryByOrder(btceAPI, orderID, start)
		if tradeFound {
			start = time.Unix(tradeHistory.Timestamp, 0)

			if sell {
				tradeAnswer, err = btceAPI.Trade(pair, "sell", enterBound, amount)
			} else {
				amount = tradeHistory.Amount * enterBound / exitBound
				tradeAnswer, err = btceAPI.Trade(pair, "buy", exitBound, amount)
			}

			checkError(err)
			orderID = tradeAnswer.OrderID
			sell = !sell
		}
	}
}

func main() {
	btceapi.ApiURL = "https://wex.nz"

	var key string
	var secret string
	var pair string
	var enterBoundString string
	var exitBoundString string
	var amountString string

	flag.StringVar(&key, "key", "", "Your API key")
	flag.StringVar(&secret, "secret", "", "Your API secret key")
	flag.StringVar(&pair, "pair", "", "Pair which you want to trade.")
	flag.StringVar(&enterBoundString, "enter", "", "Price which you want to .")
	flag.StringVar(&exitBoundString, "exit", "", "Price which you want to .")
	flag.StringVar(&amountString, "amount", "", "Your amount")

	//TODO: make normal checking of parameters.
	checkStringParameterNotEmpty("key", key)
	checkStringParameterNotEmpty("secret", secret)
	checkStringParameterNotEmpty("pair", pair)

	btceAPI := btceapi.BtceAPI{Key: key, Secret: secret}
	info, err := btceAPI.GetInfo()

	if err != nil {
		panic("ERROR: Can not connect to exchange: " + err.Error())
	}

	if info.Rights.Trade == 0 || info.Rights.Info == 0 {
		panic("ERROR: not enough rights: ")
	}

	amount, _ := strconv.ParseFloat(amountString, 64)
	enterBound, _ := strconv.ParseFloat(enterBoundString, 64)
	exitBound, _ := strconv.ParseFloat(exitBoundString, 64)

	firstCurrency := strings.Split(pair, "_")[0]
	firstCurrencyBalance := info.Funds[firstCurrency]

	if firstCurrencyBalance < amount {
		amount = firstCurrencyBalance
	}

	trade(btceAPI, pair, enterBound, exitBound, amount)
}
