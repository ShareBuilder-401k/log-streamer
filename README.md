# log-streamer

Lambda to receive logs from cloudwatch and send them to elasticsearch through a signed request. App will flatten any nested json object or array, and place messages in field prefixed with `json_log_data`. The JSON object path will be preserved in the field name before being sent to elasticsearch.

## Infrastructure
Infrastructure is managed through terraform is divided between the lambda and triggers.

### Lambda Infrastructure
This deploys the lambda and its associated DLQ. This is where you will need to provide the hostname of your elasticsearch domain. 

### Trigger Infrastrucure
Add necessary log groups to the `log_groups` variable. This will cause the lambda to be invoked anytime a log is sent to the provided log_groups