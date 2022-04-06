package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/iotexproject/iotex-core/pkg/log"
)

// HTTPServer handles requests from http protocol
type HTTPServer struct {
	svr         *http.Server
	web3Handler Web3Handler
}

// NewHTTPServer creates a new http server
func NewHTTPServer(port int, handler Web3Handler) *HTTPServer {
	svr := &HTTPServer{
		svr: &http.Server{
			Addr: ":" + strconv.Itoa(port),
		},
		web3Handler: handler,
	}

	mux := http.NewServeMux()
	mux.Handle("/", svr)
	svr.svr.Handler = mux
	return svr
}

// Start starts the http server
func (hSvr *HTTPServer) Start(_ context.Context) error {
	go func() {
		if err := hSvr.svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.L().Fatal("Node failed to serve.", zap.Error(err))
		}
	}()
	return nil
}

// Stop stops the http server
func (hSvr *HTTPServer) Stop(ctx context.Context) error {
	return hSvr.svr.Shutdown(ctx)
}

func (hSvr *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	httpResp := hSvr.web3Handler.HandlePOSTReq(req.Body)

	// write results into http reponse
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(httpResp); err != nil {
		log.Logger("api").Warn("fail to respond request.", zap.Error(err))
	}
}
