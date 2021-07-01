package intelimarketclient

/*

//
// Isso precisa ficar em um arquivo separado porque os arquivos Go
// funcionam +- como um include. Qualquer arquivo Go que chamar
// esse vai incluir a definição das funções e vai dar duplicate symbol
//

#include <stdlib.h>
#include "p9mdi_client.h"


//
// Forward declaration das funções implementadas em Go
//
void fieldChangeCallback_Go(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, char* key, char* value);
void tradeCallback_Go(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, unsigned position, struct KEY_AND_VALUE* fields);
void bookEntryCallback_Go(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, unsigned position, struct KEY_AND_VALUE* fields);


//
// Gateway function, C to Go
//

void FieldChangeCallback_cgo(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, char* key, char* value)
{
	fieldChangeCallback_Go(error_code, handle, cookie, eventCode, exchange, symbol, key, value);
}

void TradeCallback_cgo(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, unsigned position, struct KEY_AND_VALUE* fields)
{
	tradeCallback_Go(error_code, handle, cookie, eventCode, exchange, symbol, position, fields);
}

void BookEntryCallback_cgo(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, unsigned position, struct KEY_AND_VALUE* fields)
{
	bookEntryCallback_Go(error_code, handle, cookie, eventCode, exchange, symbol, position, fields);
}


#cgo CFLAGS: -I../../inc
*/
import "C"
