Restful
--------------
API Server除了支持GRPC的接口以外，还支持HTTP的Restful接口，这里给出Restful的接口定义。

注意：
- 由于在GO中时间以纳秒为单位(`time.Duration`)，所以API需要从秒转为纳秒(1秒 = 1000000000纳秒)。
- defaultValue中的body需要将需要返回的内容转为BASE64，否则会遇到`base64.CorruptInputError=illegal base64 data`错误，这个是因为`type HTTPResult struct {Body []byte}`，详细解释见[[]byte encodes as a base64-encoded ](https://stackoverflow.com/questions/31449610/illegal-base64-data-error-message)。
- 在配置API的`renderTemplate`时，`flatAttrs`为true可以省略name，为false时name必须有值，这需要调用API时对数据进行校验，错误的配置会导致程序无法提供服务。
- Nodes中使用`defaultValue`时，格式应与`renderTemplate`中定义的抽取路径相符，否则会出现`Key path not found`错误。

## 枚举值
### Status
|名称|值|备注|
| -------------|:-------------:| -------------|
|Down|0||
|Up|1||

### CircuitStatus
|名称|值|备注|
| -------------|:-------------:| -------------|
|Open|0||
|Half|1||
|Close|2||

### LoadBalance
|名称|值|备注|
| -------------|:-------------:| -------------|
|RoundRobin|0||
|IPHash|1|目前版本不支持|

### Protocol
|名称|值|备注|
| -------------|:-------------:| -------------|
|HTTP|0||
|Grpc|1|目前版本不支持|
|Dubbo|1|目前版本不支持|
|SpringCloud|2|目前版本不支持|

### Source
|名称|值|备注|
| -------------|:-------------:| -------------|
|QueryString|0||
|FormData|1||
|JSONBody|2||
|Header|3||
|Cookie|4||
|PathValue|5||

### RuleType
|名称|值|备注|
| -------------|:-------------:| -------------|
|RuleRegexp|0||

### Source
|名称|值|备注|
| -------------|:-------------:| -------------|
|CMPEQ|0||
|CMPLT|1||
|CMPLE|2||
|CMPGT|3||
|CMPGE|4||
|CMPIn|5||
|CMPMatch|6|

### RoutingStrategy
|名称|值|备注|
| -------------|:-------------:| -------------|
|Copy|0||
|Split|1||

## Cluster
### 新增/更新
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
设置id字段表示更新

Reponse
```json
{
    "code":0,
    "data":1
}
```
data字段为cluster id

### 删除
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
### 查询
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
data字段为cluster

### 列表
|URL|Method|
| -------------|:-------------:|
|/v1/clusters?after=xx&limit=xx|GET|

after：上一次的最后一个cluster id
limit：获取多少条记录

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
data字段为cluster集合
取下一批: /v1/clusters?after=3&limit=3

### 查询所有bind的server
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
data字段为cluster bind的所有server id

### unbind所有server
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
### 新增/更新
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
设置id字段表示更新

Reponse
```json
{
    "code":0,
    "data":1
}
```
data字段为server id

### 删除
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
### 查询
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
data字段为server

### 列表
|URL|Method|
| -------------|:-------------:|
|/v1/servers?after=xx&limit=xx|GET|

after：上一次的最后一个server id
limit：获取多少条记录

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
data字段为server集合
取下一批: /v1/servers?after=3&limit=3

## Bind
### 增加
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

### 删除
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
### 新增/更新
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
    "position": 0
}
```
设置id字段表示更新

Reponse
```json
{
    "code":0,
    "data":1
}
```
data字段为API id

### 删除
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
### 查询
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
        "position": 0
    }
}
```
data字段为api

### 列表
|URL|Method|
| -------------|:-------------:|
|/v1/apis?after=xx&limit=xx|GET|

after：上一次的最后一个api id
limit：获取多少条记录

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
            "position": 0
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
            "position": 0
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
            "position": 0
        }
    ]
}
```
data字段为apis集合
取下一批: /v1/apis?after=3&limit=3

## Routing
### 新增/更新
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
设置id字段表示更新

Reponse
```json
{
    "code":0,
    "data":1
}
```
data字段为routing id

### 删除
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
### 查询
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
data字段为routing

### 列表
|URL|Method|
| -------------|:-------------:|
|/v1/routings?after=xx&limit=xx|GET|

after：上一次的最后一个routing id
limit：获取多少条记录

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
data字段为server集合
取下一批: /v1/routings?after=3&limit=3
