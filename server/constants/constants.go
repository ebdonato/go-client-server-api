package constants

import (
	"time"
)

const EXCHANGE_URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
const EXCHANGE_REQUEST_TIMEOUT = 200 * time.Millisecond
const EXCHANGE_PERSIST_TIMEOUT = 10 * time.Millisecond
