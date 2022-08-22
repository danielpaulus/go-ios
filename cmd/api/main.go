package main

import (
	"context"
	"github.com/danielpaulus/go-ios-api/service"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) >= 1 {
		//when launched by launchd, the argument will be the first argument unlike when
		//launched from the shell
		if os.Args[0] == "log-to-file" {
			logFile, err := os.OpenFile("go-ios.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				log.Errorf("error opening file: %v", err)
				return
			}
			defer logFile.Close()
			mw := io.MultiWriter(os.Stdout, logFile)
			logrus.SetOutput(mw)
			log.SetOutput(mw)
			log.SetFormatter(&log.JSONFormatter{})
		}
	}

	log.WithFields(log.Fields{"args": os.Args, "version": service.GetVersion()}).Infof("starting go-iOS-API")

	const address = "0.0.0.0:16800"
	srv := service.CreateHTTPServer(address)

	c := make(chan os.Signal, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.WithFields(log.Fields{"err": err}).Error("wrapper http server failed")
			c <- syscall.SIGABRT
		}
	}()
	log.Info("Server running on:" + address)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	signal := <-c
	log.Infof("os signal:%d received, closing..", signal)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	srv.Shutdown(ctx)
}
