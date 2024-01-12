package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/Falokut/intern_test_task/internal/config"
	"github.com/Falokut/intern_test_task/internal/handler"
	"github.com/Falokut/intern_test_task/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const DefaultBalance = float32(100.0)

func main() {
	gin.SetMode(gin.ReleaseMode)

	cfg := config.GetConfig()
	logger := getLogger()
	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Fatal(err)
	}
	logger.SetLevel(logLevel)

	database, err := repository.NewPostgreDB(cfg.DBConfig)
	if err != nil {
		logger.Fatalf("can't connect to the database %s", err.Error())
	}
	defer database.Close()

	repo := repository.NewWalletRepository(logger, database, DefaultBalance)

	handler := handler.NewHandler(repo)
	router := handler.InitRouter()

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Listen.Host, cfg.Listen.Port),
		Handler: router.Handler(),
	}
	shutdown := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Errorf("Error while serving %s", err.Error())
			shutdown <- err
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGTERM)
	select {
	case <-quit:
		break
	case <-shutdown:
		break
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown error %v", err)
	} else {
		logger.Println("server exiting...")
	}
}

func getLogger() *logrus.Logger {
	l := logrus.New()
	l.SetReportCaller(true)

	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
		},
		DisableColors: false,
		FullTimestamp: true,
	}

	l.SetOutput(os.Stdout)
	return l
}
