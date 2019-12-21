package api

// import (
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"time"

// 	"github.com/go-chi/chi"
// 	"github.com/go-chi/chi/middleware"
// 	"github.com/oskoss/mi-casa/home"
// 	log "github.com/sirupsen/logrus"
// )

// func Start(port string, myHome *home.Home) {

// 	router := chi.NewRouter()
// 	router.Use(middleware.RequestID)
// 	router.Use(middleware.RealIP)
// 	router.Use(middleware.Logger)
// 	router.Use(middleware.Recoverer)
// 	router.Use(middleware.Timeout(60 * time.Second))
// 	router.Route("/v1/hvac", func(router chi.Router) {
// 		//TODO ADD STATUS
// 		// router.Get("/status", handleV1HVACStatus)
// 		router.Post("/temperature", handleV1HVACTemperature(myHome))
// 	})
// 	webServer := &http.Server{
// 		Addr:         "0.0.0.0:" + port,
// 		WriteTimeout: time.Second * 15,
// 		ReadTimeout:  time.Second * 15,
// 		IdleTimeout:  time.Second * 60,
// 		Handler:      router,
// 	}
// 	go func() {
// 		if err := webServer.ListenAndServe(); err != nil {
// 			log.WithFields(log.Fields{
// 				"err": err,
// 			}).Fatalf("server crashed")
// 		}
// 	}()

// 	channel := make(chan os.Signal, 1)
// 	signal.Notify(channel, os.Interrupt)

// 	<-channel
// }
