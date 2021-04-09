# Notification

Notification 


## Metadata
Standard kubernetes [meta.v1.ObjectMeta](https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

## Spec

Specifies the notification information for this Notification object.
It must has a one of next field.

* email
* webhook
* slack
  

### email property

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
smtpcfg|Yes|string|The name of SMTPConfig
from|Yes|string|The email account of sender
to|Yes|string|The email account of receiver
subject|Yes|string|The subject of mail
body|Yes|string|The body of mail
cc|No|string|-

### webhook property (not support yet)

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
url|Yes|string|-
message|Yes|string|-

### slack property (not support yet)

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
account|Yes|string|-
workspace|Yes|string|-
channel|Yes|string|-
message|Yes|string|-

## Status

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
type|-|string|Notification type(email, webhook, slack, etc)
endpoint|-|string|The endpoint for notification. (http://[notification_name].[notifier's_clusterip].nip.io)