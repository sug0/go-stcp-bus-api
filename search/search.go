package search

import (
    "io/ioutil"
    "net/url"
    "net/http"
    "encoding/json"
    "unsafe"
)

type BusStop struct {
    Code     string   `json:"code"`
    Name     string   `json:"name"`
    Zone     string   `json:"zone"`
    Location Location `json:"geomdesc"`
    Lines    []Line   `json:"lines"`
}

type Line struct {
    Code        string `json:"code"`
    Description string `json:"description"`
}

type Location struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}

type geomdesc struct {
    Coordinates [2]float64 `json:"coordinates"`
}

// Looks up a list of bus stops given a query string.
func BusStops(q string) ([]BusStop, error) {
    const (
        base = "https://www.stcp.pt/pt/itinerarium/callservice.php?action=srchstoplines&stopname="
    )
    rsp, err := http.Get(base + url.QueryEscape(q))
    if err != nil {
        return nil, err
    }
    defer rsp.Body.Close()
    data, err := ioutil.ReadAll(rsp.Body)
    if err != nil {
        return nil, err
    }
    var stops []BusStop
    err = json.Unmarshal(data, &stops)
    if err != nil {
        return nil, err
    }
    return stops, nil
}

func (l *Location) UnmarshalJSON(data []byte) error {
    var b []byte
    err := json.Unmarshal(data, (*string)(unsafe.Pointer(&b)))
    if err != nil {
        return err
    }
    var g geomdesc
    err = json.Unmarshal(b[:len(b):len(b)], &g)
    if err != nil {
        return err
    }
    l.Lat = g.Coordinates[0]
    l.Lng = g.Coordinates[1]
    return nil
}

func (l *Location) MarshalJSON() ([]byte, error) {
    g := geomdesc{
        Coordinates: [2]float64{l.Lat, l.Lng},
    }
    b, err := json.Marshal(g)
    if err != nil {
        panic(err)
    }
    s := *(*string)(unsafe.Pointer(&b))
    return json.Marshal(s)
}
