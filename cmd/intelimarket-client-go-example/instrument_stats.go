package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/intelitrader/intelimarket-client-go/pkg/intelimarketclient"
)

type instrumentStats struct {
	symbol     string
	eventCount int
}

type instrumentsEventsStatsPrinter struct {
	instrumentsEventsStats      map[string]*instrumentStats
	instrumentStatsLogStep      int
	instrumentStatsEventCount   int
	instrumentCountOnLogMessage int
}

func NewInstrumentsEventsStatsPrinter(instrumentCountOnLogMessage int) *instrumentsEventsStatsPrinter {
	return &instrumentsEventsStatsPrinter{make(map[string]*instrumentStats), 10000, 0, instrumentCountOnLogMessage}
}

func (self instrumentStats) String() string {
	return fmt.Sprintf("%v=%v", self.symbol, self.eventCount)
}

func (self instrumentsEventsStatsPrinter) LaunchAsyncStatisticsPrinter(channel <-chan intelimarketclient.PropertyChangeInfo) {
	go self.asyncStatisticsPrinter(channel)
}

func (self instrumentsEventsStatsPrinter) asyncStatisticsPrinter(channel <-chan intelimarketclient.PropertyChangeInfo) {
	for {
		info := <-channel
		if _, ok := self.instrumentsEventsStats[info.Symbol]; !ok {
			self.instrumentsEventsStats[info.Symbol] = &instrumentStats{info.Symbol, 0}
		}

		self.instrumentsEventsStats[info.Symbol].eventCount++
		self.instrumentStatsEventCount++

		if self.instrumentStatsEventCount%self.instrumentStatsLogStep == 0 {
			// saudades do LINQ...
			var biggest []*instrumentStats
			for _, v := range self.instrumentsEventsStats {
				biggest = append(biggest, v)
			}
			sort.Slice(biggest, func(i, j int) bool {
				return biggest[i].eventCount > biggest[j].eventCount
			})

			log.Printf("STATS: eventCount=%v, perInstrument=%v", self.instrumentStatsEventCount, biggest[:self.instrumentCountOnLogMessage])
		}
	}
}
