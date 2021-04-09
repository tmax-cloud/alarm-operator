# Monitor

Monitor fetches resources that can be fetched through the REST API GET method at specified cycles. If resource excceed
30 characters, it is replaced by "..." and up to 5 result can be stored.

## Metadata
Standard kubernetes [meta.v1.ObjectMeta](https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

## Spec

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
url|Yes|string|REST API's endpoint to fetch resource
body|Yes|string|body for query if needed in target API spec.
interval|Yes|int|Time interval in seconds

## Status

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
history|-|[]MonitorResult|List of NotificationTriggerResult


### MonitorResult

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
status|-|bool|If fetching resource success or not
value|-|string|Fetched resource value
updatedAt|-|string|Datetime of fetching resource