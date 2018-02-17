package main

import (
	"os"
	"os/signal"
	"time"

	ppf "github.com/sudachen/pprof/util"
)

func main() {

	ppf.Start(5*time.Second, 8080)
	// Will open Pprof WebUI on localhost:8080
	//  and update shadow profile every 5 sec
        // To update current pprof profile to last shadow 
        //  reopen http://localhost:8080/ 
        //  or use update command form menu VIEW

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	signal.Stop(c)

	ppf.Stop()
}