# Alarm Operator

Alarm-operator provides notification mechanism based on kubernetes.

## Prerequisites 

* go version v1.15.7+.
* kubectl version v1.18.6+.
* Access to a Kubernetes v1.19.4+ cluster.

## Getting started

### Installation
1. Install operator
    ```bash
    make install && make deploy
    ```

### Generate Notification resource and Test notification endpoint
1. Generate SMTP credential Secret and SMTPConfig
    ```bash
    cd config/sample/
    # Edit SMTP information in smtpconfig.yaml
    vi smtpconfig.yaml
    kubectl apply -f smtpconfig.yaml
    ```

2. Generate Notification resource
    ```bash
    # Edit information email_notification.yaml
    kubectl apply -f email_notification.yaml
    ```

3. Send GET request to generated notification's Endpoint in pod
    ```bash
    curl -XPOST <generated_endpoint_url>
    ```

### Generate NotificationTrigger and Monitor resource and check
1. Generate Monitor resource
    ```bash
    # Edit monitor.yaml  before applying
    kubectl apply -f monitor.yaml
    ```

2. Verify fetching resource correctly 
    ```bash
    kubectl get monitor <generated_monitor_name> -o yaml
    ```
   
3. Generate NotificationTrigger resource
    ```bash
    # Edit information notificationtrigger.yaml
    kubectl apply -f notificationtrigger.yaml
    ```

4. Verify if trigger working correctly
    ```bash
    kubectl get monitor <generated_notificationtrigger_name> -o yaml
    ```

## Feature

* Mail notification
* Webhook notification (working)
* Slack notification (working)
* Monitoring resource and specify condition to trigger notification

## Development


## License