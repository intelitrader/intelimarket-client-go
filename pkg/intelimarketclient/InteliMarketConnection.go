package intelimarketclient

import (
	"fmt"
	"strings"
    "runtime"
    "strconv"

	intelimarketclient "github.com/intelitrader/intelimarket-client-go/internal"
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

//
// Códigos de erro do uso da lib InteliMarket/Tio.
//
// Esses códigos são gerados internamentes pela lib. Algumas APIs possuem o retorno
// desses códigos mais um código de erro interno. O erro interno se refere a chamadas
// do sistema operacional ou chamadas internas não-documentadas nesta lib.
//
// Para verificar o comportamento de determinada operação que retorne um código de erro
// interno consultar a seguinte documentação:
//
// Se estiver executando a lib em ambiente Windows:
//  - System Error Codes (ref https://docs.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-)
//  - Windows Sockets Error Codes (ref https://docs.microsoft.com/en-us/windows/win32/winsock/windows-sockets-error-codes-2)
//
// Se estiver executando a lib em ambiente Linux/UNIX:
//  - Man errno ou Linux Error Codes for C Programming Language (ref https://www.thegeekstuff.com/2010/10/linux-error-codes/)
//
const (
  // Sucesso ao realizar a operação.
	TIO_SUCCESS = 0
  // Erro ao realizar a operação.
	TIO_ERROR_GENERIC = -1
  // Erro de rede.
	TIO_ERROR_NETWORK = -2
  // Erro de protocolo.
	TIO_ERROR_PROTOCOL = -3
  // Faltando parâmetro na operação.
	TIO_ERROR_MISSING_PARAMETER = -4
  // Objeto não encontrado.
	TIO_ERROR_NO_SUCH_OBJECT = -5
  // Timeout ao realizar a operação.
	TIO_ERROR_TIMEOUT = -6
  // Memória não disponível para continuar operação.
	TIO_ERROR_OUT_OF_MEMORY = -7
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

type BookChangeInfo struct {
	Exchange, Symbol     string
	OrderId              string
	Broker               string
	Price                string
	Quantity             string
	Date                 string
	Time                 string
	Position             string
	Side                 string
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

func (self BookChangeInfo) String() string {
	ret := fmt.Sprintf("%v.%v -> %v<-->%v@%v -> %v",
		self.Exchange,
		self.Symbol,
		self.Price,
		self.Quantity,
		self.Broker,
		self.Position)

	return ret
}

type InteliMarketConnection struct {
	c_connection          *intelimarketclient.P9GoConnection
	hostname              string
	port                  uint16
	propertyChangeChannel chan PropertyChangeInfo
	tradeChangeChannel    chan TradeChangeInfo
	bookChangeChannel 	  chan BookChangeInfo
    chansize              int
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

func p9_OnBookCallback(eventCookie interface{}, eventCode uint32, exchange string, symbol string, position uint32, fields map[string]string) {
	intelimarketConnection := eventCookie.(*InteliMarketConnection)

	LogTrace("p9_OnBookCallback connection:%v symbol:%v position:%v fields:%v", intelimarketConnection, symbol, position, fields)

	bookInfo := BookChangeInfo{}

    bookInfo.Exchange = exchange
    bookInfo.Symbol = symbol
    bookInfo.Position = strconv.FormatUint(uint64(position), 10)

	for k, v := range fields { 
		switch k {
		case "OrderID":
			bookInfo.OrderId = v
		case "MDEntryPx":
			bookInfo.Price = v
		case "MDEntrySize":
			bookInfo.Quantity = v
		case "MDEntryDate":
			bookInfo.Date = v
		case "MDEntryTime":
			bookInfo.Time = v
		case "Broker":
			bookInfo.Broker = v
		case "Side":
			bookInfo.Side = v
		}
	}
	if len(bookInfo.OrderId) > 0 {
		intelimarketConnection.bookChangeChannel <- bookInfo
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

func (self *InteliMarketConnection) GetBookChangeChannel() <-chan BookChangeInfo {
	return self.bookChangeChannel
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
	self.c_connection, err = intelimarketclient.P9mdi_connect(server, port, logpath, p9_OnPropertyCallback, p9_OnTradeCallback, p9_OnBookCallback, self)
	LogTrace("Connecting to %v:%v", server, port)

	if err != nil {
		return err
	}

    LogStats("intelimarketclient::connected")

	self.hostname = server
	self.port = port
    self.chansize = chansize

	self.tradeChangeChannel = make(chan TradeChangeInfo, self.chansize)
		LogStats("intelimarketclient::trade_channel_created")
	self.propertyChangeChannel = make(chan PropertyChangeInfo, self.chansize)
		LogStats("intelimarketclient::property_channel_created")
	self.bookChangeChannel = make(chan BookChangeInfo, self.chansize)
		LogStats("intelimarketclient::book_channel_created")
		
	return nil
}

/* A função DispatchPendingMessage é blocante e deve ser chamada em loop até o assinante desejar continuar
recebendo eventos. Ela fica aguardando pelo número de segundos especificado no parâmetro timeoutSeconds 
pelo recebimento de eventos. Ao finalizar o número de segundos especificado é retornado ao chamador 
o status TIO_ERROR_TIMEOUT (-6).

O erro mais comum retornado fora o timeout é o TIO_ERROR_NETWORK (-2), quando acontece um problema 
de rede como perda de conexão.

Junto do código de erro conhecido é fornecido um código de erro interno que está relacionado a
chamadas de funções internas ou do sistema operacional. No caso de erros de rede, por exemplo,
o erro interno irá conter um código que pode ser buscado na listagem de erros de rede do
sistema operacional. No caso do Windows o erro será de winsock, mas no caso de Linux há
casos que o erro interno não é significativo e irá conter lixo (não-implementado).

Porém, há exemplos de uso em que o erro interno ocorrido junto de -2 foi 104, o que significa
"Connection reset by peer", casos em que o InteliMarket fecha a conexão de um assinante que não
respondeu aos eventos em tempo hábil para evitar consumo excessivo de recursos do lado servidor.

Outro erro interno reportado é o código 4, ou "Interrupted system call" (EINTR), que significa
que um signal foi disparado durante uma chamada em progresso. O cancelamento de uma operação
de rede pode retornar este código.
*/
func (self *InteliMarketConnection) DispatchPendingMessage(timeoutSeconds int) (int, int) {
	if self.c_connection != nil {
        internalError := 0
        LogStats("intelimarketclient::dispatching")
        result := intelimarketclient.P9mdi_dispatch_pending_events(self.c_connection, timeoutSeconds)
        internalError = intelimarketclient.P9mdi_get_last_error(self.c_connection)
        LogStats(fmt.Sprintf("intelimarketclient::dispatched (result %d, internal error %d)", result, internalError))
        if result == -6 {
            LogTrace("DispatchPendingMessage timeout; sending ping")
            pingResult := intelimarketclient.P9mdi_ping(self.c_connection, "intelimarket-go")
            if pingResult < 0 {
                internalError = intelimarketclient.P9mdi_get_last_error(self.c_connection)
                return pingResult, internalError
            }
        }
        return result, internalError
    } else {
        return -2, 0
    }
}

func (self *InteliMarketConnection) Disconnect() {
	if self.c_connection == nil {
		return
	}
	intelimarketclient.P9mdi_disconnect(self.c_connection)

	self.c_connection = nil
}

func (self *InteliMarketConnection) SubscribeInstrumentProperties(symbol string) {
	if self.propertyChangeChannel == nil {
		self.propertyChangeChannel = make(chan PropertyChangeInfo, self.chansize)
		LogStats("intelimarketclient::property_channel_created")
	}

	intelimarketclient.P9mdi_subscribe_instrument_properties(self.c_connection, symbol)
}

/*
A função SubscribeInstrumentTrades assina eventos de trade e permite receber eventos passados através 
do parâmetro position. Esse parâmetro é chamado internamente de snapshotSize e significa a quantidade 
de eventos que será entregue ao assinante antes dos eventos incrementais (o que seguem após a assinatura 
ser concluída). Se passado 0 apenas os eventos incrementais serão repassados, o que chamamos internamente 
de modo de assinatura IncrementalOnly. Se passado um valor diferente de 0 o modo de assinatura será 
SnapshotPlusIncremental, onde os últimos position eventos antes dos eventos incrementais serão 
repassados ao assinante.
*/
func (self *InteliMarketConnection) SubscribeInstrumentTrades(symbol string, position string) {
	if self.tradeChangeChannel == nil {
		self.tradeChangeChannel = make(chan TradeChangeInfo, self.chansize)
		LogStats("intelimarketclient::trade_channel_created")
	}

	int_position, _ := strconv.ParseInt(position, 10, 32)
	intelimarketclient.P9mdi_subscribe_instrument_trades(self.c_connection, symbol, int32(int_position))
}

func (self *InteliMarketConnection) SubscribeInstrumentOrderBook(symbol string, orderBookSize string) {
	if self.bookChangeChannel == nil {
		self.bookChangeChannel = make(chan BookChangeInfo, self.chansize)
		LogStats("intelimarketclient::book_channel_created")
	}

	int_size, _ := strconv.ParseInt(orderBookSize, 10, 32)

	intelimarketclient.P9mdi_subscribe_instrument_order_book(self.c_connection, symbol, 1, int32(int_size)) //BOOK_SIDE_BUY
	intelimarketclient.P9mdi_subscribe_instrument_order_book(self.c_connection, symbol, 2, int32(int_size)) //BOOK_SIDE_SELL
}

func (self *InteliMarketConnection) SubscribeGroupProperties(groupName string) {
	//
	// Eu descobri esse nome olhando os grupos que o umdf_feeder gera pelo TioExplorer
	// (tudo que começa com __meta__/groups dentro do tio)
	//
	tioGroupName := fmt.Sprintf("intelimarket/security_type/%v/properties", strings.ToLower(groupName))
	intelimarketclient.P9mdi_subscribe_group(self.c_connection, tioGroupName, 0)
}

/*
A função SubscribeGroupTrades assina eventos de trade de um grupo de instrumentos e permite receber 
eventos passados através do parâmetro position. Esse parâmetro é chamado internamente de snapshotSize 
e significa a quantidade de eventos que será entregue ao assinante antes dos eventos incrementais 
(o que seguem após a assinatura ser concluída). Se passado 0 todos os eventos antes dos incrementais serão 
repassados. Se passado um valor maior que 0 os últimos position eventos antes dos eventos incrementais serão 
repassados ao assinante.

Note que o parâmetro position para assinatura de grupos se refere aos eventos dos instrumentos 
individualmente, e não em conjunto. Por exemplo, se for usado um position 100 serão enviados os útimos 100
eventos para cada instrumento pertencente ao grupo assinado. Caso um instrumento não possua essa quantidade
de eventos serão enviados menos eventos. Caso um instrumento possua mais eventos que essa quantidade
serão enviados no máximo os últimos 100 eventos.
*/
func (self *InteliMarketConnection) SubscribeGroupTrades(groupName string, position string) {
	//
	// Eu descobri esse nome olhando os grupos que o umdf_feeder gera pelo TioExplorer
	// (tudo que começa com __meta__/groups dentro do tio)
	//
	int_position, _ := strconv.ParseInt(position, 10, 32)
	tioGroupName := fmt.Sprintf("intelimarket/security_type/%v/trades", strings.ToLower(groupName))
	intelimarketclient.P9mdi_subscribe_group(self.c_connection, tioGroupName, -int32(int_position))
}

func (self *InteliMarketConnection) SubscribeGroupBook(groupName string) {
	//
	// Eu descobri esse nome olhando os grupos que o umdf_feeder gera pelo TioExplorer
	// (tudo que começa com __meta__/groups dentro do tio)
	//
	tioGroupName := fmt.Sprintf("intelimarket/security_type/%v/book", strings.ToLower(groupName))
	intelimarketclient.P9mdi_subscribe_group(self.c_connection, tioGroupName, 0)
}
