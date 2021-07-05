package main

import (
	"log"
    "flag"
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
    "time"

	"bitbucket.org/intelitrader/intelimarket-client-go/pkg/intelimarketclient"
)

var g_verbose bool = false
var g_tradeCount int = 0
var g_propertyCount int = 0
var g_bookCount int = 0
var g_lastTick time.Time = time.Now()

func PrintLogTrace(line string) {
    intelimarketclient.LogTrace(line)
    log.Printf("%s\n", line)
}

func LogStats() {
    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)
    var mb = mem.HeapAlloc / 1024 / 1024
    line := fmt.Sprintf("EVENTS: properties %d, trades %d, book %d, memory %dMB", g_propertyCount, g_tradeCount, g_bookCount, mb)
    PrintLogTrace(line)
}

func Tick() {
    t0 := time.Now()
    if t0.Sub(g_lastTick).Seconds() > 2 {
        g_lastTick = t0
        LogStats()
    }
}

func PropertyToStdOut(channel <-chan intelimarketclient.PropertyChangeInfo) {

    PrintLogTrace(fmt.Sprintf("PropertyToStdOut, reading %s", channel))

	for {
		info := <-channel
        line := fmt.Sprintf("Property change: %s", info)
        if g_verbose {
            log.Println(line)
        }
        intelimarketclient.LogTrace(line)
        g_propertyCount += 1
        Tick()
	}
}

func TradeToStdOut(channel <-chan intelimarketclient.TradeChangeInfo) {

    PrintLogTrace(fmt.Sprintf("TradeToStdOut, reading %s", channel))

	for {
		info := <-channel
        line := fmt.Sprintf("Trade event: %s", info)
        if g_verbose {
            log.Println(line)
        }
        intelimarketclient.LogTrace(line)
        g_tradeCount += 1
        Tick()
	}
}

func BookToStdOut(channel <-chan intelimarketclient.BookChangeInfo) {

    PrintLogTrace(fmt.Sprintf("BookToStdOut, reading %s", channel))

    for {
        info := <-channel
        line := fmt.Sprintf("Book event: %s", info)
        if g_verbose {
            log.Println(line)
        }
        intelimarketclient.LogTrace(line)
        g_bookCount += 1
        Tick()
    }
}

type arrayFlags []string

func (i *arrayFlags) String() string {
    return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
    *i = append(*i, value)
    return nil
}

func main() {

    serverPtr := flag.String("server", "demo.intelitrader.com.br", "where to connect")
    portPtr := flag.Int("port", 2605, "port to connect")
    logpathPtr := flag.String("log-path", ".", "path to log")
    snapshotSizePtr := flag.String("snapshot-size", "0", "how many past events")
    timeoutPtr := flag.Int("timeout", 10, "timeout in seconds")
    verbosePtr := flag.Bool("verbose", false, "verbose output")
    dispatchLoopPtr := flag.Int("dispatch-loop", 0, "how many dispatch loops before exit")
    subscribePropertiesPtr := flag.Bool("subscribe-properties", false, "subscribe properties")
    subscribeTradesPtr := flag.Bool("subscribe-trades", false, "subscribe trades")
    subscribeBookPtr := flag.Bool("subscribe-book", false, "subscribe book")
    subscriptionTypePtr := flag.Int("subscription-type", 1, "subscription type when subscribing to book\noptions: Snapshot(0), Snapshot plus Incremental (1), Incremental (2)")
    chansizePtr := flag.Int("chan-size", 1024, "how many allocated channels to each event channel")

    var groups arrayFlags
    flag.Var(&groups, "group", "group(s)")
    var instruments arrayFlags
    flag.Var(&instruments, "instrument", "instrument(s)")

    flag.Parse()

    server := *serverPtr
    port := uint16(*portPtr)
    logpath := *logpathPtr
    snapshotSize := *snapshotSizePtr
    timeout := *timeoutPtr
    dispatchLoop := *dispatchLoopPtr
    subscribeProperties := *subscribePropertiesPtr
    subscribeTrades := *subscribeTradesPtr
    subscribeBook := *subscribeBookPtr
    subscriptionType := strconv.Itoa(*subscriptionTypePtr)
    chansize := *chansizePtr

    g_verbose = *verbosePtr

    connection := &intelimarketclient.InteliMarketConnection{}
    for {

        for {
            log.Printf("Connecting to %s:%d\n", server, port)
            err := connection.Connect(server, port, logpath, chansize)
            if err == nil {
                break
            }
            PrintLogTrace(fmt.Sprintf("Connection error; waiting for %d seconds to try again.", timeout))
            time.Sleep(time.Duration(timeout) * time.Second)
        }

        var connected = true
        log.Println("Connected!")
        PrintLogTrace(strings.Join(os.Args, " "))
        LogStats()

        if subscribeTrades {
            go TradeToStdOut(connection.GetTradeChangeChannel())
            for _, symbol := range instruments {
                PrintLogTrace(fmt.Sprintf("Subscribing trades to %s", symbol))
                connection.SubscribeInstrumentTrades(symbol, snapshotSize)
            }
            for _, symbol := range groups {
                PrintLogTrace(fmt.Sprintf("Subscribing trades to group %s", symbol))
                connection.SubscribeGroupTrades(symbol, snapshotSize)
            }
        }

        if subscribeProperties {
            go PropertyToStdOut(connection.GetPropertyChangeChannel())
            for _, symbol := range instruments {
                PrintLogTrace(fmt.Sprintf("Subscribing properties to %s", symbol))
                    connection.SubscribeInstrumentProperties(symbol)
            }
            for _, symbol := range groups {
                PrintLogTrace(fmt.Sprintf("Subscribing properties to group %s", symbol))
                    connection.SubscribeGroupProperties(symbol)
            }
        }
		
		if subscribeBook {
            go BookToStdOut(connection.GetBookChangeChannel())
            for _, symbol := range instruments {
                PrintLogTrace(fmt.Sprintf("Subscribing book to %s", symbol))
                connection.SubscribeInstrumentOrderBook(symbol, subscriptionType)
            }
            for _, symbol := range groups {
                PrintLogTrace(fmt.Sprintf("Subscribing book to group %s", symbol))
                connection.SubscribeGroupBook(symbol)
            }
        }


        PrintLogTrace("Dispatching pending messages")
        for dloop := 0; dispatchLoop == 0 || dloop < dispatchLoop; dloop++ {
            result, internalError := connection.DispatchPendingMessage(timeout)
            if g_verbose {
                PrintLogTrace(fmt.Sprintf("Dispatch #%d result %d (internal error %d)", dloop, result, internalError))
                LogStats()
            }
            if result == -2 {
                PrintLogTrace(fmt.Sprintf("NETWORK ERROR (internal error %d); reconnecting", internalError))
                connection.Disconnect()
                connected = false
                break
            }
        }

        if connected {
            PrintLogTrace("Dispatch loop done and still connected; disconnecting")
            connection.Disconnect()
        }
    }
}
