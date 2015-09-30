# Proboviro
An inflexible alerts notifier.

## Info
Proboviro is a webhook for Prometheus Alertmanager sending alarms as
push notification to an Android phone using NotifyMyAndroid.
* [Prometheus](https://github.com/prometheus/prometheus)
* [Alertmanager](https://github.com/prometheus/alertmanager)
* [NotifyMyAndroid](https://www.notifymyandroid.com)

It is a webserver which accepts only JSON POST with this structure:
```json
{
   "version": "1",
   "status": "firing",
   "alert": [
      {
         "summary": "summary",
         "description": "description",
         "labels": {
            "alertname": "TestAlert"
         },
         "payload": {
            "activeSince": "2015-06-01T12:55:47.356+01:00",
            "alertingRule": "ALERT TestAlert IF absent(metric_name) FOR
0y WITH ",
            "generatorURL":
"http://localhost:9090/graph#%5B%7B%22expr%22%3A%22absent%28metric_name%29%22%2C%22tab%22%3A0%7D%5D",
            "value": "1"
         }
      }
   ]
}
```

To run Proboviro you must specify an API key for NotifyMyAndroid.
