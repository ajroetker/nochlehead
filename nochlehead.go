package main

import (
    "net/http"
    "os"
    "log"
    "flag"
)

var (
    logFile *string = flag.String("log", "", "File to log to")
)

func main() {
    flag.Parse()
    if *logFile != "" {
        // Open the logfile
        logfile, err := os.OpenFile( *logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
        if err != nil {
            log.Fatalf("error opening file: %v", err)
        }
        defer logfile.Close()
        // From here on out log to the logfile
        log.SetOutput(logfile)
    }

    http.HandleFunc("/", GameStorageHandler)

    listenAt := "0.0.0.0:8080"
    log.Printf("Nochlehead server listening at `http://%s:%s`.\n", listenAt)
    err := http.ListenAndServe(listenAt, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}
