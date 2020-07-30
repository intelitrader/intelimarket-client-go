package main

import (
	"log"
    "flag"
    "strings"
    "time"

	"bitbucket.org/intelitrader/intelimarket-client-go/pkg/intelimarketclient"
)

func PropertyToStdOut(channel <-chan intelimarketclient.PropertyChangeInfo) {

	log.Println("PropertyToStdOut, reading", channel)

	for {
		info := <-channel
		log.Println("Property change:", info)
	}
}

func TradeToStdOut(channel <-chan intelimarketclient.TradeChangeInfo) {

	log.Println("TradeToStdOut, reading", channel)

	for {
		info := <-channel
		log.Println("Trade change:", info)
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
    for {
        connection := intelimarketclient.InteliMarketConnection{}
        hostnamePtr := flag.String("hostname", "demo.intelitrader.com.br", "where to connect")
        portPtr := flag.Int("port", 2605, "port to connect")
        snapshotSizePtr := flag.String("snapshot-size", "0", "how many past events")
        timeoutPtr := flag.Int("timeout", 10, "timeout in seconds")
        var groups arrayFlags
        flag.Var(&groups, "group", "group(s)")
        var instruments arrayFlags
        flag.Var(&instruments, "instrument", "instrument(s)")

        flag.Parse()

        hostname := *hostnamePtr
        port := uint16(*portPtr)
        snapshotSize := *snapshotSizePtr
        timeout := *timeoutPtr

        log.Printf("Connecting to %s:%d\n", hostname, port)
        err := connection.Connect(hostname, port)

        for err != nil {
            log.Println("Connecting...", err)
            err = connection.Connect(hostname, port)
        }

        log.Println("Connected!")

        go PropertyToStdOut(connection.GetPropertyChangeChannel())
        go TradeToStdOut(connection.GetTradeChangeChannel())

        for _, symbol := range groups {
            log.Printf("Subscribing to group %s\n", symbol)
            connection.SubscribeGroupTrades(symbol, "0")
        }

        for _, symbol := range instruments {
            log.Printf("Subscribing to %s\n", symbol)
            connection.SubscribeInstrumentTrades(symbol, snapshotSize)
        }

        for {
            result := connection.DispatchPendingMessage(timeout)
            if result == -2 {
                connection.Disconnect()
                log.Printf("Network error. Waiting for %d seconds to reconnect.\n", timeout)
                time.Sleep(time.Duration(timeout) * time.Second)
                break
            }
        }
    }
}
