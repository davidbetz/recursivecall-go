// MIT License

// Copyright (c) 2016-2018 David Betz

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "runtime"
    "strings"
    "strconv"
    "time"
)

var port = os.Getenv("PORT")
var procsString = os.Getenv("PROCS")
var addressSuffix = os.Getenv("ADDRESS_SUFFIX")

func call(address string, outputChannel chan<- string, errChannel chan<- error) {
    var uri = "http://" + address + "." + addressSuffix + ":" + port
    log.Println("+CALLING ", uri)
    response, err := http.Get(uri)
    if err != nil {
        errChannel <- errors.New("500")
    } else {
        contents, err := ioutil.ReadAll(response.Body)
        if response.StatusCode == 200 {
            outputChannel <- fmt.Sprintf("%s", contents)
        } else {
            errChannel <- err
        }
        defer response.Body.Close()
    }
}

func getId(host string) string {
    log.Println("host", host)
    var id string
    //+ if . in host, use first part
    withSuffixParts := strings.Split(host, ".")
    if(len(withSuffixParts) > 0) {
        id = withSuffixParts[0]
    }
    //+ if : in host, use host
    if(len(id) == 0) {
        parts := strings.Split(host, ":")
        id = parts[0]
    }
    //+ if len(host) > 0, default to g
    if(len(id) != 1) {
        log.Println("Argument length != 1. Defaulting to g.")
        id = "g"
    }
    return id
}

func handler(w http.ResponseWriter, r *http.Request) {
    var id = getId(r.Host)
    var called = []rune(id)[0]
    outputChannel := make(chan string)
    errChannel := make(chan error, 1)
    log.Println("+CALLED id", id, "|called", called)
    if called > 97 {
        go call(string(called-1), outputChannel, errChannel)
    } else {
        go func() {
            outputChannel <- "$"
        }()
    }
    select {
        case value := <- outputChannel:
            value += "," + id
            log.Println("-RETURNING " + value)
            w.Write([]byte(value))
        case <-errChannel:
            w.WriteHeader(http.StatusInternalServerError)
    }
}

func check(letter string) {
    response, err := http.Get("http://" + letter + "." + addressSuffix + ":" + port)
    if err != nil {
        panic(err)
    } else {
        contents, _ := ioutil.ReadAll(response.Body)
        log.Println(fmt.Sprintf("%s", contents))
    }
}

func waste(procs int) {
    log.Println("wasting CPU cycles for 10 seconds...")
    q := make(chan bool)

    for i := 0; i < procs; i++ {
        go func() {
            for {
                select {
                case <- q:
                    return
                default:
                }
            }
        }()
    }

    time.Sleep(10 * time.Second)
    for i := 0; i < procs; i++ {
        q <- true
    }
    log.Println("done CPU cylcles.")
}

func main() {
    procs, err := strconv.Atoi(procsString)
    if err != nil {
        procs = 1
    }
    log.Println("PROCS", procsString)
    if procs > 0 {
        runtime.GOMAXPROCS(procs)
    }
    var arg string
    if(len(os.Args) > 1) {
        arg = os.Args[1]
    }
    if len(port) == 0 {
        port = "3000"
    }
    log.Println("PORT", port)
    log.Println("ADDRESS_SUFFIX", addressSuffix)
    switch arg {
    case "check":
        check(getId(os.Args[2]))
    case "waste":
        waste(procs)
    default:
        http.HandleFunc("/", handler)
        log.Println("Starting server on port", port)
        log.Fatal(http.ListenAndServe(":" + port, nil))
    }
}
