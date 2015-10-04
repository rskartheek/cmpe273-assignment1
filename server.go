package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
	"time"
)

//random function - function to calculate random numbers
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

//Args struct
type Args struct {
	A string
	B string
}

//SavedStockObjects struct - To save stock data in memory

type savedStockObject struct {
	tradeID int
	stocks  []purchasedStock
}

// purchasedStock struct - To  save stock  data in memory
type purchasedStock struct {
	symbol         string
	numberOfStocks int
	amountSpent    float64
	price          float64
}

//array to store purchased stocks for using in GetStockDetails function
var stockDataObjectArray []purchasedStock

// stockObjectArray - To save arrays of purchasedStock items
var stockObjectArray []savedStockObject

var stockDataObject savedStockObject

var stockObjectMap map[int]savedStockObject

//Stock struct is defined here
type Stock struct {
	List struct {
		Resources []struct {
			Resource struct {
				Fields struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					UTCTime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}

//StockResponse struct is defined here
type StockResponse struct {
	tradeID        int
	investedStocks string
	balance        float64
}

type oldStockValues struct {
	oldValues []purchasedStock
}

//Buy method
func (t *Stock) Buy(args *Args, reply *string) error {
	budget, err := strconv.ParseFloat(args.B, 64)
	if err != nil {
		fmt.Println("Error", err)
	}
	// Splitting strings using ,'
	stocks := strings.Split(args.A, ",")
	stockMap := make(map[string]string)
	for _, stock := range stocks {
		stockPercent := strings.Split(stock, ":")
		stockMap[stockPercent[0]] = stockPercent[1]
	}
	stockBudget := make(map[string]float64)
	var queryParamBuilder string
	count := 0
	for k, v := range stockMap {
		pct, err := strconv.ParseFloat(v, 64)
		if err == nil {
			//Calculating amount for buying stocks
			amtForStock := (pct * budget) / 100
			stockBudget[k] = amtForStock
		}
		if count == 0 {
			queryParamBuilder = k
		} else if len(stockMap) > count {
			queryParamBuilder = queryParamBuilder + "," + k
		}
		count++
	}
	url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", queryParamBuilder)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	// read json http response
	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var stock Stock

	err = json.Unmarshal(jsonDataFromHttp, &stock) // here!
	if err != nil {
		panic(err)
	}

	var totalRem float64
	var responseString string
	//Function for forming responseString
	var details StockResponse
	details.tradeID = random(1, 100000)
	stockDataObject.tradeID = details.tradeID
	for k := 0; k < len(stock.List.Resources); k++ {
		var purchasedStockObject purchasedStock
		symbol := stock.List.Resources[k].Resource.Fields.Symbol
		amount := stockBudget[symbol]
		price := stock.List.Resources[k].Resource.Fields.Price
		convPrice, err := strconv.ParseFloat(price, 64)
		purchasedStockObject.symbol = symbol
		if err == nil {
			quo := amount / convPrice
			purchasedStockObject.numberOfStocks = int(quo)
			remainder := math.Mod(amount, convPrice)
			amountUsed := amount - remainder
			responseString = responseString + " " + stock.List.Resources[k].Resource.Fields.Symbol + " :" + strconv.Itoa(int(quo)) + ":" +
				price
			totalRem = totalRem + remainder
			purchasedStockObject.amountSpent = amountUsed
			purchasedStockObject.price = convPrice
		}
		stockDataObject.stocks = append(stockDataObject.stocks, purchasedStockObject)

	}

	details.investedStocks = responseString
	details.balance = totalRem
	*reply = "TradeID " + strconv.Itoa(details.tradeID) + "\n" + "stocks " + details.investedStocks + "\n" + "unvestedAmount " + strconv.FormatFloat(details.balance, 'f', -1, 64)
	fmt.Println(details)
	stockObjectArray = append(stockObjectArray, stockDataObject)
	//stockObjectMap[details.tradeID] = stockDataObject
	return nil
}

//StockDetails struct - To get StockDetails based on tradeID
func (t *Stock) GetStockDetails(id string, reply *string) error {
	// var oldValuesArray []oldStockValues
	var queryParamBuilder string
	intID, err := strconv.Atoi(id)
	if err == nil {
		for i := 0; i < len(stockObjectArray); i++ {
			if stockObjectArray[i].tradeID == intID {
				count := 0
				// oldValuesArray := stockObjectArray[i].stocks
				arrayLength := len(stockObjectArray[i].stocks)
				stockDataObjectArray = stockObjectArray[i].stocks
				for k := 0; k < arrayLength; k++ {
					if count == 0 {
						queryParamBuilder = stockObjectArray[i].stocks[k].symbol
					} else if arrayLength > count {
						queryParamBuilder = queryParamBuilder + "," + stockObjectArray[i].stocks[k].symbol
					}
					count++
				}
			}
		}
		url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", queryParamBuilder)

		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		// read json http response
		jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}

		var stock Stock

		err = json.Unmarshal(jsonDataFromHttp, &stock) // here!
		if err != nil {
			panic(err)
		}

		stockQueryData := stock.List.Resources
		var responseString string
		var currentValue float64
		// test struct data
		for i := 0; i < len(stockDataObjectArray); i++ {
			for k := 0; k < len(stockQueryData); k++ {
				if stockDataObjectArray[i].symbol == stockQueryData[k].Resource.Fields.Symbol {
					newPrice := stockQueryData[k].Resource.Fields.Price
					convertedPrice, err := strconv.ParseFloat(newPrice, 64)
					if err != nil {
						fmt.Println("Error")
					}
					oldPrice := stockDataObjectArray[i].price
					if (oldPrice - convertedPrice) > 0 {
						responseString = responseString + stockDataObjectArray[i].symbol + ":" + strconv.Itoa(stockDataObjectArray[i].numberOfStocks) + ": -" + newPrice + " "
					} else if (oldPrice - convertedPrice) < 0 {
						responseString = responseString + stockDataObjectArray[i].symbol + ":" + strconv.Itoa(stockDataObjectArray[i].numberOfStocks) + ": +" + newPrice + " "
					} else if (oldPrice - convertedPrice) == 0 {
						responseString = responseString + stockDataObjectArray[i].symbol + ":" + strconv.Itoa(stockDataObjectArray[i].numberOfStocks) + ": " + newPrice + " "
					}
					currentValue = currentValue + (convertedPrice * float64(stockDataObjectArray[i].numberOfStocks))
				}
			}
		}
		*reply = "stock" + ":" + responseString + "\n" + "currentMarketValue: " + strconv.FormatFloat(currentValue, 'f', -1, 64)
		
	}
	return nil
}
func main() {
	stock := new(Stock)
	rpc.Register(stock)

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1234")
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		jsonrpc.ServeConn(conn)
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
