# NotificationTrigger

NotificationTrigger specifies the condition to perform notification. Trigger conditions are specified through comparison 
operators and two operands. One of the operands is retrieved by specifying the name of the Monitor and the path to the 
resource field, and the other is specified as a string.

When the trigger is executed, the result is added to the **.Status.History** field, and up to 5 are saved.
If the maximum allowd number is exceeded, the oldest one is deleted and new one saved.


## Metadata
Standard kubernetes [meta.v1.ObjectMeta](https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

## Spec

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
notification|Yes|string|The name of Notification to trigger on match condition
monitor|Yes|string|The name of Monitor to fetch operand1
fieldPath|Yes|string|The field path of fetched resource to evaluate as operand1 which from the monitor. (ex: hits.total.value)
op|Yes|string|The comparasion operator which to evaluate fieldPath with operand. (gt(<), gte(<=), eq(=), lte(>=), lt(>))
operand|Yes|string|operand2 to be compared

## Status

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
history|-|[]NotificationTriggerResult|History of the result 


### NotificationTriggerResult

**FieldName**|**Requried**|**Type**|**Description**
:-----:|:-----:|:-----:|:-----:
triggered|-|bool|If triggered or not
message|-|string|Message as to why the notification failed
updatedAt|-|string|Datetime of trigger executed