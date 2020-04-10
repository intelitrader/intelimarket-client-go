package intelimarketclient

import (
	"fmt"
	"bitbucket.org/intelitrader/intelimarket-client-go/internal"
)

type PropertyChangeInfo struct {
	exchange, symbol string
	propertyKey, propertyValue string
}


type InteliMarketConnection struct {
	c_connection *intelimarketclient.P9GoConnection
	hostname string
	port uint16
	propertyChangeChannel chan PropertyChangeInfo
}

func (self InteliMarketConnection) SubscribeGroup(group string) {

}

func p9_OnPropertyCallback(eventCookie interface{}, exchange string, symbol string, key string, value string) {
	intelimarketConnection := eventCookie.(*InteliMarketConnection)

	intelimarketclient.LogTrace("p9_OnPropertyCallback", intelimarketConnection, symbol, key, value)

	if key == "" {
		return
	}

	intelimarketConnection.propertyChangeChannel <- PropertyChangeInfo{exchange, symbol, key, value}
}

func (self *InteliMarketConnection) String() string {
	return fmt.Sprintf("<InteliMarketConnection %v:%v>", self.hostname, self.port)
}

func (self *InteliMarketConnection) Connect(server string, port uint16) (<-chan PropertyChangeInfo, error) {
	intelimarketclient.LogTrace("Connecting to %v:%v", server, port)

	self.c_connection = intelimarketclient.P9mdi_connect(server, port, p9_OnPropertyCallback, self)

	self.propertyChangeChannel = make(chan PropertyChangeInfo, 1024 * 1024)

	self.hostname = server
	self.port = port

	return self.propertyChangeChannel, nil
}

func (self *InteliMarketConnection) DispatchPendingMessage() {
	intelimarketclient.P9mdi_dispatch_pending_events(self.c_connection)
}

func (self *InteliMarketConnection) Disconnect() {
	if self.c_connection == nil {
		return
	}
	intelimarketclient.P9mdi_disconnect(self.c_connection)
	close(self.propertyChangeChannel)

	self.c_connection = nil
	self.propertyChangeChannel = nil
}

func (self *InteliMarketConnection) SubscribeInstrumentProperties(symbol string) {
	intelimarketclient.P9mdi_subscribe_instrument_properties(self.c_connection, symbol)
}
