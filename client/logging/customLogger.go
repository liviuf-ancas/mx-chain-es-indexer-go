package logging

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("indexer/client/requests")

type CustomLogger struct{}

func (cl *CustomLogger) LogRoundTrip(
	req *http.Request,
	res *http.Response,
	err error,
	_ time.Time,
	dur time.Duration,
) error {
	var (
		nReq int64
		nRes int64
	)

	if req != nil && req.Body != nil && req.Body != http.NoBody {
		nReq, _ = io.Copy(ioutil.Discard, req.Body)
	}
	if res != nil && res.Body != nil && res.Body != http.NoBody {
		nRes, _ = io.Copy(ioutil.Discard, res.Body)
	}

	if err != nil {
		log.Warn("elastic client", "error", err.Error())
	}

	if req != nil && res != nil {
		logInformation(req, res, err, dur, nReq, nRes)
	}

	return nil
}

func logInformation(
	req *http.Request,
	res *http.Response,
	err error,
	dur time.Duration,
	nReq int64,
	nRes int64,
) {
	logData := []interface{}{
		"method", req.Method,
		"status code", res.StatusCode,
		"duration", dur,
		"request bytes", nReq,
		"response bytes", nRes,
		"URL", req.URL.String(),
	}
	if err != nil {
		log.Warn("elastic client", logData...)
		return
	}

	log.Debug("elastic client", logData...)
}

// RequestBodyEnabled makes the client pass request body to logger
func (cl *CustomLogger) RequestBodyEnabled() bool {
	return true
}

// ResponseBodyEnabled makes the client pass response body to logger
func (cl *CustomLogger) ResponseBodyEnabled() bool {
	return true
}
