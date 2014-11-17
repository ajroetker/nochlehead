package main

import (
    "fmt"
    "log"
    "io"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "os"
    mgo  "gopkg.in/mgo.v2"
)

var session *mgo.Session

func init() {
    // These are the environment variables set by docker
    // when linking containers where the name of the elasticsearch
    // container is `mongo`
    mgoHost := os.Getenv("MONGO_PORT_27017_TCP_ADDR")
    var mgoUrl string
    if mgoHost == "" {
        mgoUrl = mgoHost
    } else {
        mgoUrl = "localhost"
    }
    session, err := mgo.Dial(mgoUrl)
    logFatalf("failed to connect to mongo", err)
    logFatalf("failed to migrate indexes", initdb(session))
}

func logFatalf( msg string, err error ) {
    var prefix string
    if msg != "" {
        prefix = fmt.Sprintf("%v: ", msg)
    }
    if err != nil {
        log.Fatalf("%v%v", prefix, err)
    }
}

func initdb(session *mgo.Session) error {
    c := session.DB("pinnochle").C("finished")
    index := mgo.Index{
        Key: []string{"sequence", "trump", "traded", "bid"},
        Unique: true,
        DropDups: true,
        Background: true,
        Sparse: true,
        Name: "sequence_trump_traded_bid",
    }
    return c.EnsureIndex(index)
}

func ( c *CompiledGame ) Store() error {
    collection := session.DB("pinnochle").C("finished")
    return collection.Insert(c)
}

func (c *CompiledGame) Validate() error {
    return nil
}

func ReadGame( body io.Reader, c *CompiledGame ) error {
    contents, err := ioutil.ReadAll(body)
    if err != nil {
        return err
    }
    if len(contents) == 0 {
        return fmt.Errorf("no data in request")
    }
    return json.Unmarshal(contents, c)
}

func GameStorageHandler(w http.ResponseWriter, r *http.Request) {
    var c *CompiledGame
    if err := ReadGame(r.Body, c); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    if err := c.Validate(); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    if err := c.Store(); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}
