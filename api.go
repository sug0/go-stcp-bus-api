package stcpbusapi

import (
    "fmt"
    "bytes"

    "golang.org/x/net/html"
    "golang.org/x/net/html/atom"
    "github.com/yhat/scrape"
    "github.com/valyala/fasthttp"
    "github.com/valyala/bytebufferpool"
    "github.com/coocood/freecache"
)

// The fasthttp.RequestHandler for this API endpoint. Responses
// are in JSON, arguments are taken from the request
// path. Example: http://localhost:8080/BCM1
var Handler fasthttp.RequestHandler

const (
    errSTCPOffline = -1 - iota
    errNoParse
    errNoBuses
)

var errHandleFuns = [...]func(*fasthttp.RequestCtx){
    func(ctx *fasthttp.RequestCtx) {ctx.WriteString(`{"erro":"O API da STCP est\u00e1 offline."}`)},
    func(ctx *fasthttp.RequestCtx) {ctx.WriteString(`{"erro":"O API respondeu com HTML inv\u00e1lido."}`)},
    func(ctx *fasthttp.RequestCtx) {ctx.WriteString(`{"carros":[]}`)},
}

var cache *freecache.Cache

// 25 seconds cache timeout
const cacheTimeout = 25

func init() {
    Handler = stcpHandler
    cache = freecache.NewCache(1 << 14)
}

func stcpHandler(ctx *fasthttp.RequestCtx) {
    stop := ctx.Path()

    // no bus stop code in path
    if bytes.Equal(stop, []byte("/")) {
        ctx.WriteString(`{"erro":"Nenhum c\u00f3digo de paragem encontrado no caminho."}`)
        return
    }

    // trim outer slashes
    stop = stop[1:]
    n := len(stop) - 1

    if stop[n] == '/' {
        stop = stop[:n]
    }

    // try to fetch from cache
    rsp, err := cache.Get(stop)
    if err == nil {
        // cache hit!
        ctx.Write(rsp)
        return
    }

    // download bus info
    buses, errno := getBuses(stop)

    // error handling
    if errno < 0 {
        errHandleFuns[-errno - 1](ctx)
        return
    }

    // skip table headers
    buses = buses[1:]
    n = len(buses) - 1

    var buf bytes.Buffer

    // write beginning
    buf.WriteString(`{"carros":[`)

    for i := 0; i < n; i++ {
        fmtBus(&buf, true, buses[i])
    }

    // write end
    fmtBus(&buf, false, buses[n])
    buf.WriteString(`]}`)

    // save to cache, and write response
    rsp = buf.Bytes()
    cache.Set(stop, rsp, cacheTimeout)
    ctx.Write(rsp)
}

func fmtBus(buf *bytes.Buffer, comma bool, bus *html.Node) {
    // print car
    td := bus.FirstChild.NextSibling
    fmt.Fprintf(buf, `{"carro":"%s",`, scrape.Text(td))

    // print time
    td = td.NextSibling.NextSibling
    fmt.Fprintf(buf, `"hora":"%s",`, scrape.Text(td))

    // print await time
    td = td.NextSibling.NextSibling
    fmt.Fprintf(buf, `"espera":"%s"}`, scrape.Text(td))

    if comma {
        buf.WriteString(",")
    }
}

func getBuses(code []byte) ([]*html.Node, int) {
    buf, body, errno := downloadHTML(code)
    if errno != 0 {
        return nil, errno
    }
    defer bytebufferpool.Put(buf)

    root, err := html.Parse(bytes.NewReader(body))
    if err != nil {
        return nil, errNoParse
    }

    // parse rows
    rows := scrape.FindAll(root, matchRow)

    // no rows found -- no buses in the next 60 min :(
    if len(rows) == 0 {
        return nil, errNoBuses
    }

    return rows, 0
}

func downloadHTML(code []byte) (*bytebufferpool.ByteBuffer, []byte, int) {
    buf := bytebufferpool.Get()
    path := "https://www.stcp.pt/pt/itinerarium/soapclient.php?codigo="+string(code)
    _, rsp, err := fasthttp.Get(buf.B, path)
    if err != nil {
        return nil, nil, errSTCPOffline
    }
    return buf, rsp, 0
}

func matchRow(n *html.Node) bool {
    return n.DataAtom == atom.Tr
}
