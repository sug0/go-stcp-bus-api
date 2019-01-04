package stcpbusapi

import (
    "fmt"
    "bytes"

    "golang.org/x/net/html"
    "golang.org/x/net/html/atom"
    "github.com/yhat/scrape"
    "github.com/valyala/fasthttp"
    "github.com/valyala/bytebufferpool"
)

// The http.Handler for this API endpoint. Responses
// are in JSON, arguments are taken from the request
// path.
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

func init() {
    Handler = stcpHandler
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

    // write beginning
    ctx.WriteString(`{"carros":[`)

    for i := 0; i < n; i++ {
        fmtBus(ctx, true, buses[i])
    }

    // write end
    fmtBus(ctx, false, buses[n])
    ctx.WriteString(`]}`)
}

func fmtBus(ctx *fasthttp.RequestCtx, comma bool, bus *html.Node) {
    // print car
    td := bus.FirstChild.NextSibling
    fmt.Fprintf(ctx, `{"carro":"%s",`, scrape.Text(td))

    // print time
    td = td.NextSibling.NextSibling
    fmt.Fprintf(ctx, `"tempo":"%s",`, scrape.Text(td))

    // print await time
    td = td.NextSibling.NextSibling
    fmt.Fprintf(ctx, `"espera":"%s"}`, scrape.Text(td))

    if comma {
        ctx.WriteString(",")
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
