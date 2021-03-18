# SMTPConfig

SMTPConfig is settings for SMTP.

## Metadata
Standard kubernetes [meta.v1.ObjectMeta](https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

## Spec

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
host|Yes|string|The SMTP server's hostname
port|Yes|int|The SMTP server's port number
secret|Yes|string|The secret name which contain a SMTP account


## Status
