package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/runns/shared/middleware"
	"github.com/runns/shared/utils"
)

type route struct{ prefix, envKey string }

var routes = []route{
	{"/api/auth", "AUTH_SERVICE_URL"},
	{"/api/orders", "ORDER_SERVICE_URL"},
	{"/api/dispatch", "DISPATCH_SERVICE_URL"},
	{"/api/payments", "PAYMENT_SERVICE_URL"},
	{"/api/realtime", "REALTIME_SERVICE_URL"},
	{"/ws", "REALTIME_SERVICE_URL"},
	{"/api/riders", "RIDER_SERVICE_URL"},
	{"/api/upload", "UPLOAD_SERVICE_URL"},
}

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "gateway healthy", nil) })
	for _, rt := range routes {
		targetURL := os.Getenv(rt.envKey)
		if targetURL == "" {
			log.Printf("[gateway] WARNING: %s not set, skipping %s", rt.envKey, rt.prefix)
			continue
		}
		tgt, err := url.Parse(targetURL)
		if err != nil {
			log.Fatalf("invalid URL for %s: %v", rt.envKey, err)
		}
		proxy := httputil.NewSingleHostReverseProxy(tgt)
		prefix := rt.prefix
		mux.HandleFunc(prefix+"/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[gateway] %s %s → %s", r.Method, r.URL.Path, targetURL)
			proxy.ServeHTTP(w, r)
		})
		mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[gateway] %s %s → %s", r.Method, r.URL.Path, targetURL)
			proxy.ServeHTTP(w, r)
		})
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { utils.NotFound(w, "route not found") })
	handler := middleware.Chain(mux, middleware.CORS, middleware.Logger)
	port := utils.GetEnv("GATEWAY_PORT", "8001")
	log.Printf("[api-gateway] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		log.Fatal(err)
	}
}
