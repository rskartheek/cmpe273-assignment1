package main

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"
	"os"
)

//Args struct type
type Args struct {
	A string `json:“stockSymbolAndPercentage”`
	B string `json:“budget”`
}

type StockResponse struct {
	tradeID        int
	responseString string
	balance        float64
}

func main() {
	service := "localhost:1234"
	if len(os.Args) == 2 {
		client, err := jsonrpc.Dial("tcp", service)
		if err != nil {
			log.Fatal("dialing:", err)
		}

		var reply string
		err = client.Call("Stock.GetStockDetails", os.Args[1], &reply)
		if err != nil {
			log.Fatal("stock error:", err)
		}
		fmt.Println(reply)
	}
	if len(os.Args) < 2 {
		fmt.Println("Usage: ", os.Args[0], "server")
		os.Exit(1)
	}
	if len(os.Args) > 2 {

		client, err := jsonrpc.Dial("tcp", service)
		if err != nil {
			log.Fatal("dialing:", err)
		}

		args := Args{os.Args[1], os.Args[2]}
		var reply string
		err = client.Call("Stock.Buy", args, &reply)
		if err != nil {
			log.Fatal("stock error:", err)
		}
		fmt.Println(reply)
	}
}
