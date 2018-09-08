package main

import (
    "net/http"

    "github.com/sugoiuguu/go-stcp-bus-api"
)

func main() {
    // * Requests are made to http://localhost:8080/<bus stop code>
    // * Bus stop codes can be found here http://www.stcp.pt/pt/viajar/linhas/
    panic(http.ListenAndServe(":8080", stcpbusapi.Handler))
}
