package main

import (
	"log"

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

func main() {
	connection := intelimarketclient.InteliMarketConnection{}
	hostname := "demo.intelitrader.com.br"

	log.Println("Connecting to", hostname)
	err := connection.Connect(hostname, 2605)
	defer connection.Disconnect()

	if err != nil {
		log.Println("error: ", err)
		return
	}

	log.Println("Connected!")

	instruments := []string{
		"WINM19",
		"RAIL3", "BTOW3", "AZUL4", "BRFS3", "VVAR3", "BRKM5", "VALE3", "ECOR3", "CYRE3", "ABEV3", "MRVE3", "PETR3", "EMBR3", "MULT3", "TIMP3", "LAME4", "BBSE3", "NATU3",
		"ITSA4", "GGBR4", "FLRY3", "GOAU4", "BBAS3", "IRBR3", "WEGE3", "B3SA3", "RADL3", "PCAR4", "CIEL3", "KROT3", "SANB11", "HYPE3", "MGLU3", "ENBR3", "UGPA3", "SUZB3",
		"IGTA3", "SMLS3", "USIM5", "BRDT3", "PETR4", "ESTC3", "SBSP3", "ELET3", "ELET6", "QUAL3", "EQTL3", "BRAP4", "JBSS3", "VIVT4", "ITUB4", "LREN3", "BRML3", "MRFG3",
		"CVCB3", "BBDC4", "CSAN3", "GOLL4", "CSNA3", "BBDC3", "TAEE11", "RENT3", "CMIG4", "EGIE3", "KLBN11", "CCRO3", "FBOK34", "CLGN34", "AVON34", "NFLX34", "AMZO34",
		"GOGL35", "GOGL34", "BERK34", "DISB34", "SLBG34", "MMMC34", "CHVX34", "UTEC34", "LBRN34", "HSHY34", "HONB34", "JNJB34", "UPSS34", "UPAC34", "GSGI34", "COTY34",
		"PEPB34", "BOAC34", "LMTB34", "EBAY34", "MCDC34", "HALI34", "NIKE34", "QCOM34", "MOSC34", "KMBB34", "KHCB34", "FDXB34", "GEOO34", "HPQB34", "HOME34", "MRCK34",
		"AIGB34", "WUNI34", "COCA34", "TIFF34", "XRXB34", "BOXP34", "DHER34", "MDLZ34", "USBC34", "CMCS34", "JPMC34", "CSCO34", "BMYB34", "AXPB34", "ATTB34", "VERZ34",
		"MSCD34", "ACNB34", "ORCL34", "GDBR34", "FCXO34", "ABTT34", "PGCO34", "COLG34", "CATP34", "FDMO34", "BONY34", "MSBR34", "ARNC34", "FMXB34", "DWDP34", "TEXA34",
		"CTGP34", "ITLC34", "METB34", "SBUB34", "WFCO34", "BOEI34", "PFIZ34", "WALM34", "IBMB34", "AAPL34", "EXXO34", "COPH34", "TGTB34", "MSFT34", "DUKB34", "ARMT34",
		"LILY34", "VISA34", "AMGN34", "RAIL3F", "BTOW3F", "AZUL4F", "BRFS3F", "VVAR3F", "BRKM5F", "VALE3F", "ECOR3F", "CYRE3F", "ABEV3F", "MRVE3F", "PETR3F", "EMBR3F",
		"MULT3F", "TIMP3F", "LAME4F", "BBSE3F", "NATU3F", "ITSA4F", "GGBR4F", "FLRY3F", "GOAU4F", "BBAS3F", "IRBR3F", "WEGE3F", "B3SA3F", "RADL3F", "PCAR4F", "CIEL3F",
		"KROT3F", "SANB11F", "HYPE3F", "MGLU3F", "ENBR3F", "UGPA3F", "SUZB3F", "IGTA3F", "SMLS3F", "USIM5F", "BRDT3F", "PETR4F", "ESTC3F", "SBSP3F", "ELET3F", "ELET6F",
		"QUAL3F", "EQTL3F", "BRAP4F", "JBSS3F", "VIVT4F", "ITUB4F", "LREN3F", "BRML3F", "MRFG3F", "CVCB3F", "BBDC4F", "CSAN3F", "GOLL4F", "CSNA3F", "BBDC3F", "TAEE11F",
		"RENT3F", "CMIG4F", "EGIE3F", "KLBN11F", "CCRO3F", "FBOK34F", "CLGN34F", "AVON34F", "NFLX34F", "AMZO34F", "GOGL35F", "GOGL34F", "BERK34F", "DISB34F", "SLBG34F",
		"MMMC34F", "CHVX34F", "UTEC34F", "LBRN34F", "HSHY34F", "HONB34F", "JNJB34F", "UPSS34F", "UPAC34F", "GSGI34F", "COTY34F", "PEPB34F", "BOAC34F", "LMTB34F",
		"EBAY34F", "MCDC34F", "HALI34F", "NIKE34F", "QCOM34F", "MOSC34F", "KMBB34F", "KHCB34F", "FDXB34F", "GEOO34F", "HPQB34F", "HOME34F", "MRCK34F", "AIGB34F",
		"WUNI34F", "COCA34F", "TIFF34F", "XRXB34F", "BOXP34F", "DHER34F", "MDLZ34F", "USBC34F", "CMCS34F", "JPMC34F", "CSCO34F", "BMYB34F", "AXPB34F", "ATTB34F",
		"VERZ34F", "MSCD34F", "ACNB34F", "ORCL34F", "GDBR34F", "FCXO34F", "ABTT34F", "PGCO34F", "COLG34F", "CATP34F", "FDMO34F", "BONY34F", "MSBR34F", "ARNC34F",
		"FMXB34F", "DWDP34F", "TEXA34F", "CTGP34F", "ITLC34F", "METB34F", "SBUB34F", "WFCO34F", "BOEI34F", "PFIZ34F", "WALM34F", "IBMB34F", "AAPL34F", "EXXO34F",
		"COPH34F", "TGTB34F", "MSFT34F", "DUKB34F", "ARMT34F", "LILY34F", "VISA34F", "AMGN34F"}

	groups := []string{"CS", "PS", "FUT"}

	groupCount := 1
	instrumentCount := 0

	go PropertyToStdOut(connection.GetPropertyChangeChannel())
	go TradeToStdOut(connection.GetTradeChangeChannel())
	//statsPrinter := NewInstrumentsEventsStatsPrinter(10)
	//statsPrinter.LaunchAsyncStatisticsPrinter(connection.GetPropertyChangeChannel())

	for i, groupName := range groups[:groupCount] {
		log.Printf("Subscribing to group %v (%v/%v)\n", groupName, i+1, groupCount)
		connection.SubscribeGroupProperties(groupName)
	}

	for i, symbol := range instruments[:instrumentCount] {
		log.Printf("Subscribing to %v (%v/%v)\n", symbol, i+1, instrumentCount)
		connection.SubscribeInstrumentProperties(symbol)
	}

	for {
		connection.DispatchPendingMessage()
	}
}
