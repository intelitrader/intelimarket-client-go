package intelimarketclient

import (
	"fmt"
	"strings"
    "runtime"
    "strconv"

	intelimarketclient "bitbucket.org/intelitrader/intelimarket-client-go/internal"
)

type EventCode = uint

const (
	EventCodeSet          uint = 0x14
	EventCodeInsert       uint = 0x15
	EventCodeDelete       uint = 0x16
	EventCodePushBack     uint = 0x17
	EventCodePushFront    uint = 0x18
	EventCodePopBack      uint = 0x19
	EventCodePopFront     uint = 0x1A
	EventCodeClear        uint = 0x1B
	EventCodeSnapshotEnd  uint = 0x23
	EventCodeSubscription uint = 0x66
)

type PropertyChangeInfo struct {
	Exchange, Symbol string
	EventCode        EventCode
	Key, Value       string
}

type TradeChangeInfo struct {
	Exchange, Symbol     string
	Buyer, Seller        string
	TradeId              string
	Price                string
	Quantity             string
	NetChangePreviousDay string
	Date                 string
	Time                 string
	Position             string
}

func LogTrace(format string, args ...interface{}) {
	intelimarketclient.LogTrace(format, args...)
}

func eventCodeToString(eventCode EventCode) string {
	switch eventCode {
	case EventCodeSet:
		return "Set"
	case EventCodeInsert:
		return "Insert"
	case EventCodeDelete:
		return "Delete"
	case EventCodePushBack:
		return "PushBack"
	case EventCodePushFront:
		return "PushFront"
	case EventCodePopBack:
		return "PopBack"
	case EventCodePopFront:
		return "PopFront"
	case EventCodeClear:
		return "Clear"
	case EventCodeSnapshotEnd:
		return "SnapshotEnd"
	case EventCodeSubscription:
		return "Subscription"
	default:
		return "WHAT?"
	}
}

func (self PropertyChangeInfo) String() string {
	ret := fmt.Sprintf("%v.%v - %v",
		self.Exchange,
		self.Symbol,
		eventCodeToString(self.EventCode))

	if self.Key != "" {
		ret += fmt.Sprintf(" - %v=%v",
			self.Key,
			self.Value)
	}
	return ret
}

func (self TradeChangeInfo) String() string {
	ret := fmt.Sprintf("%v.%v.%v - %v@%v - %v<-->%v %v",
		self.Exchange,
		self.Symbol,
		self.TradeId,
		self.Quantity,
		self.Price,
		self.Buyer,
		self.Seller,
        self.Position)

	return ret
}

type InteliMarketConnection struct {
	c_connection          *intelimarketclient.P9GoConnection
	hostname              string
	port                  uint16
	propertyChangeChannel chan PropertyChangeInfo
	tradeChangeChannel    chan TradeChangeInfo
}

func p9_OnPropertyCallback(eventCookie interface{}, eventCode uint32, exchange string, symbol string, key string, value string) {
	intelimarketConnection := eventCookie.(*InteliMarketConnection)

	LogTrace("p9_OnPropertyCallback", intelimarketConnection, symbol, key, value)

	intelimarketConnection.propertyChangeChannel <- PropertyChangeInfo{exchange, symbol, EventCode(eventCode), key, value}
}

func p9_OnTradeCallback(eventCookie interface{}, eventCode uint32, exchange string, symbol string, position uint32, fields map[string]string) {
	intelimarketConnection := eventCookie.(*InteliMarketConnection)

	LogTrace("p9_OnTradeCallback connection:%v symbol:%v position:%v fields:%v", intelimarketConnection, symbol, position, fields)

	tradeInfo := TradeChangeInfo{}

    tradeInfo.Exchange = exchange
    tradeInfo.Symbol = symbol
    tradeInfo.Position = strconv.FormatUint(uint64(position), 10)

	for k, v := range fields { 
		switch k {
		case "MDEntryBuyer":
			tradeInfo.Buyer = v
		case "MDEntrySeller":
			tradeInfo.Seller = v
		case "TradeID":
			tradeInfo.TradeId = v
		case "MDEntryPx":
			tradeInfo.Price = v
		case "MDEntrySize":
			tradeInfo.Quantity = v
		case "NetChgPrevDay":
			tradeInfo.NetChangePreviousDay = v
		case "MDEntryDate":
			tradeInfo.Date = v
		case "MDEntryTime":
			tradeInfo.Time = v
		}
	}

    if len(tradeInfo.TradeId) > 0 {
	    intelimarketConnection.tradeChangeChannel <- tradeInfo
    }
}

func (self *InteliMarketConnection) String() string {
	return fmt.Sprintf("<InteliMarketConnection %v:%v>", self.hostname, self.port)
}

func (self *InteliMarketConnection) GetPropertyChangeChannel() <-chan PropertyChangeInfo {
	return self.propertyChangeChannel
}

func (self *InteliMarketConnection) GetTradeChangeChannel() <-chan TradeChangeInfo {
	return self.tradeChangeChannel
}

func LogStats(event string) {
    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)
    var mb = mem.TotalAlloc / 1024 / 1024
    line := fmt.Sprintf("event: %s, memory %dMB", event, mb)
    LogTrace(line)
}

func (self *InteliMarketConnection) Connect(server string, port uint16, logpath string, chansize int) error {
	var err error

    LogStats("intelimarketclient::connecting")
	self.c_connection, err = intelimarketclient.P9mdi_connect(server, port, logpath, p9_OnPropertyCallback, p9_OnTradeCallback, self)
	LogTrace("Connecting to %v:%v", server, port)

	if err != nil {
		return err
	}

    LogStats("intelimarketclient::connected")
	self.propertyChangeChannel = make(chan PropertyChangeInfo, chansize)
	self.tradeChangeChannel = make(chan TradeChangeInfo, chansize)
    LogStats("intelimarketclient::channels_created")

	self.hostname = server
	self.port = port

	return nil
}

func (self *InteliMarketConnection) DispatchPendingMessage(timeoutSeconds int) int {
	if self.c_connection != nil {
        LogStats("intelimarketclient::dispatching")
        result := intelimarketclient.P9mdi_dispatch_pending_events(self.c_connection, timeoutSeconds)
        LogStats("intelimarketclient::dispatched")
        if result == -6 {
            LogTrace("DispatchPendingMessage timeout; sending ping")
            pingResult := intelimarketclient.P9mdi_ping(self.c_connection, "intelimarket-go")
            if pingResult < 0 {
                return pingResult
            }
        }
        return result
    } else {
        return -2
    }
}

func (self *InteliMarketConnection) Disconnect() {
	if self.c_connection == nil {
		return
	}
	intelimarketclient.P9mdi_disconnect(self.c_connection)
	close(self.propertyChangeChannel)
	close(self.tradeChangeChannel)

	self.c_connection = nil
	self.propertyChangeChannel = nil
	self.tradeChangeChannel = nil
}

func (self *InteliMarketConnection) SubscribeInstrumentProperties(symbol string) {
	intelimarketclient.P9mdi_subscribe_instrument_properties(self.c_connection, symbol)
}

func (self *InteliMarketConnection) SubscribeInstrumentTrades(symbol string, position string) {
	int_position, _ := strconv.ParseInt(position, 10, 32)
	intelimarketclient.P9mdi_subscribe_instrument_trades(self.c_connection, symbol, int32(int_position))
}

func (self *InteliMarketConnection) SubscribeGroupProperties(groupName string) {
	//
	// Eu descobri esse nome olhando os grupos que o umdf_feeder gera pelo TioExplorer
	// (tudo que começa com __meta__/groups dentro do tio)
	//
	tioGroupName := fmt.Sprintf("intelimarket/security_type/%v/properties", strings.ToLower(groupName))
	intelimarketclient.P9mdi_subscribe_group(self.c_connection, tioGroupName, 0)
}

func (self *InteliMarketConnection) SubscribeGroupTrades(groupName string, position string) {
	//
	// Eu descobri esse nome olhando os grupos que o umdf_feeder gera pelo TioExplorer
	// (tudo que começa com __meta__/groups dentro do tio)
	//
	int_position, _ := strconv.ParseInt(position, 10, 32)
	tioGroupName := fmt.Sprintf("intelimarket/security_type/%v/trades", strings.ToLower(groupName))
	intelimarketclient.P9mdi_subscribe_group(self.c_connection, tioGroupName, -int32(int_position))
}
