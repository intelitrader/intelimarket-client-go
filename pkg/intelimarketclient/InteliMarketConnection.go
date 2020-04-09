package intelimarketclient

import (
	"log"
	"bitbucket.org/intelitrader/intelimarket-client-go/internal"
)



type InteliMarketConnection struct {
	c_connection *intelimarketclient.P9Connection
}

func (imc InteliMarketConnection) SubscribeGroup(group string) {

}

func (self *InteliMarketConnection) Connect(server string, port uint16) {
	log.Printf("Connecting to %v:%v", server, port)

	self.c_connection = intelimarketclient.P9mdi_connect(server, port)
}

func (self *InteliMarketConnection) Disconnect() {
	intelimarketclient.P9mdi_disconnect(self.c_connection)

	self.c_connection = nil
}