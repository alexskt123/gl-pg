package main

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"flag"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"github.com/xh-dev-go/xhUtils/flagUtils/FlagString"
)

var mongoConn *mgo.Session

type MyEntity struct {
	Data []byte `json:"data" bson:"data"`
}

func createConnection(connVar, userVar, pwVar string) (*mgo.Session, error) {
	dialInfo := mgo.DialInfo{
		Addrs: []string{
			connVar},
		Username: userVar,
		Password: pwVar,
	}
	fmt.Println(dialInfo)
	tlsConfig := &tls.Config{}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}
	return mgo.DialWithInfo(&dialInfo)
}

func main() {
	connVar := flagString.New("conn", "connection").BindCmd()
	userVar := flagString.New("user", "user").BindCmd()
	pwVar := flagString.New("pw", "password").BindCmd()

	flag.Parse()

	var err error
	mongoConn, err = createConnection(connVar.Value(), userVar.Value(), pwVar.Value())
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/save", post)
	http.HandleFunc("/read", get)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func post(w http.ResponseWriter, req *http.Request) {
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	session := mongoConn.Copy()
	defer session.Close()

	entity := MyEntity{Data: payload}
	err = session.DB("test").C("data").Insert(entity)
	if err != nil {
		panic(err)
	}
}

func get(w http.ResponseWriter, req *http.Request) {
	session := mongoConn.Copy()
	defer session.Close()

	entity := MyEntity{}
	err := session.DB("test").C("data").Find(bson.M{}).One(&entity)
	if err != nil {
		panic(err)
	}

	w.Write(entity.Data)
	w.Write([]byte{10})
}