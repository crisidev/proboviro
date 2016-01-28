package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dustin/go-nma"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	APP_NAME    = "proboviro"
	APP_VERSION = "0.1"
	APP_AUTHOR  = "bigo@crisidev.org"
	APP_SITE    = "https://github.com/crisidev/proboviro"
)

var (
	lg         Logger
	severities = map[string]int{
		"info": 3,
		"page": 2,
	}

	// Flags
	flagDebug  = kingpin.Flag("debug", "enable debug mode").Short('D').Bool()
	flagBind   = kingpin.Flag("bind", "bind to address").Short('b').Default(":8082").String()
	flagApiKey = kingpin.Flag("apikey", "notifymyandroid API key").Short('a').Required().String()
)

// Simple logging structure
type Logger struct {
	output *log.Logger
	stderr *log.Logger
}

func init() {
	kingpin.Version(APP_VERSION)
	kingpin.Parse()
	lg.SetupLog()
}

// Setup debugging to stderr
func (l Logger) SetupLog() {
	lg.output = log.New(os.Stdout, "", 0)
	lg.output.SetFlags(log.Ldate | log.Ltime)
}

// Logs to stardard output
func (l Logger) Out(msg string) {
	if *flagDebug {
		l.output.Println(msg)
	}
}

// Logs to standard output without carriage return
func (l Logger) OutRaw(msg string) {
	if *flagDebug {
		fmt.Printf(msg)
	}
}

// Logs a fatal error
func (l Logger) Fatal(err error) {
	if err != nil {
		l.output.Printf(fmt.Sprintf("ERROR: %s", err.Error()))
	}
	os.Exit(1)
}

// Logs an erro
func (l Logger) Error(err error) {
	if err != nil {
		l.output.Printf(fmt.Sprintf("ERROR: %s", err.Error()))
	}
}

type SingleAlert struct {
	Annotations struct {
		ActiveSince  string `json:"activeSince"`
		AlertingRule string `json:"alertingRule"`
		Description  string `json:"description"`
		GeneratorURL string `json:"generatorURL"`
		Runbook      string `json:"runbook"`
		Summary      string `json:"summary"`
		Value        string `json:"value"`
	} `json:"annotations"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
	StartsAt     string            `json:"startsAt"`
}

type Alerts struct {
	Alert   []SingleAlert `json:"alert"`
	Status  string        `json:"status"`
	Version string        `json:"version"`
}

func DoAlert(rw http.ResponseWriter, req *http.Request, notifier *nma.NMA) {
	alerts, err := DecodeJson(req.Body)
	if err != nil {
		lg.Error(err)
	}
	for _, alert := range alerts.Alert {
		lg.Out(fmt.Sprintf("alert struct: %+v\n", alert))
		go HandleAlert(alert, notifier)
	}
}

func DecodeJson(body io.ReadCloser) (alerts Alerts, err error) {
	decoder := json.NewDecoder(body)
	err = decoder.Decode(&alerts)
	if err != nil {
		return
	}
	return
}

func HandleAlert(alert SingleAlert, notifier *nma.NMA) {
	severity := alert.Labels["severity"]
	if severity != "" {
		lg.Out(fmt.Sprintf("got new alert with severity: %s. send to notifymyandroid", severity))
		if severities[severity] == 0 {
			severity = "info"
		}
		go Page(alert, severities[severity], notifier)
	} else {
		lg.Error(errors.New("\"severity\" tag not found. this alert is not for me"))
	}
}

func Page(alert SingleAlert, severity int, notifier *nma.NMA) {
	notification := nma.Notification{
		Application: fmt.Sprintf("%s:%s", APP_NAME, alert.Labels["alertname"]),
		Description: alert.Annotations.Description,
		Event:       alert.Annotations.Summary,
		Priority:    1,
		URL:         alert.Annotations.GeneratorURL,
	}
	if err := notifier.Notify(&notification); err != nil {
		lg.Error(err)
	}
}

func main() {
	lg.Out(fmt.Sprintf("starting proboviro. debug: %t, key: %s, bind: %s", *flagDebug, *flagApiKey, *flagBind))
	notifier := nma.New(*flagApiKey)
	http.HandleFunc("/irc", func(w http.ResponseWriter, req *http.Request) {
		DoAlert(w, req, notifier)
	})
	err := http.ListenAndServe(*flagBind, nil)
	if err != nil {
		lg.Fatal(err)
	}
	os.Exit(0)
}
