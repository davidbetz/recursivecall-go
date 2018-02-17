package main

import (
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
)

var port = os.Getenv("PORT")

func call(address string, ch chan<- string, errChannel chan<- error) {
    var uri = "http://" + address + ":" + port
    log.Println("+CALLING ", uri)
    response, err := http.Get(uri)
    if err != nil {
        errChannel <- errors.New("500")
    } else {
        contents, _ := ioutil.ReadAll(response.Body)
        result := fmt.Sprintf("%s", contents)
        if response.StatusCode == 200 {
            ch <- result
        } else {
            errChannel <- errors.New(result)
        }
        defer response.Body.Close()
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    var id = strings.Split(r.Host, ":")[0]
    var called = []rune(id)[0]
    errChannel := make(chan error, 1)
    log.Println("+CALLED id:", id, "|called:", called)
    ch := make(chan string)
    if called > 97 {
        go call(string(called-1), ch, errChannel)
    } else {
        go func() {
            ch <- "$"
        }()
    }
    ret := func(v string) {
        v += "," + id
        log.Println("-RETURNING " + v)
        w.Write([]byte(v))
    }
    select {
    case value := <-ch:
        ret(value)
    case err := <-errChannel:
        w.WriteHeader(500)
        ret(fmt.Sprintf("%s", err))
    }
}

func main() {
    if len(port) == 0 {
        port = "3000"
    }

    http.HandleFunc("/", handler)
    log.Println("Starting server on port", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
