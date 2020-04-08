package intelimarketclient

import "testing"

func TestP9Connection(*testing.T) {

	var connection = P9mdi_connect("demo.intelitrader.com.br", 2605)

	P9mdi_disconnect(connection)
}