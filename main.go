package main

import (
        "encoding/json"
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
                Analysis string `json:"analysis"`
        } `json:"data"`
}

// Alertmanager response struct
type AlertmanagerAlert struct {
        Labels map[string]string `json:"labels"`
}

// HTML content moved to external file

// Query Prometheus
func queryPrometheus(promQuery string) (PrometheusResponse, error) {

	prometheusURL  := "http://127.0.0.1:9090/api/v1/query"
	if os.Getenv("PROMETHEUS_URL") != "" {
	         prometheusURL = os.Getenv("PROMETHEUS_URL")
	}

        url := fmt.Sprintf("%s?query=%s", prometheusURL, promQuery)

        resp, err := http.Get(url)
	log.Info("Prometheus response:",resp)
        if err != nil {
                return PrometheusResponse{}, err
        }
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                return PrometheusResponse{}, err
        }

        var prometheusResponse PrometheusResponse
        err = json.Unmarshal(body, &prometheusResponse)
        if err != nil {
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
                return nil, err
        }
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
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
	log.GetFormatter().(*log.TextFormatter).SetTemplate(logTemplate)

        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                // Serve the HTML file
                http.ServeFile(w, r, "index.html")
        })

        http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
                query := r.URL.Query().Get("query")
                if query == "" {
                        http.Error(w, "Query parameter is required", http.StatusBadRequest)
                        return
                }

                prometheusResponse, err := queryPrometheus(query)
                if err != nil {
                        log.Errorf("Error querying Prometheus: %v", err)
                        http.Error(w, "Failed to fetch data from Prometheus", http.StatusInternalServerError)
                        return
                }

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

                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(alerts)
        })

        log.Info("Server is running on http://0.0.0.0:8080")
        log.Fatal(http.ListenAndServe(":8080", nil))
}
