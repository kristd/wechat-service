package module

import (
    "gopkg.in/mgo.v2"
    "sync"
    "service/conf"
    "github.com/golang/glog"
)

var (
    once sync.Once
    dbSession *mgo.Session
)

func GetDBInstant() *mgo.Session {
    once.Do(func() {
        var err error
        dbSession = nil

        dbSession, err = mgo.Dial(conf.Config.MONGODB)
        if err != nil {
            glog.Error("[GetDBInstant] Connect to mongodb failed, err = ", err)
        }
    })

    return dbSession
}

func DisConnect() {
    if dbSession != nil {
        dbSession.Close()
    }
}
