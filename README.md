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

### Generate notification endpoint and Test notification
1. Generate SMTP credential Secret and SMTPConfig
    ```bash
    cd config/sample/
    # Edit SMTP account yours in smtp_auth_sample.yaml
    vi smtp_auth_sample.yaml
    kubectl apply -f smtp_auth_sample.yaml
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

### NotificationTrigger and Monitor
1. Generate Monitor resource
    ```bash
    # Edit information monitor.yaml
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