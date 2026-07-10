package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intelitrader/intelimarket-client-go/pkg/intelimarketclient"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type instrumentStat struct {
	receivedTrades int64
	lastTradeCount int64
	hasTradeCount  bool
	maxTradeIdNum  int64
	hasMaxTradeId  bool
}

type state struct {
	totalTrades    int64
	receivedTrades int64
	lastLogTotal   int64
	lastLogTime    time.Time
	instruments    map[string]*instrumentStat
	symbolsOrder   []string
}

func runCollector(ctx context.Context, tradeCh <-chan intelimarketclient.TradeChangeInfo, propCh <-chan intelimarketclient.PropertyChangeInfo, logEvery int64, tolerance int64) {
	s := &state{lastLogTime: time.Now(), instruments: make(map[string]*instrumentStat)}
	for {
		select {
		case info := <-tradeCh:
			stat, ok := s.instruments[info.Symbol]
			if !ok {
				stat = &instrumentStat{}
				s.instruments[info.Symbol] = stat
				s.symbolsOrder = append(s.symbolsOrder, info.Symbol)
			}

			if info.TradeId != "" {
				if tradeIdNum, err := strconv.ParseInt(info.TradeId, 10, 64); err == nil {
					if stat.hasMaxTradeId && tradeIdNum <= stat.maxTradeIdNum {
						continue
					}
					stat.maxTradeIdNum = tradeIdNum
					stat.hasMaxTradeId = true
				}
			}

			s.receivedTrades++
			stat.receivedTrades++
			s.totalTrades++

			now := time.Now()
			delta := s.totalTrades - s.lastLogTotal
			if delta >= logEvery && now.Sub(s.lastLogTime) >= time.Second {
				elapsed := now.Sub(s.lastLogTime).Seconds()
				PrintLog("trades total=%d delta=%d rate=%.2f/s trigger=%s",
					s.totalTrades, delta, float64(delta)/elapsed, info.Symbol)
				for _, sym := range s.symbolsOrder {
					st := s.instruments[sym]
					serverCount := int64(0)
					if st.hasTradeCount {
						serverCount = st.lastTradeCount
					}
					delta := st.receivedTrades - serverCount
					if delta > tolerance {
						PrintLog("  %-10s server=%-12d local=%-12d delta=%-+6d [MISMATCH]",
							sym, serverCount, st.receivedTrades, delta)
					}
				}
				s.lastLogTotal = s.totalTrades
				s.lastLogTime = now
			}

		case info := <-propCh:
			stat, ok := s.instruments[info.Symbol]
			if !ok {
				stat = &instrumentStat{}
				s.instruments[info.Symbol] = stat
				s.symbolsOrder = append(s.symbolsOrder, info.Symbol)
			}
			if info.Key == "TradeCount" {
				if v, err := strconv.ParseInt(info.Value, 10, 64); err == nil {
					stat.lastTradeCount = v
					stat.hasTradeCount = true
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func PrintLog(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	intelimarketclient.LogTrace(line)
	log.Printf("%s\n", line)
}

func main() {
	serverPtr := flag.String("server", "demo.intelitrader.com.br", "server to connect")
	portPtr := flag.Int("port", 2605, "port to connect")
	logpathPtr := flag.String("log-path", ".", "path to log files")
	snapshotSizePtr := flag.String("snapshot-size", "0", "trade snapshot size (0 for incremental only)")
	logEveryPtr := flag.Int("log-every", 1000, "log a summary every N trades")
	tolerancePtr := flag.Int64("tolerance", 10, "ignore mismatch if local-server <= N")

	var groups arrayFlags
	flag.Var(&groups, "group", "instrument group(s) to subscribe (can be used multiple times)")
	var instruments arrayFlags
	flag.Var(&instruments, "instrument", "specific instrument(s) to subscribe (can be used multiple times)")

	flag.Parse()

	if len(groups) == 0 && len(instruments) == 0 {
		PrintLog("ERROR: at least one -group or -instrument must be specified")
		flag.Usage()
		os.Exit(1)
	}

	server := *serverPtr
	port := uint16(*portPtr)
	logpath := *logpathPtr
	snapshotSize := *snapshotSizePtr
	logEvery := int64(*logEveryPtr)
	tolerance := *tolerancePtr

	connection := &intelimarketclient.InteliMarketConnection{}

	PrintLog("Connecting to %s:%d", server, port)
	err := connection.Connect(server, port, logpath, 65536)
	if err != nil {
		PrintLog("FATAL: connection failed: %v", err)
		os.Exit(1)
	}

	connectTime := time.Now()
	PrintLog("Connected!")
	PrintLog("Args: %s", strings.Join(os.Args, " "))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runCollector(ctx, connection.GetTradeChangeChannel(), connection.GetPropertyChangeChannel(), logEvery, tolerance)

	for _, symbol := range instruments {
		PrintLog("Subscribing properties to instrument %s", symbol)
		connection.SubscribeInstrumentProperties(symbol)
		PrintLog("Subscribing trades to instrument %s", symbol)
		connection.SubscribeInstrumentTrades(symbol, snapshotSize)
	}

	for _, group := range groups {
		PrintLog("Subscribing properties to group %s", group)
		connection.SubscribeGroupProperties(group)
		PrintLog("Subscribing trades to group %s", group)
		connection.SubscribeGroupTrades(group, snapshotSize)
	}

	for {
		result, internalError := connection.DispatchPendingMessage(10)
		if result == -2 {
			if internalError == 4 {
				// EINTR is normal, just dispatch again
				continue
			}
			PrintLog("NETWORK ERROR (internal error %d); reconnecting", internalError)
			connection.Disconnect()
			break
		}
	}

	connectedDuration := time.Since(connectTime)
	cancel()

	PrintLog("FATAL: connection lost")
	PrintLog("  connectedFor:  %s", connectedDuration)
	os.Exit(1)
}
