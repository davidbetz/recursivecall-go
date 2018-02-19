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
    if (id == "localhost") {
        id = os.Getenv("NAME")
    }
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

func check(letter string) {
    response, err := http.Get("http://" + letter + ":" + port)
    if err != nil {
        panic(err)
    } else {
        contents, _ := ioutil.ReadAll(response.Body)
        log.Println(fmt.Sprintf("%s", contents))
    }
}

func main() {
    var arg string
    if(len(os.Args) > 1) {
        arg = os.Args[1]
    }
    if len(port) == 0 {
        port = "3000"
    }
    switch arg {
        case "check":
            if(len(os.Args) < 3 || len(os.Args[2]) != 1) {
                panic("letter required")
            }
            check(os.Args[2])
        default:
            http.HandleFunc("/", handler)
            log.Println("Starting server on port", port)
            log.Fatal(http.ListenAndServe(":"+port, nil))
    }
}
