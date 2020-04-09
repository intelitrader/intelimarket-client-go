package main

import (
	"bitbucket.org/intelitrader/intelimarket-client-go/pkg/intelimarketclient"
)

func main() {

	connection := intelimarketclient.InteliMarketConnection{}

	connection.Connect("demo.intelitrader.com.br", 2605)

	print("e ae debug?\n")

}