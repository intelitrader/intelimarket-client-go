package main

import (
	"fmt"
	"bitbucket.org/intelitrader/intelimarket-client-go/pkg/intelimarketclient"
)

func PropertyToStdOut(channel <-chan intelimarketclient.PropertyChangeInfo) {

	fmt.Println("PropertyToStdOut, reading", channel)

	for ;; {
		info := <-channel
		fmt.Println("Property change:", info)
	}
}

func main() {
	connection := intelimarketclient.InteliMarketConnection{}

	propertyChangeChannel, err := connection.Connect("demo.intelitrader.com.br", 2605)
	defer connection.Disconnect()

	if err != nil {
		fmt.Println("error: ", err);
		return
	}

	connection.SubscribeInstrumentProperties("PETR4")

	go PropertyToStdOut(propertyChangeChannel)

	for ;; {
		connection.DispatchPendingMessage()
	}	
}