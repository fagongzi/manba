Restful
--------------
In addition to GRPC API, API server supports Restful HTTP API. The following is about Restful API.

Attention:
- Because the time unit in Go is nanosecond (`time.Duration`), time in APIs needs to be in nanosecond. 1 second equals to 1,000,000,000 nanoseconds.
- body in defaultValue needs to be in base64, otherwise `base64.CorruptInputError=illegal base64 data` error occurs. This is due to `type HTTPResult struct {Body []byte}`，For detailed information, please visit [[]byte encodes as a base64-encoded ](https://stackoverflow.com/questions/31449610/illegal-base64-data-error-message)。
- When configuring `renderTemplate`, if `flatAttrs` is true, name can be omitted. If false, name must be configured not to be null. This requires data checking when calling APIs and wrong configuration leads to service unavailability.
- `defaultValue` in Nodes should match `renderTemplate`'s, otherwise `Key path not found` error occurs.

## ENUM
### Status
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|Down|0||
|Up|1||

### CircuitStatus
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|Open|0||
|Half|1||
|Close|2||

### LoadBalance
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|RoundRobin|0||
|IPHash|1|Currently Version Not Supported|

### Protocol
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|HTTP|0||
|Grpc|1|Currently Version Not Supported|
|Dubbo|1|Currently Version Not Supported|
|SpringCloud|2|Currently Version Not Supported|

### Source
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|QueryString|0||
|FormData|1||
|JSONBody|2||
|Header|3||
|Cookie|4||
|PathValue|5||

### RuleType
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|RuleRegexp|0||

### CMP
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|CMPEQ|0||
|CMPLT|1||
|CMPLE|2||
|CMPGT|3||
|CMPGE|4||
|CMPIn|5||
|CMPMatch|6|

### RoutingStrategy
|Name|Value|Comment|
| -------------|:-------------:| -------------|
|Copy|0||
|Split|1||

## Cluster
### New/Update
|URL|Method|
| -------------|:-------------:|
|/v1/clusters|PUT|

JSON Body
```json
{
    "id":1,
    "name":"cluster name",
    "loadBalance":0
}
```
1 in id field means update.

Reponse
```json
{
    "code":0,
    "data":1
}
```
data field represents cluster id

### Delete
|URL|Method|
| -------------|:-------------:|
|/v1/clusters/{id}|DELETE|

Reponse
```json
{
    "code":0,
    "data":"null"
}
```
### Query
|URL|Method|
| -------------|:-------------:|
|/v1/clusters/{id}|GET|

Reponse
```json
{
    "code":0,
    "data":{
        "id":1,
        "name":"cluster name",
        "loadBalance":0
    }
}
```
data field represents cluster id

### List
|URL|Method|
| -------------|:-------------:|
|/v1/clusters?after=xx&limit=xx|GET|

after：the recent created cluster id
limit：how many id we need

Reponse
```json
{
    "code":0,
    "data":[
        {
            "id":1,
            "name":"cluster name",
            "loadBalance":0
        },
        {
            "id":2,
            "name":"cluster name",
            "loadBalance":0
        },
        {
            "id":3,
            "name":"cluster name",
            "loadBalance":0
        }
    ]
}
```
data field is a collection of clusters
Next batch: /v1/clusters?after=3&limit=3

### Query all the servers binded
|URL|Method|
| -------------|:-------------:|
|/v1/clusters/{id}/binds|GET|

Reponse
```json
{
    "code":0,
    "data":[
        1,
        2,
        3
    ]
}
```
data field represents ids of all servers binded to the cluster.

### Unbind All Server
|URL|Method|
| -------------|:-------------:|
|/v1/clusters/{id}/binds|DELETE|

Reponse
```json
{
    "code":0,
    "data":"null"
}
```

## Server
### New/Update
|URL|Method|
| -------------|:-------------:|
|/v1/servers|PUT|

JSON Body
```json
{
    "id":1,
    "addr":"127.0.0.1:8080",
    "protocol":0,
    "maxQPS":100,
    "heathCheck":{
        "path":"/check-heath",
        "body":"OK",
        "checkInterval":10000000000,
        "timeout":30000000000
    },
    "circuitBreaker":{
        "closeTimeout":10000000000,
        "halfTrafficRate":50,
        "rateCheckPeriod":10000000000,
        "failureRateToClose":20,
        "succeedRateToOpen":30
    }
}
```
1 in id field means update.

Reponse
```json
{
    "code":0,
    "data":1
}
```
data field means server id

### Delete
|URL|Method|
| -------------|:-------------:|
|/v1/servers/{id}|DELETE|

Reponse
```json
{
    "code":0,
    "data":"null"
}
```
### Query
|URL|Method|
| -------------|:-------------:|
|/v1/servers/{id}|GET|

Reponse
```json
{
    "code":0,
    "data":{
        "id":1,
        "addr":"127.0.0.1:8080",
        "protocol":0,
        "maxQPS":100,
        "heathCheck":{
            "path":"/check-heath",
            "body":"OK",
            "checkInterval":10000000000,
            "timeout":30000000000
        },
        "circuitBreaker":{
            "closeTimeout":10000000000,
            "halfTrafficRate":50,
            "rateCheckPeriod":10000000000,
            "failureRateToClose":20,
            "succeedRateToOpen":30
        }
    }
}
```
data field reflects server info.

### List
|URL|Method|
| -------------|:-------------:|
|/v1/servers?after=xx&limit=xx|GET|

after：the recent server id
limit：how many servers from which we need info

Reponse
```json
{
    "code":0,
    "data":[
        {
            "id":1,
            "addr":"127.0.0.1:8081",
            "protocol":0,
            "maxQPS":100,
            "heathCheck":{
                "path":"/check-heath",
                "body":"OK",
                "checkInterval":10000000000,
                "timeout":30000000000
            },
            "circuitBreaker":{
                "closeTimeout":10000000000,
                "halfTrafficRate":50,
                "rateCheckPeriod":10000000000,
                "failureRateToClose":20,
                "succeedRateToOpen":30
            }
        },
        {
            "id":2,
            "addr":"127.0.0.1:8082",
            "protocol":0,
            "maxQPS":100,
            "heathCheck":{
                "path":"/check-heath",
                "body":"OK",
                "checkInterval":10000000000,
                "timeout":30000000000
            },
            "circuitBreaker":{
                "closeTimeout":10000000000,
                "halfTrafficRate":50,
                "rateCheckPeriod":10000000000,
                "failureRateToClose":20,
                "succeedRateToOpen":30
            }
        },
        {
            "id":3,
            "addr":"127.0.0.1:8083",
            "protocol":0,
            "maxQPS":100,
            "heathCheck":{
                "path":"/check-heath",
                "body":"OK",
                "checkInterval":10000000000,
                "timeout":30000000000
            },
            "circuitBreaker":{
                "closeTimeout":10000000000,
                "halfTrafficRate":50,
                "rateCheckPeriod":10000000000,
                "failureRateToClose":20,
                "succeedRateToOpen":30
            }
        }
    ]
}
```
data fields has the collection of servers
The next batch: /v1/servers?after=3&limit=3

## Bind
### Add
|URL|Method|
| -------------|:-------------:|
|/v1/binds|PUT|

JSON Body
```json
{
    "clusterID":1,
    "serverID":2
}
```

Reponse
```json
{
    "code":0,
    "data":"null"
}
```

### Delete
|URL|Method|
| -------------|:-------------:|
|/v1/binds|DELETE|

JSON Body
```json
{
    "clusterID":1,
    "serverID":2
}
```

Reponse
```json
{
    "code":0,
    "data":"null"
}
```

## API
### New/Update
|URL|Method|
| -------------|:-------------:|
|/v1/apis|PUT|

JSON Body
```json
{
    "id":1,
    "name":"demo-api",
    "urlPattern":"^/api/users/(\\d+)$",
    "method":"GET",
    "domain":"www.xxx.com",
    "status":1,
    "ipAccessControl":{
        "whitelist":[
            "127.*",
            "192.168.*",
            "172.17.*",
            "172.17.1.1"
        ],
        "blacklist":[
            "127.*",
            "192.168.*",
            "172.17.*",
            "172.17.1.1"
        ]
    },
    "defaultValue":{
        "code": 200,
        "body":"aGVsbG8gd29ybGQ=",
        "headers":[
            {
                "name":"xx",
                "value":"xx"
            }
        ],
        "cookies":[
            {
                "name":"xx",
                "value":"xx"
            }
        ]
    },
    "nodes":[
        {
            "clusterID":1,
            "urlRewrite":"/users?id=$1",
            "attrName":"user",
            "validations":[
                {
                    "parameter":{
                        "name":"id",
                        "source":0,
                        "index":0
                    },
                    "required":true,
                    "rules":[
                        {
                            "ruleType":0,
                            "expression":"^\\d+$"
                        }
                    ]
                }
            ],
            "cache":{
                "keys":[
                    {
                        "name":"id",
                        "source":0,
                        "index":0
                    }
                ],
                "deadline":100,
                "conditions":[
                    {
                        "parameter":{
                            "name":"id",
                            "source":0,
                            "index":0
                        },
                        "cmp":3,
                        "expect":"100"
                    }
                ]
            }
        },
        {
            "clusterID":2,
            "urlRewrite":"/users/$1/account",
            "attrName":"account",
            "validations":[
                {
                    "parameter":{
                        "name":"",
                        "source":5,
                        "index":1
                    },
                    "required":true,
                    "rules":[
                        {
                            "ruleType":0,
                            "expression":"^\\d+$"
                        }
                    ]
                }
            ],
            "defaultValue":{
                "code": 200,
                "body":"eyJjb2RlIjoxLCAiZGF0YSI6eyJtZXNzYWdlIjogIuacjeWKoeS4jeWPr+eU\nqOi/meaYr+afkOiKgueCueeahOm7mOiupOWAvO+8gSJ9fQ==",
                "headers":[
                    {
                        "name":"Content-Type",
                        "value":"application/json"
                    }
                ]
            }
        }
    ],
    "authFilter":"CUSTOM_JWT",
    "renderTemplate":{
        "objects":[
            {
                "name":"",
                "attrs":[
                    {
                        "name":"user",
                        "extractExp":"user.data"
                    },
                    {
                        "name":"account",
                        "extractExp":"account.data"
                    }
                ],
                "flatAttrs":true
            }
        ]
    },
    "useDefault": false,
    "matchRule": 0,
    "position": 0,
    "tags": [
        {
            "name": "tag3",
            "value": "value3"
        }
    ]
}
```
1 in id field means update.

Reponse
```json
{
    "code":0,
    "data":1
}
```
data field represents API id

### Delete
|URL|Method|
| -------------|:-------------:|
|/v1/apis/{id}|DELETE|

Reponse
```json
{
    "code":0,
    "data":"null"
}
```
### Query
|URL|Method|
| -------------|:-------------:|
|/v1/apis/{id}|GET|

Reponse
```json
{
    "code":0,
    "data":{
        "id":1,
        "name":"demo-api",
        "urlPattern":"^/api/users/(\\d+)$",
        "method":"GET",
        "domain":"www.xxx.com",
        "status":1,
        "ipAccessControl":{
            "whitelist":[
                "127.*",
                "192.168.*",
                "172.17.*",
                "172.17.1.1"
            ],
            "blacklist":[
                "127.*",
                "192.168.*",
                "172.17.*",
                "172.17.1.1"
            ]
        },
        "defaultValue":{
            "code": 200,
            "body":"aGVsbG8gd29ybGQ=",
            "headers":[
                {
                    "name":"xx",
                    "value":"xx"
                }
            ],
            "cookies":[
                {
                    "name":"xx",
                    "value":"xx"
                }
            ]
        },
        "nodes":[
            {
                "clusterID":1,
                "urlRewrite":"/users?id=$1",
                "attrName":"user",
                "validations":[
                    {
                        "parameter":{
                            "name":"id",
                            "source":0,
                            "index":0
                        },
                        "required":true,
                        "rules":[
                            {
                                "ruleType":0,
                                "expression":"^\\d+$"
                            }
                        ]
                    }
                ],
                "cache":{
                    "keys":[
                        {
                            "name":"id",
                            "source":0,
                            "index":0
                        }
                    ],
                    "deadline":100,
                    "conditions":[
                        {
                            "parameter":{
                                "name":"id",
                                "source":0,
                                "index":0
                            },
                            "cmp":3,
                            "expect":"100"
                        }
                    ]
                }
            },
            {
                "clusterID":2,
                "urlRewrite":"/users/$1/account",
                "attrName":"account",
                "validations":[
                    {
                        "parameter":{
                            "name":"",
                            "source":5,
                            "index":1
                        },
                        "required":true,
                        "rules":[
                            {
                                "ruleType":0,
                                "expression":"^\\d+$"
                            }
                        ]
                    }
                ]
            }
        ],
        "authFilter":"CUSTOM_JWT",
        "renderTemplate":{
            "objects":[
                {
                    "name":"",
                    "attrs":[
                        {
                            "name":"user",
                            "extractExp":"user.data"
                        },
                        {
                            "name":"account",
                            "extractExp":"account.data"
                        }
                    ],
                    "flatAttrs":true
                }
            ]
        },
        "useDefault": false,
        "matchRule": 0,
        "position": 0,
        "tags": [
            {
                "name": "tag3",
                "value": "value3"
            }
        ]
    }
}
```
data field represents api id.

### List
|URL|Method|
| -------------|:-------------:|
|/v1/apis?after=xx&limit=xx|GET|

after：the recent api id
limit: how many apis from which we need info

Reponse
```json
{
    "code":0,
    "data":[
        {
            "id":1,
            "name":"demo-api",
            "urlPattern":"^/api/users/(\\d+)$",
            "method":"GET",
            "domain":"www.xxx.com",
            "status":1,
            "ipAccessControl":{
                "whitelist":[
                    "127.*",
                    "192.168.*",
                    "172.17.*",
                    "172.17.1.1"
                ],
                "blacklist":[
                    "127.*",
                    "192.168.*",
                    "172.17.*",
                    "172.17.1.1"
                ]
            },
            "defaultValue":{
                "code": 200,
                "body":"aGVsbG8gd29ybGQ=",
                "headers":[
                    {
                        "name":"xx",
                        "value":"xx"
                    }
                ],
                "cookies":[
                    {
                        "name":"xx",
                        "value":"xx"
                    }
                ]
            },
            "nodes":[
                {
                    "clusterID":1,
                    "urlRewrite":"/users?id=$1",
                    "attrName":"user",
                    "validations":[
                        {
                            "parameter":{
                                "name":"id",
                                "source":0,
                                "index":0
                            },
                            "required":true,
                            "rules":[
                                {
                                    "ruleType":0,
                                    "expression":"^\\d+$"
                                }
                            ]
                        }
                    ],
                    "cache":{
                        "keys":[
                            {
                                "name":"id",
                                "source":0,
                                "index":0
                            }
                        ],
                        "deadline":100,
                        "conditions":[
                            {
                                "parameter":{
                                    "name":"id",
                                    "source":0,
                                    "index":0
                                },
                                "cmp":3,
                                "expect":"100"
                            }
                        ]
                    }
                },
                {
                    "clusterID":2,
                    "urlRewrite":"/users/$1/account",
                    "attrName":"account",
                    "validations":[
                        {
                            "parameter":{
                                "name":"",
                                "source":5,
                                "index":1
                            },
                            "required":true,
                            "rules":[
                                {
                                    "ruleType":0,
                                    "expression":"^\\d+$"
                                }
                            ]
                        }
                    ]
                }
            ],
            "authFilter":"CUSTOM_JWT",
            "renderTemplate":{
                "objects":[
                    {
                        "name":"",
                        "attrs":[
                            {
                                "name":"user",
                                "extractExp":"user.data"
                            },
                            {
                                "name":"account",
                                "extractExp":"account.data"
                            }
                        ],
                        "flatAttrs":true
                    }
                ]
            },
            "useDefault": false,
            "matchRule": 0,
            "position": 0,
            "tags": [
                {
                    "name": "tag1",
                    "value": "value1"
                }
            ]
        },
        {
            "id":2,
            "name":"demo-api-2",
            "urlPattern":"^/api/users/(\\d+)$",
            "method":"GET",
            "domain":"www.xxx.com",
            "status":1,
            "ipAccessControl":{
                "whitelist":[
                    "127.*",
                    "192.168.*",
                    "172.17.*",
                    "172.17.1.1"
                ],
                "blacklist":[
                    "127.*",
                    "192.168.*",
                    "172.17.*",
                    "172.17.1.1"
                ]
            },
            "defaultValue":{
                "code": 200,
                "body":"aGVsbG8gd29ybGQ=",
                "headers":[
                    {
                        "name":"xx",
                        "value":"xx"
                    }
                ],
                "cookies":[
                    {
                        "name":"xx",
                        "value":"xx"
                    }
                ]
            },
            "nodes":[
                {
                    "clusterID":1,
                    "urlRewrite":"/users?id=$1",
                    "attrName":"user",
                    "validations":[
                        {
                            "parameter":{
                                "name":"id",
                                "source":0,
                                "index":0
                            },
                            "required":true,
                            "rules":[
                                {
                                    "ruleType":0,
                                    "expression":"^\\d+$"
                                }
                            ]
                        }
                    ],
                    "cache":{
                        "keys":[
                            {
                                "name":"id",
                                "source":0,
                                "index":0
                            }
                        ],
                        "deadline":100,
                        "conditions":[
                            {
                                "parameter":{
                                    "name":"id",
                                    "source":0,
                                    "index":0
                                },
                                "cmp":3,
                                "expect":"100"
                            }
                        ]
                    }
                },
                {
                    "clusterID":2,
                    "urlRewrite":"/users/$1/account",
                    "attrName":"account",
                    "validations":[
                        {
                            "parameter":{
                                "name":"",
                                "source":5,
                                "index":1
                            },
                            "required":true,
                            "rules":[
                                {
                                    "ruleType":0,
                                    "expression":"^\\d+$"
                                }
                            ]
                        }
                    ]
                }
            ],
            "authFilter":"CUSTOM_JWT",
            "renderTemplate":{
                "objects":[
                    {
                        "name":"",
                        "attrs":[
                            {
                                "name":"user",
                                "extractExp":"user.data"
                            },
                            {
                                "name":"account",
                                "extractExp":"account.data"
                            }
                        ],
                        "flatAttrs":true
                    }
                ]
            },
            "useDefault": false,
            "matchRule": 0,
            "position": 0,
            "tags": [
                {
                    "name": "tag2",
                    "value": ""
                }
            ]
        },
        {
            "id":3,
            "name":"demo-api-3",
            "urlPattern":"^/api/users/(\\d+)$",
            "method":"GET",
            "domain":"www.xxx.com",
            "status":1,
            "ipAccessControl":{
                "whitelist":[
                    "127.*",
                    "192.168.*",
                    "172.17.*",
                    "172.17.1.1"
                ],
                "blacklist":[
                    "127.*",
                    "192.168.*",
                    "172.17.*",
                    "172.17.1.1"
                ]
            },
            "defaultValue":{
                "code": 200,
                "body":"aGVsbG8gd29ybGQ=",
                "headers":[
                    {
                        "name":"xx",
                        "value":"xx"
                    }
                ],
                "cookies":[
                    {
                        "name":"xx",
                        "value":"xx"
                    }
                ]
            },
            "nodes":[
                {
                    "clusterID":1,
                    "urlRewrite":"/users?id=$1",
                    "attrName":"user",
                    "validations":[
                        {
                            "parameter":{
                                "name":"id",
                                "source":0,
                                "index":0
                            },
                            "required":true,
                            "rules":[
                                {
                                    "ruleType":0,
                                    "expression":"^\\d+$"
                                }
                            ]
                        }
                    ],
                    "cache":{
                        "keys":[
                            {
                                "name":"id",
                                "source":0,
                                "index":0
                            }
                        ],
                        "deadline":100,
                        "conditions":[
                            {
                                "parameter":{
                                    "name":"id",
                                    "source":0,
                                    "index":0
                                },
                                "cmp":3,
                                "expect":"100"
                            }
                        ]
                    }
                },
                {
                    "clusterID":2,
                    "urlRewrite":"/users/$1/account",
                    "attrName":"account",
                    "validations":[
                        {
                            "parameter":{
                                "name":"",
                                "source":5,
                                "index":1
                            },
                            "required":true,
                            "rules":[
                                {
                                    "ruleType":0,
                                    "expression":"^\\d+$"
                                }
                            ]
                        }
                    ]
                }
            ],
            "authFilter":"CUSTOM_JWT",
            "renderTemplate":{
                "objects":[
                    {
                        "name":"",
                        "attrs":[
                            {
                                "name":"user",
                                "extractExp":"user.data"
                            },
                            {
                                "name":"account",
                                "extractExp":"account.data"
                            }
                        ],
                        "flatAttrs":true
                    }
                ]
            },
            "useDefault": false,
            "matchRule": 0,
            "position": 0,
            "tags": [
                {
                    "name": "tag3",
                    "value": "value3"
                }
            ]
        }
    ]
}
```
data field represents a list of apis
The next batch: /v1/apis?after=3&limit=3

## Routing
### New/Update
|URL|Method|
| -------------|:-------------:|
|/v1/routings|PUT|

JSON Body
```json
{
    "id":1,
    "clusterID":2,
    "conditions":[
        {
            "parameter":{
                "name":"id",
                "source":4,
                "index":0
            },
            "cmp":6,
            "expect":"^.+[2]$"
        }
    ],
    "strategy":1,
    "trafficRate":10,
    "status":1,
    "api":1,
    "name":"test-AB"
}
```
1 in id field means update.

Reponse
```json
{
    "code":0,
    "data":1
}
```
data field has routing id.

### Delete
|URL|Method|
| -------------|:-------------:|
|/v1/routings/{id}|DELETE|

Reponse
```json
{
    "code":0,
    "data":"null"
}
```
### Query
|URL|Method|
| -------------|:-------------:|
|/v1/routings/{id}|GET|

Reponse
```json
{
    "code":0,
    "data":{
        "id":1,
        "clusterID":2,
        "conditions":[
            {
                "parameter":{
                    "name":"id",
                    "source":4,
                    "index":0
                },
                "cmp":6,
                "expect":"^.+[2]$"
            }
        ],
        "strategy":1,
        "trafficRate":10,
        "status":1,
        "api":1,
        "name":"test-AB"
    }
}
```
data field reflect the queried routing.

### List
|URL|Method|
| -------------|:-------------:|
|/v1/routings?after=xx&limit=xx|GET|

after：the recent routing id
limit：how many records

Reponse
```json
{
    "code":0,
    "data":[
        {
            "id":1,
            "clusterID":2,
            "conditions":[
                {
                    "parameter":{
                        "name":"id",
                        "source":4,
                        "index":0
                    },
                    "cmp":6,
                    "expect":"^.+[2]$"
                }
            ],
            "strategy":1,
            "trafficRate":10,
            "status":1,
            "api":1,
            "name":"test-AB"
        },
        {
            "id":2,
            "clusterID":2,
            "conditions":[
                {
                    "parameter":{
                        "name":"id",
                        "source":4,
                        "index":0
                    },
                    "cmp":6,
                    "expect":"^.+[2]$"
                }
            ],
            "strategy":1,
            "trafficRate":10,
            "status":1,
            "api":1,
            "name":"test-AB"
        },
        {
            "id":3,
            "clusterID":2,
            "conditions":[
                {
                    "parameter":{
                        "name":"id",
                        "source":4,
                        "index":0
                    },
                    "cmp":6,
                    "expect":"^.+[2]$"
                }
            ],
            "strategy":1,
            "trafficRate":10,
            "status":1,
            "api":1,
            "name":"test-AB"
        }
    ]
}
```
data fields has a collection of server info
The next batch: /v1/routings?after=3&limit=3
