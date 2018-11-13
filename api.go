package stcpbusapi

import (
    "io"
    "fmt"
    "net/http"

    "golang.org/x/net/html"
    "golang.org/x/net/html/atom"
    "github.com/yhat/scrape"
)

type handler struct{}

// The http.Handler for this API endpoint. Responses
// are in JSON, arguments are taken from the request
// path.
var Handler handler

const (
    errSTCPOffline = -1 - iota
    errNoParse
    errNoBuses
)

var errHandleFuns = [...]func(http.ResponseWriter){
    func(w http.ResponseWriter) {fmt.Fprintf(w, `{"erro":"O API da STCP esta offline."}`)},
    func(w http.ResponseWriter) {fmt.Fprintf(w, `{"erro":"O API respondeu com HTML invalido."}`)},
    func(w http.ResponseWriter) {fmt.Fprintf(w, `{"carros":[]}`)},
}

func (handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // no bus stop code in path
    if r.URL.Path == "/" {
        fmt.Fprintf(w, `{"erro":"Nenhum codigo de paragem encontrado no caminho."}`)
        return
    }

    // trim outer slashes
    stop := r.URL.Path[1:]
    n := len(stop) - 1

    if stop[n] == '/' {
        stop = stop[:n]
    }

    // download bus info
    buses, errno := getBuses(stop)

    // error handling
    if errno < 0 {
        errHandleFuns[-errno - 1](w)
        return
    }

    // skip table headers
    buses = buses[1:]
    n = len(buses) - 1

    // write beginning
    fmt.Fprintf(w, `{"carros":[`)

    for i := 0; i < n; i++ {
        fmtBus(w, true, buses[i])
    }

    // write end
    fmtBus(w, false, buses[n])
    fmt.Fprintf(w, `]}`)
}

func fmtBus(w http.ResponseWriter, comma bool, bus *html.Node) {
    // print car
    td := bus.FirstChild.NextSibling
    fmt.Fprintf(w, `{"carro":"%s",`, scrape.Text(td))

    // print time
    td = td.NextSibling.NextSibling
    fmt.Fprintf(w, `"tempo":"%s",`, scrape.Text(td))

    // print await time
    td = td.NextSibling.NextSibling
    fmt.Fprintf(w, `"espera":"%s"}`, scrape.Text(td))

    if comma {
        fmt.Fprintf(w, ",")
    }
}

func getBuses(code string) ([]*html.Node, int) {
    body, errno := downloadHTML(code)
    if errno != 0 {
        return nil, errno
    }
    defer body.Close()

    root, err := html.Parse(body)
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

func downloadHTML(code string) (io.ReadCloser, int) {
    rsp, err := http.Get("https://www.stcp.pt/pt/itinerarium/soapclient.php?codigo="+code)
    if err != nil {
        return nil, errSTCPOffline
    }
    return rsp.Body, 0
}

func matchRow(n *html.Node) bool {
    return n.DataAtom == atom.Tr
}
