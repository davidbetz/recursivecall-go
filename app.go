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
        contents, err := ioutil.ReadAll(response.Body)
        if response.StatusCode == 200 {
            ch <- fmt.Sprintf("%s", contents)
        } else {
            errChannel <- err
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
    select {
        case value := <-ch:
            value += "," + id
            log.Println("-RETURNING " + value)
            w.Write([]byte(value))
        case <-errChannel:
            w.WriteHeader(http.StatusInternalServerError)
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
