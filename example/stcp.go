package main

import (
    "github.com/valyala/fasthttp"
    "github.com/sugoiuguu/go-stcp-bus-api"
)

func main() {
    // * Requests are made to http://localhost:8080/<bus stop code>
    // * Bus stop codes can be found here http://www.stcp.pt/pt/viajar/linhas/
    panic(fasthttp.ListenAndServe(":8080", stcpbusapi.Handler))
}
