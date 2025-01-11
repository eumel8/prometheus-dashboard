package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

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

// Alertmanager response struct
type AlertmanagerAlert struct {
	Labels map[string]string `json:"labels"`
}

// HTML content moved to external file

// Query Prometheus
func queryPrometheus(promQuery string) (PrometheusResponse, error) {

	prometheusURL := "http://127.0.0.1:9090/api/v1/query"
	if os.Getenv("PROMETHEUS_URL") != "" {
		prometheusURL = os.Getenv("PROMETHEUS_URL")
	}

	url := fmt.Sprintf("%s?query=%s", prometheusURL, promQuery)

	resp, err := http.Get(url)
	if err != nil {
		return PrometheusResponse{}, err
	}
	defer resp.Body.Close()

	log.Infof("Prometheus response status: %s", resp.Status)
	log.Debugf("Prometheus response body: %v", resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error querying Prometheus: %v", err)
		return PrometheusResponse{}, err
	}

	var prometheusResponse PrometheusResponse
	err = json.Unmarshal(body, &prometheusResponse)
	if err != nil {
		log.Errorf("Error unmarshalling prometheusResponse: %v", err)
		return PrometheusResponse{}, err
	}

	return prometheusResponse, nil
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
	log.Infof("Alertmanager response status: %s", resp.Status)
	log.Debugf("Alertmanager response body: %v", resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading body Alertmanager: %v", err)
		return nil, err
	}

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

func main() {
	logLevel := flag.String("logLevel", "info", "loglevel of app, e.g info, debug, warn, error, fatal")
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the HTML file
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if query == "" {
			log.Error("Query parameter is required")
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}

		prometheusResponse, err := queryPrometheus(query)
		if err != nil {
			log.Errorf("Error querying Prometheus: %v", err)
			http.Error(w, "Failed to fetch data from Prometheus", http.StatusInternalServerError)
			return
		}
		log.Debug("Prometheus handler response: %v", prometheusResponse)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prometheusResponse)
	})

	http.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		alerts, err := queryAlertmanager()
		if err != nil {
			log.Errorf("Error querying Alertmanager: %v", err)
			http.Error(w, "Failed to fetch data from Alertmanager", http.StatusInternalServerError)
			return
		}
		log.Debug("Alertmanager handler response: %v", alerts)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alerts)
	})

	log.Info("Server is running on http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
