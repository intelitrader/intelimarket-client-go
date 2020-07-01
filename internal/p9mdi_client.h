#pragma once

#ifndef BOOL
#define BOOL int
#endif

#ifndef TRUE
#define TRUE 1
#endif

#ifndef FALSE
#define FALSE 0
#endif

#define _WINSOCK_DEPRECATED_NO_WARNINGS

#ifdef _MSC_VER
#define _CRT_SECURE_NO_WARNINGS
#include <stdio.h>
#include <tchar.h>
#include <assert.h>
#include <WinSock2.h>
#pragma comment(lib ,"ws2_32.lib")
#endif


#ifdef __cplusplus
extern "C" {
#endif

enum ConnectionStatusEnum
{
	ConnectionStatusEnum_Connected,
	ConnectionStatusEnum_Disconnected
};

enum SubscriptionType
{
	SnapshotOnly,
	SnapshotPlusIncremental,
	IncrementalOnly
};

#define BOOK_SIDE_BUY		1
#define BOOK_SIDE_SELL		2

#define FULL_ORDER_BOOK     0xFFFFFFFF
#define ALL_TRADES          0xFFFFFFFF

#define P9MDI_INFO_TYPE_UNKNOWN				0
#define P9MDI_INFO_TYPE_PROPERTIES			1
#define P9MDI_INFO_TYPE_BOOK_BUY			2
#define P9MDI_INFO_TYPE_BOOK_SELL			3
#define P9MDI_INFO_TYPE_TRADES				4
#define P9MDI_INFO_TYPE_SECURITY_LIST		5
#define P9MDI_INFO_TYPE_INSTRUMENT_STATS	6
#define P9MDI_INFO_TYPE_STATS_BUYERS		7
#define P9MDI_INFO_TYPE_STATS_SELLERS		8

struct KEY_AND_VALUE
{
	const char* key;
	const char* value;	
};

#define P9MDI_EVENT_SET					0x14
#define P9MDI_EVENT_INSERT				0x15
#define P9MDI_EVENT_DELETE				0x16
#define P9MDI_EVENT_PUSH_BACK			0x17
#define P9MDI_EVENT_PUSH_FRONT			0x18
#define P9MDI_EVENT_POP_BACK			0x19
#define P9MDI_EVENT_POP_FRONT			0x1A
#define P9MDI_EVENT_CLEAR				0x1B

#define P9MDI_EVENT_SNAPSHOT_END		0x23

#define P9MDI_EVENT_SUBSCRIPTION		0x66


#define P9MDI_DEBUG_FLAG_DUMP_MESSAGES_TO_STDOUT 0x01

#define P9MDI_ERROR_NO_INTELIMARKET_ON_THIS_TIO 0xFF01

#define P9_FAILED(x) (x<0)

struct P9MDI_CONNECTION;

typedef void (*FieldChangeCallback)(int /*error_code*/, void* /*handle*/, void* /*cookie*/, unsigned /*eventCode*/, 
									const char* /*exchange*/, const char* /*symbol*/, 
									const char* /*key*/, const char* /*value*/ );

typedef void (*ListChangeCallback)(int /*error_code*/, void* /*handle*/, void* /*cookie*/, unsigned /*eventCode*/, 
								   const char* /*exchange*/, const char* /*symbol*/, 
								   unsigned /*position*/, struct KEY_AND_VALUE* /*fields*/);

typedef void (*MapChangeCallback)(int /*error_code*/, void* /*handle*/, void* /*cookie*/, unsigned /*eventCode*/, 
								  const char* /*exchange*/, const char* /*symbol*/, 
								  const char* /*key*/, struct KEY_AND_VALUE* /*fields*/);


typedef void(*NetChangeCallback)(int /*error_code*/, void* /*handle*/, void* /*cookie*/, unsigned /*eventCode*/,
								const char* /*exchange*/, const char* /*symbol*/, const char* /*net_change*/);

typedef void (*GroupPropertyChangeCallback)(int /*result*/, void* /*handle*/, void* /*cookie*/, 
											const char* /*groupName*/, 
											const char* /*symbol*/, struct KEY_AND_VALUE* /*fields*/);

typedef void(*DisconnectionCallback)(struct P9MDI_CONNECTION*);



const char* p9mdi_get_field_value(struct KEY_AND_VALUE* field_array, const char* key);
const char* p9mdi_get_event_code_name(unsigned int event_code);

int p9mdi_connect(const char* server, unsigned short port, const char* username, const char* password, const char* log_path, struct P9MDI_CONNECTION** connection);
int	p9mdi_disconnect(struct P9MDI_CONNECTION* connection);

void p9mdi_set_debug_flags(int flags);

// use it if you need the socket fd to async io (select)
int p9mdi_get_connection_fd(struct P9MDI_CONNECTION* connection);

int p9mdi_subscribe_security_list(
	struct P9MDI_CONNECTION* connection,
	const char* exchange,	
	enum SubscriptionType subscriptionType,
	MapChangeCallback map_change_callback,
	void* cookie);

int p9mdi_subscribe_instruments_stats(
	struct P9MDI_CONNECTION* connection,
	const char* exchange,
	MapChangeCallback stats_callback,
	void* cookie);

int p9mdi_subscribe_broker_stats(
	struct P9MDI_CONNECTION* connection,
	const char* exchange,
	const char* symbol,
	MapChangeCallback broker_stats_callback,
	void* cookie);

int p9mdi_subscribe_instrument_order_book(
	struct P9MDI_CONNECTION* connection,
	const char* exchange,
	const char* symbol,
	unsigned int side,
	unsigned int orderBookSize,
	ListChangeCallback list_change_callback,
	void* cookie);

int p9mdi_subscribe_instrument_trades(
	struct P9MDI_CONNECTION* connection,
	const char* exchange,
	const char* symbol,
	unsigned snapshotSize,
	ListChangeCallback list_change_callback,
	void* cookie);

int p9mdi_subscribe_instrument_properties(
	struct P9MDI_CONNECTION* connection, 
	const char* exchange,
	const char* symbol,
	enum SubscriptionType subscriptionType,
	FieldChangeCallback field_change_callback,
	void* cookie);

int p9mdi_query_instrument_properties(
	struct P9MDI_CONNECTION* connection, 
	const char* exchange,
	const char* symbol,
	const char** fields,
	unsigned int fieldCount,
	MapChangeCallback map_change_callback,
	void* cookie);

int p9mdi_set_subscribe_group_callback(
	struct P9MDI_CONNECTION* connection, 
	FieldChangeCallback properties_callback,
	ListChangeCallback book_buy_callback,
	ListChangeCallback book_sell_callback,
	ListChangeCallback trades_callback,
	void* cookie);

int p9mdi_subscribe_group(
	struct P9MDI_CONNECTION* connection, 
	const char* groupName,
	int start);

int p9mdi_unsubscribe(struct P9MDI_CONNECTION* connection, void* subscription_handle);

int p9mdi_dispatch_pending_events(struct P9MDI_CONNECTION* connection, BOOL receive);

int p9mdi_set_asynchronous(struct P9MDI_CONNECTION* connection, DisconnectionCallback disconnect_callback);

void p9mdi_log(const char* what);

void p9mdi_enable_message_dump_to_log(BOOL yes_or_no);

int p9mdi_test();

struct KEY_AND_VALUE* p9mdi_get_next_key_value_field(struct KEY_AND_VALUE* fields);

#ifdef __cplusplus
} //extern "C"
#endif
