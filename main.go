package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"net/http"
	"os"

	log "github.com/gookit/slog"
)

const (
	logTemplate = "[{{datetime}}] [{{level}}] {{caller}} {{message}} \n"
)

// Prometheus response struct
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Thanos response struct
type ThanosResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
		Analysis map[string]interface{} `json:"analysis,omitempty"`
	} `json:"data"`
}

// Alertmanager response struct
type AlertmanagerAlert struct {
	Labels map[string]string `json:"labels"`
}

// HTML content moved to external file

// Query Prometheus
func queryPrometheus(promQuery string, thanosEnabled bool) (interface{}, error) {
	prometheusURL := "http://127.0.0.1:9090/api/v1/query"
	if os.Getenv("PROMETHEUS_URL") != "" {
		prometheusURL = os.Getenv("PROMETHEUS_URL")
	}

	url := fmt.Sprintf("%s?query=%s", prometheusURL, promQuery)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Info("Prometheus response status: ", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error querying Prometheus: %v", err)
		return nil, err
	}

	log.Debug("Prometheus response body: %s", body)

	if thanosEnabled {
		var thanosResponse ThanosResponse
		err = json.Unmarshal(body, &thanosResponse)
		if err != nil {
			log.Errorf("Error unmarshalling thanosResponse: %v", err)
			return nil, err
		}
		return thanosResponse, nil
	} else {
		var prometheusResponse PrometheusResponse
		err = json.Unmarshal(body, &prometheusResponse)
		if err != nil {
			log.Errorf("Error unmarshalling prometheusResponse: %v", err)
			return nil, err
		}
		return prometheusResponse, nil
	}
}

// Query Alertmanager
func queryAlertmanager() (map[string]int, error) {
	alertmanagerURL := "http://127.0.0.1:9093/api/v2/alerts"
	if os.Getenv("ALERTMANAGER_URL") != "" {
		alertmanagerURL = os.Getenv("ALERTMANAGER_URL")
	}

	resp, err := http.Get(alertmanagerURL)
	if err != nil {
		log.Errorf("Error querying Alertmanager: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Info("Alertmanager response status: ", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading body Alertmanager: %v", err)
		return nil, err
	}
	log.Debug("Alertmanager response body: %s", body)

	var alerts []AlertmanagerAlert
	err = json.Unmarshal(body, &alerts)
	if err != nil {
		return nil, err
	}

	// Aggregate alerts by severity
	severityCount := make(map[string]int)
	for _, alert := range alerts {
		severity := alert.Labels["severity"]
		if severity != "" {
			severityCount[severity]++
		}
	}

	return severityCount, nil
}

// LoggingMiddleware logs each incoming HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Infof("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		duration := time.Since(startTime)
		log.Infof("Handled request for %s in %v", r.URL.Path, duration)
	})
}

func main() {
	logLevel := flag.String("logLevel", "info", "loglevel of app, e.g info, debug, warn, error, fatal")
	thanosEnabled := flag.Bool("thanos", false, "enable Thanos response struct")
	flag.Parse()

	// set log level
	switch *logLevel {
	case "fatal":
		log.SetLogLevel(log.FatalLevel)
	case "trace":
		log.SetLogLevel(log.TraceLevel)
	case "debug":
		log.SetLogLevel(log.DebugLevel)
	case "error":
		log.SetLogLevel(log.ErrorLevel)
	case "warn":
		log.SetLogLevel(log.WarnLevel)
	case "info":
		log.SetLogLevel(log.InfoLevel)
	default:
		log.SetLogLevel(log.InfoLevel)
	}

	log.GetFormatter().(*log.TextFormatter).SetTemplate(logTemplate)

	http.Handle("/", LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve the HTML file
		http.ServeFile(w, r, "index.html")
	})))

	http.Handle("/api/query", LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if query == "" {
			log.Error("Query parameter is required")
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}

		response, err := queryPrometheus(query, *thanosEnabled)
		if err != nil {
			log.Errorf("Error querying Prometheus: %v", err)
			http.Error(w, "Failed to fetch data from Prometheus", http.StatusInternalServerError)
			return
		}
		log.Debug("Prometheus handler response: %v", response)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})))

	http.Handle("/api/alerts", LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		alerts, err := queryAlertmanager()
		if err != nil {
			log.Errorf("Error querying Alertmanager: %v", err)
			http.Error(w, "Failed to fetch data from Alertmanager", http.StatusInternalServerError)
			return
		}
		log.Debug("Alertmanager handler response: %v", alerts)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alerts)
	})))

	log.Info("Server is running on http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
