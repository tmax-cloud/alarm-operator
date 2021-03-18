# Alarm Operator

Alarm-operator provides notification mechanism based on kubernetes.

## Prerequisites 

* go version v1.15.7+.
* kubectl version v1.18.6+.
* [ko](https://github.com/google/ko) version v0.6.0+
* Access to a Kubernetes v1.19.4+ cluster.

## Getting started

1. Install operator
    ```bash
    make install && make deploy
    ```

2. Generate SMTP credential Secret and SMTPConfig
    ```bash
    cd config/sample/
    # Edit SMTP account yours in smtp_auth_sample.yaml
    vi smtp_auth_sample.yaml
    kubectl apply -f smtp_auth_sample.yaml
    # Edit SMTP information in tmax.io_v1alpha1_smtpconfig.yaml
    vi tmax.io_v1alpha1_smtpconfig.yaml
    kubectl apply -f tmax.io_v1alpha1_smtpconfig.yaml
    ```

3. Generate Notification resource
    ```bash
    # Edit information email_notification.yaml
    kubectl apply -f email_notification.yaml
    ```

4. Send GET request to generated notification's Endpoint in pod
    ```bash
    curl <generated_endpoint_url>
    ```


## Feature

* Mail notification
* Webhook notification (working)
* Slack notification (working)

## Development


## License