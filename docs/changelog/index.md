# CHANGELOG
---
title: core v0.1.0
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
highlight_theme: darkula
headingLevel: 2

---

<!-- Generator: Widdershins v4.0.1 -->

<h1 id="core">core v0.1.0</h1>

> Scroll down for code samples, example requests and responses. Select a language for code samples from the tabs above or the mobile navigation menu.

 License: 

<h1 id="core-cron-jobs">cron-jobs</h1>

Cron job management

## get_all_jobs

<a id="opIdget_all_jobs"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/cron-jobs \
  -H 'Accept: application/json'

```

```http
GET /api/v1/cron-jobs HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/cron-jobs',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/cron-jobs',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/cron-jobs', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/cron-jobs', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/cron-jobs");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/cron-jobs", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/cron-jobs`

> Example responses

> 200 Response

```json
[
  {
    "bash_script": "string",
    "command": "string",
    "created_at": "2019-08-24T14:15:22Z",
    "description": "string",
    "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
    "is_active": true,
    "last_run_at": "2019-08-24T14:15:22Z",
    "name": "string",
    "resource_limits": null,
    "schedule": "string",
    "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
    "updated_at": "2019-08-24T14:15:22Z",
    "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
  }
]
```

<h3 id="get_all_jobs-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|List of all cron jobs|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|None|

<h3 id="get_all_jobs-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|[[CronJob](#schemacronjob)]|false|none|none|
|» bash_script|string,null|false|none|none|
|» command|string|true|none|none|
|» created_at|string,null(date-time)|false|none|none|
|» description|string,null|false|none|none|
|» id|string(uuid)|true|none|none|
|» is_active|boolean|true|none|none|
|» last_run_at|string,null(date-time)|false|none|none|
|» name|string|true|none|none|
|» resource_limits|any|false|none|none|
|» schedule|string|true|none|none|
|» tenant_id|string(uuid)|true|none|none|
|» updated_at|string,null(date-time)|false|none|none|
|» user_id|string(uuid)|true|none|none|

<aside class="success">
This operation does not require authentication
</aside>

## create_job

<a id="opIdcreate_job"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/cron-jobs \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
POST /api/v1/cron-jobs HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "bash_script": "string",
  "command": "string",
  "description": "string",
  "is_active": true,
  "name": "string",
  "resource_limits": null,
  "schedule": "string",
  "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
  "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/cron-jobs',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/api/v1/cron-jobs',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post('/api/v1/cron-jobs', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/cron-jobs', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/cron-jobs");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/cron-jobs", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/cron-jobs`

> Body parameter

```json
{
  "bash_script": "string",
  "command": "string",
  "description": "string",
  "is_active": true,
  "name": "string",
  "resource_limits": null,
  "schedule": "string",
  "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
  "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
}
```

<h3 id="create_job-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[NewCronJob](#schemanewcronjob)|true|none|

> Example responses

> 201 Response

```json
{
  "bash_script": "string",
  "command": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "description": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "is_active": true,
  "last_run_at": "2019-08-24T14:15:22Z",
  "name": "string",
  "resource_limits": null,
  "schedule": "string",
  "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
  "updated_at": "2019-08-24T14:15:22Z",
  "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
}
```

<h3 id="create_job-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|Cron job created successfully|[CronJob](#schemacronjob)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|None|

<aside class="success">
This operation does not require authentication
</aside>

## get_job

<a id="opIdget_job"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/cron-jobs/{job_id} \
  -H 'Accept: application/json'

```

```http
GET /api/v1/cron-jobs/{job_id} HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/cron-jobs/{job_id}',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/cron-jobs/{job_id}',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/cron-jobs/{job_id}', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/cron-jobs/{job_id}', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/cron-jobs/{job_id}");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/cron-jobs/{job_id}", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/cron-jobs/{job_id}`

<h3 id="get_job-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|job_id|path|string(uuid)|true|Unique identifier of the cron job|

> Example responses

> 200 Response

```json
{
  "bash_script": "string",
  "command": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "description": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "is_active": true,
  "last_run_at": "2019-08-24T14:15:22Z",
  "name": "string",
  "resource_limits": null,
  "schedule": "string",
  "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
  "updated_at": "2019-08-24T14:15:22Z",
  "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
}
```

<h3 id="get_job-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Cron job found|[CronJob](#schemacronjob)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Cron job not found|None|

<aside class="success">
This operation does not require authentication
</aside>

## update_job

<a id="opIdupdate_job"></a>

> Code samples

```shell
# You can also use wget
curl -X PUT /api/v1/cron-jobs/{job_id} \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
PUT /api/v1/cron-jobs/{job_id} HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "bash_script": "string",
  "command": "string",
  "description": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "is_active": true,
  "name": "string",
  "resource_limits": null,
  "schedule": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/cron-jobs/{job_id}',
{
  method: 'PUT',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.put '/api/v1/cron-jobs/{job_id}',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.put('/api/v1/cron-jobs/{job_id}', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('PUT','/api/v1/cron-jobs/{job_id}', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/cron-jobs/{job_id}");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("PUT");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("PUT", "/api/v1/cron-jobs/{job_id}", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`PUT /api/v1/cron-jobs/{job_id}`

> Body parameter

```json
{
  "bash_script": "string",
  "command": "string",
  "description": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "is_active": true,
  "name": "string",
  "resource_limits": null,
  "schedule": "string"
}
```

<h3 id="update_job-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|job_id|path|string(uuid)|true|Unique identifier of the cron job to update|
|body|body|[UpdateCronJob](#schemaupdatecronjob)|true|none|

> Example responses

> 200 Response

```json
{
  "bash_script": "string",
  "command": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "description": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "is_active": true,
  "last_run_at": "2019-08-24T14:15:22Z",
  "name": "string",
  "resource_limits": null,
  "schedule": "string",
  "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
  "updated_at": "2019-08-24T14:15:22Z",
  "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
}
```

<h3 id="update_job-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Cron job updated successfully|[CronJob](#schemacronjob)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|None|

<aside class="success">
This operation does not require authentication
</aside>

## delete_job

<a id="opIddelete_job"></a>

> Code samples

```shell
# You can also use wget
curl -X DELETE /api/v1/cron-jobs/{job_id}

```

```http
DELETE /api/v1/cron-jobs/{job_id} HTTP/1.1

```

```javascript

fetch('/api/v1/cron-jobs/{job_id}',
{
  method: 'DELETE'

})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

result = RestClient.delete '/api/v1/cron-jobs/{job_id}',
  params: {
  }

p JSON.parse(result)

```

```python
import requests

r = requests.delete('/api/v1/cron-jobs/{job_id}')

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('DELETE','/api/v1/cron-jobs/{job_id}', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/cron-jobs/{job_id}");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("DELETE");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("DELETE", "/api/v1/cron-jobs/{job_id}", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`DELETE /api/v1/cron-jobs/{job_id}`

<h3 id="delete_job-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|job_id|path|string(uuid)|true|Unique identifier of the cron job to delete|

<h3 id="delete_job-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|204|[No Content](https://tools.ietf.org/html/rfc7231#section-6.3.5)|Cron job deleted successfully|None|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|None|

<aside class="success">
This operation does not require authentication
</aside>

## get_logs_of_cron_job_by_id

<a id="opIdget_logs_of_cron_job_by_id"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/cron-jobs/{job_id}/logs \
  -H 'Accept: application/json'

```

```http
GET /api/v1/cron-jobs/{job_id}/logs HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/cron-jobs/{job_id}/logs',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/cron-jobs/{job_id}/logs',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/cron-jobs/{job_id}/logs', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/cron-jobs/{job_id}/logs', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/cron-jobs/{job_id}/logs");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/cron-jobs/{job_id}/logs", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/cron-jobs/{job_id}/logs`

<h3 id="get_logs_of_cron_job_by_id-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|job_id|path|string(uuid)|true|Unique identifier of the cron job to get logs for|

> Example responses

> 200 Response

```json
[
  {
    "created_at": "2019-08-24T14:15:22Z",
    "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
    "level": "string",
    "message": "string",
    "topic": "string"
  }
]
```

<h3 id="get_logs_of_cron_job_by_id-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|List of logs for the specified cron job|Inline|

<h3 id="get_logs_of_cron_job_by_id-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|[[Log](#schemalog)]|false|none|none|
|» created_at|string,null(date-time)|false|none|none|
|» id|string(uuid)|true|none|none|
|» level|string|true|none|none|
|» message|string|true|none|none|
|» topic|string,null|false|none|none|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="core-file-management">file-management</h1>

File management

## create_file

<a id="opIdcreate_file"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/files \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
POST /api/v1/files HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "name": "string",
  "path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/api/v1/files',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post('/api/v1/files', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/files', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/files", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/files`

> Body parameter

```json
{
  "name": "string",
  "path": "string"
}
```

<h3 id="create_file-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[CreateDirectoryRequest](#schemacreatedirectoryrequest)|true|none|

> Example responses

> 201 Response

```json
{
  "message": "string",
  "success": true
}
```

<h3 id="create_file-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|File created successfully|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## delete_file

<a id="opIddelete_file"></a>

> Code samples

```shell
# You can also use wget
curl -X DELETE /api/v1/files \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
DELETE /api/v1/files HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files',
{
  method: 'DELETE',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.delete '/api/v1/files',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.delete('/api/v1/files', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('DELETE','/api/v1/files', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("DELETE");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("DELETE", "/api/v1/files", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`DELETE /api/v1/files`

> Body parameter

```json
{
  "path": "string"
}
```

<h3 id="delete_file-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[DeleteFileRequest](#schemadeletefilerequest)|true|none|

> Example responses

> 200 Response

```json
{
  "message": "string",
  "success": true
}
```

<h3 id="delete_file-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|File deleted successfully|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## copy_folder_and_files_recursively

<a id="opIdcopy_folder_and_files_recursively"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/files/copy \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
POST /api/v1/files/copy HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "from_path": "string",
  "to_path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files/copy',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/api/v1/files/copy',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post('/api/v1/files/copy', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/files/copy', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/copy");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/files/copy", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/files/copy`

> Body parameter

```json
{
  "from_path": "string",
  "to_path": "string"
}
```

<h3 id="copy_folder_and_files_recursively-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[MoveFileRequest](#schemamovefilerequest)|true|none|

> Example responses

> 200 Response

```json
{
  "message": "string",
  "success": true
}
```

<h3 id="copy_folder_and_files_recursively-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Folder and files copied successfully|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## create_directory

<a id="opIdcreate_directory"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/files/directories \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
POST /api/v1/files/directories HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "name": "string",
  "path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files/directories',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/api/v1/files/directories',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post('/api/v1/files/directories', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/files/directories', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/directories");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/files/directories", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/files/directories`

> Body parameter

```json
{
  "name": "string",
  "path": "string"
}
```

<h3 id="create_directory-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[CreateDirectoryRequest](#schemacreatedirectoryrequest)|true|none|

> Example responses

> 201 Response

```json
{
  "message": "string",
  "success": true
}
```

<h3 id="create_directory-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|Directory created successfully|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|Directory already exists|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## delete_directory

<a id="opIddelete_directory"></a>

> Code samples

```shell
# You can also use wget
curl -X DELETE /api/v1/files/directories \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
DELETE /api/v1/files/directories HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files/directories',
{
  method: 'DELETE',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.delete '/api/v1/files/directories',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.delete('/api/v1/files/directories', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('DELETE','/api/v1/files/directories', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/directories");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("DELETE");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("DELETE", "/api/v1/files/directories", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`DELETE /api/v1/files/directories`

> Body parameter

```json
{
  "path": "string"
}
```

<h3 id="delete_directory-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[DeleteFileRequest](#schemadeletefilerequest)|true|none|

> Example responses

> 200 Response

```json
{
  "message": "string",
  "success": true
}
```

<h3 id="delete_directory-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Directory deleted successfully|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## calculate_directory_size

<a id="opIdcalculate_directory_size"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/files/directories/size \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
POST /api/v1/files/directories/size HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files/directories/size',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/api/v1/files/directories/size',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post('/api/v1/files/directories/size', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/files/directories/size', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/directories/size");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/files/directories/size", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/files/directories/size`

> Body parameter

```json
{
  "path": "string"
}
```

<h3 id="calculate_directory_size-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[DeleteFileRequest](#schemadeletefilerequest)|true|none|

> Example responses

> 200 Response

```json
{
  "data": 0,
  "message": "string",
  "success": true
}
```

<h3 id="calculate_directory_size-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Directory size calculated successfully|[FileSizeResponse](#schemafilesizeresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## create_directory_with_parent

<a id="opIdcreate_directory_with_parent"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/files/directories/with-parent \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
POST /api/v1/files/directories/with-parent HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "name": "string",
  "path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files/directories/with-parent',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/api/v1/files/directories/with-parent',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post('/api/v1/files/directories/with-parent', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/files/directories/with-parent', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/directories/with-parent");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/files/directories/with-parent", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/files/directories/with-parent`

> Body parameter

```json
{
  "name": "string",
  "path": "string"
}
```

<h3 id="create_directory_with_parent-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[CreateDirectoryRequest](#schemacreatedirectoryrequest)|true|none|

> Example responses

> 201 Response

```json
{
  "message": "string",
  "success": true
}
```

<h3 id="create_directory_with_parent-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|Directory created successfully|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## calculate_disk_usage

<a id="opIdcalculate_disk_usage"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/files/disk-usage \
  -H 'Accept: application/json'

```

```http
GET /api/v1/files/disk-usage HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/files/disk-usage',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/files/disk-usage',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/files/disk-usage', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/files/disk-usage', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/disk-usage");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/files/disk-usage", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/files/disk-usage`

> Example responses

> 200 Response

```json
{
  "data": {},
  "message": "string",
  "success": true
}
```

<h3 id="calculate_disk_usage-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Disk usage calculated successfully|[DiskUsageResponse](#schemadiskusageresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## list_files_and_directories_in_path

<a id="opIdlist_files_and_directories_in_path"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/files/list?path=string \
  -H 'Accept: application/json'

```

```http
GET /api/v1/files/list?path=string HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/files/list?path=string',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/files/list',
  params: {
  'path' => 'string'
}, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/files/list', params={
  'path': 'string'
}, headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/files/list', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/list?path=string");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/files/list", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/files/list`

<h3 id="list_files_and_directories_in_path-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|path|query|string|true|Path to list files and directories from|

> Example responses

> 200 Response

```json
{
  "data": [
    {
      "created_at": "2019-08-24T14:15:22Z",
      "extension": "string",
      "file_type": "File",
      "group_id": 0,
      "is_hidden": true,
      "name": "string",
      "owner_id": 0,
      "path": "string",
      "permissions": 0,
      "size": 0,
      "updated_at": "2019-08-24T14:15:22Z"
    }
  ],
  "message": "string",
  "success": true
}
```

<h3 id="list_files_and_directories_in_path-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Files and directories listed successfully|[FileListResponse](#schemafilelistresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[FileListResponse](#schemafilelistresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## calculate_memory_usage

<a id="opIdcalculate_memory_usage"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/files/memory-usage \
  -H 'Accept: application/json'

```

```http
GET /api/v1/files/memory-usage HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/files/memory-usage',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/files/memory-usage',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/files/memory-usage', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/files/memory-usage', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/memory-usage");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/files/memory-usage", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/files/memory-usage`

> Example responses

> 200 Response

```json
{
  "data": {},
  "message": "string",
  "success": true
}
```

<h3 id="calculate_memory_usage-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Memory usage calculated successfully|[MemoryUsageResponse](#schemamemoryusageresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

## move_folder_and_files_recursively

<a id="opIdmove_folder_and_files_recursively"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/files/move \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json'

```

```http
POST /api/v1/files/move HTTP/1.1

Content-Type: application/json
Accept: application/json

```

```javascript
const inputBody = '{
  "from_path": "string",
  "to_path": "string"
}';
const headers = {
  'Content-Type':'application/json',
  'Accept':'application/json'
};

fetch('/api/v1/files/move',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/api/v1/files/move',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post('/api/v1/files/move', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/files/move', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/files/move");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/files/move", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/files/move`

> Body parameter

```json
{
  "from_path": "string",
  "to_path": "string"
}
```

<h3 id="move_folder_and_files_recursively-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[MoveFileRequest](#schemamovefilerequest)|true|none|

> Example responses

> 200 Response

```json
{
  "message": "string",
  "success": true
}
```

<h3 id="move_folder_and_files_recursively-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Folder and files moved successfully|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Internal server error|[CreateDirectoryResponse](#schemacreatedirectoryresponse)|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="core-marketplace">marketplace</h1>

Application marketplace

## list_applications

<a id="opIdlist_applications"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/marketplace/applications \
  -H 'Accept: application/json'

```

```http
GET /api/v1/marketplace/applications HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/marketplace/applications',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/marketplace/applications',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/marketplace/applications', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/marketplace/applications', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/marketplace/applications");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/marketplace/applications", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/marketplace/applications`

> Example responses

> 200 Response

```json
[
  {
    "alternatives": "string",
    "app_type": "string",
    "created_at": "2019-08-24T14:15:22Z",
    "description": "string",
    "icon": "string",
    "icon_type": "string",
    "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
    "license": "string",
    "name": "string",
    "reference": "string",
    "repository": "string",
    "repository_link": "string",
    "source": "string",
    "stars": 0,
    "stars_display": "string",
    "tags": [
      "string"
    ],
    "website": "string"
  }
]
```

<h3 id="list_applications-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|List of applications|Inline|

<h3 id="list_applications-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|[[Application](#schemaapplication)]|false|none|none|
|» alternatives|string,null|false|none|none|
|» app_type|string|true|none|none|
|» created_at|string,null(date-time)|false|none|none|
|» description|string,null|false|none|none|
|» icon|string,null|false|none|none|
|» icon_type|string,null|false|none|none|
|» id|string(uuid)|true|none|none|
|» license|string,null|false|none|none|
|» name|string|true|none|none|
|» reference|string|true|none|none|
|» repository|string,null|false|none|none|
|» repository_link|string,null|false|none|none|
|» source|string,null|false|none|none|
|» stars|integer(int32)|true|none|none|
|» stars_display|string|true|none|none|
|» tags|[string]|true|none|none|
|» website|string,null|false|none|none|

<aside class="success">
This operation does not require authentication
</aside>

## install_application

<a id="opIdinstall_application"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/marketplace/applications \
  -H 'Content-Type: application/json'

```

```http
POST /api/v1/marketplace/applications HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = '{
  "app_id": "string",
  "app_name": "string",
  "installation_id": "string"
}';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/marketplace/applications',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.post '/api/v1/marketplace/applications',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.post('/api/v1/marketplace/applications', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/marketplace/applications', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/marketplace/applications");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/marketplace/applications", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/marketplace/applications`

> Body parameter

```json
{
  "app_id": "string",
  "app_name": "string",
  "installation_id": "string"
}
```

<h3 id="install_application-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[InstallApplication](#schemainstallapplication)|true|none|

<h3 id="install_application-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application installed successfully|None|

<aside class="success">
This operation does not require authentication
</aside>

## uninstall_application

<a id="opIduninstall_application"></a>

> Code samples

```shell
# You can also use wget
curl -X DELETE /api/v1/marketplace/applications \
  -H 'Content-Type: application/json'

```

```http
DELETE /api/v1/marketplace/applications HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = '{
  "app_id": "string",
  "app_name": "string",
  "installation_id": "string"
}';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/marketplace/applications',
{
  method: 'DELETE',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.delete '/api/v1/marketplace/applications',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.delete('/api/v1/marketplace/applications', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('DELETE','/api/v1/marketplace/applications', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/marketplace/applications");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("DELETE");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("DELETE", "/api/v1/marketplace/applications", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`DELETE /api/v1/marketplace/applications`

> Body parameter

```json
{
  "app_id": "string",
  "app_name": "string",
  "installation_id": "string"
}
```

<h3 id="uninstall_application-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[InstallApplication](#schemainstallapplication)|true|none|

<h3 id="uninstall_application-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application uninstalled successfully|None|

<aside class="success">
This operation does not require authentication
</aside>

## get_application_by_name

<a id="opIdget_application_by_name"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/marketplace/applications/by-reference?app_id=string \
  -H 'Accept: application/json'

```

```http
GET /api/v1/marketplace/applications/by-reference?app_id=string HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/marketplace/applications/by-reference?app_id=string',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/marketplace/applications/by-reference',
  params: {
  'app_id' => 'string'
}, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/marketplace/applications/by-reference', params={
  'app_id': 'string'
}, headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/marketplace/applications/by-reference', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/marketplace/applications/by-reference?app_id=string");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/marketplace/applications/by-reference", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/marketplace/applications/by-reference`

<h3 id="get_application_by_name-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|app_id|query|string|true|Application ID|

> Example responses

> 200 Response

```json
{
  "alternatives": "string",
  "app_type": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "description": "string",
  "icon": "string",
  "icon_type": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "license": "string",
  "name": "string",
  "reference": "string",
  "repository": "string",
  "repository_link": "string",
  "source": "string",
  "stars": 0,
  "stars_display": "string",
  "tags": [
    "string"
  ],
  "website": "string"
}
```

<h3 id="get_application_by_name-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application details|[Application](#schemaapplication)|

<aside class="success">
This operation does not require authentication
</aside>

## store_application_details

<a id="opIdstore_application_details"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/marketplace/applications/store-all

```

```http
GET /api/v1/marketplace/applications/store-all HTTP/1.1

```

```javascript

fetch('/api/v1/marketplace/applications/store-all',
{
  method: 'GET'

})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

result = RestClient.get '/api/v1/marketplace/applications/store-all',
  params: {
  }

p JSON.parse(result)

```

```python
import requests

r = requests.get('/api/v1/marketplace/applications/store-all')

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/marketplace/applications/store-all', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/marketplace/applications/store-all");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/marketplace/applications/store-all", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/marketplace/applications/store-all`

<h3 id="store_application_details-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application details stored successfully|None|

<aside class="success">
This operation does not require authentication
</aside>

## get_application_logs

<a id="opIdget_application_logs"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/marketplace/applications/{app_id}/logs \
  -H 'Accept: application/json'

```

```http
GET /api/v1/marketplace/applications/{app_id}/logs HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/marketplace/applications/{app_id}/logs',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/marketplace/applications/{app_id}/logs',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/marketplace/applications/{app_id}/logs', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/marketplace/applications/{app_id}/logs', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/marketplace/applications/{app_id}/logs");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/marketplace/applications/{app_id}/logs", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/marketplace/applications/{app_id}/logs`

<h3 id="get_application_logs-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|app_id|path|string|true|Application ID|

> Example responses

> 200 Response

```json
[
  {
    "created_at": "2019-08-24T14:15:22Z",
    "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
    "level": "string",
    "message": "string",
    "topic": "string"
  }
]
```

<h3 id="get_application_logs-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application logs|Inline|

<h3 id="get_application_logs-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|[[Log](#schemalog)]|false|none|none|
|» created_at|string,null(date-time)|false|none|none|
|» id|string(uuid)|true|none|none|
|» level|string|true|none|none|
|» message|string|true|none|none|
|» topic|string,null|false|none|none|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="core-self-host">self-host</h1>

Self-hosting operations

## list_all_hosted_applications

<a id="opIdlist_all_hosted_applications"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/self-host/applications \
  -H 'Accept: application/json'

```

```http
GET /api/v1/self-host/applications HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/self-host/applications',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/self-host/applications',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/self-host/applications', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/self-host/applications', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/self-host/applications", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/self-host/applications`

> Example responses

> 200 Response

```json
[
  null
]
```

<h3 id="list_all_hosted_applications-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|List of all hosted applications|Inline|

<h3 id="list_all_hosted_applications-responseschema">Response Schema</h3>

<aside class="success">
This operation does not require authentication
</aside>

## install_application

<a id="opIdinstall_application"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/self-host/applications \
  -H 'Content-Type: application/json'

```

```http
POST /api/v1/self-host/applications HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = '{
  "application_name": "string",
  "build_pack": "Dockerfile",
  "build_variables": "string",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "Dev",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string"
}';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/self-host/applications',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.post '/api/v1/self-host/applications',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.post('/api/v1/self-host/applications', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/self-host/applications', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/self-host/applications", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/self-host/applications`

> Body parameter

```json
{
  "application_name": "string",
  "build_pack": "Dockerfile",
  "build_variables": "string",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "Dev",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string"
}
```

<h3 id="install_application-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[NewApplicationRequest](#schemanewapplicationrequest)|true|none|

<h3 id="install_application-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application installed successfully|None|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to install application|None|

<aside class="success">
This operation does not require authentication
</aside>

## force_deploy

<a id="opIdforce_deploy"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/self-host/applications/deploy \
  -H 'Content-Type: application/json'

```

```http
POST /api/v1/self-host/applications/deploy HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = '{
  "application_name": "string",
  "build_pack": "string",
  "build_variables": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string",
  "updated_at": "2019-08-24T14:15:22Z"
}';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/self-host/applications/deploy',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.post '/api/v1/self-host/applications/deploy',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.post('/api/v1/self-host/applications/deploy', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/self-host/applications/deploy', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications/deploy");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/self-host/applications/deploy", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/self-host/applications/deploy`

> Body parameter

```json
{
  "application_name": "string",
  "build_pack": "string",
  "build_variables": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string",
  "updated_at": "2019-08-24T14:15:22Z"
}
```

<h3 id="force_deploy-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[UpdateApplicationForDatabase](#schemaupdateapplicationfordatabase)|true|none|

<h3 id="force_deploy-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application deployed successfully|None|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to deploy application|None|

<aside class="success">
This operation does not require authentication
</aside>

## force_deploy_without_cache

<a id="opIdforce_deploy_without_cache"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/self-host/applications/deploy-without-cache \
  -H 'Content-Type: application/json'

```

```http
POST /api/v1/self-host/applications/deploy-without-cache HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = '{
  "application_name": "string",
  "build_pack": "string",
  "build_variables": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string",
  "updated_at": "2019-08-24T14:15:22Z"
}';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/self-host/applications/deploy-without-cache',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.post '/api/v1/self-host/applications/deploy-without-cache',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.post('/api/v1/self-host/applications/deploy-without-cache', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/self-host/applications/deploy-without-cache', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications/deploy-without-cache");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/self-host/applications/deploy-without-cache", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/self-host/applications/deploy-without-cache`

> Body parameter

```json
{
  "application_name": "string",
  "build_pack": "string",
  "build_variables": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string",
  "updated_at": "2019-08-24T14:15:22Z"
}
```

<h3 id="force_deploy_without_cache-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[UpdateApplicationForDatabase](#schemaupdateapplicationfordatabase)|true|none|

<h3 id="force_deploy_without_cache-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application deployed successfully without cache|None|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to deploy application|None|

<aside class="success">
This operation does not require authentication
</aside>

## get_application_by_id

<a id="opIdget_application_by_id"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/self-host/applications/{id} \
  -H 'Accept: application/json'

```

```http
GET /api/v1/self-host/applications/{id} HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/self-host/applications/{id}',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/self-host/applications/{id}',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/self-host/applications/{id}', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/self-host/applications/{id}', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications/{id}");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/self-host/applications/{id}", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/self-host/applications/{id}`

<h3 id="get_application_by_id-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|id|path|string|true|Application ID|

> Example responses

> 200 Response

```json
null
```

<h3 id="get_application_by_id-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application details|Inline|

<h3 id="get_application_by_id-responseschema">Response Schema</h3>

<aside class="success">
This operation does not require authentication
</aside>

## update_application_details

<a id="opIdupdate_application_details"></a>

> Code samples

```shell
# You can also use wget
curl -X PUT /api/v1/self-host/applications/{id} \
  -H 'Content-Type: application/json'

```

```http
PUT /api/v1/self-host/applications/{id} HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = '{
  "application_name": "string",
  "build_pack": "string",
  "build_variables": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string",
  "updated_at": "2019-08-24T14:15:22Z"
}';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/self-host/applications/{id}',
{
  method: 'PUT',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.put '/api/v1/self-host/applications/{id}',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.put('/api/v1/self-host/applications/{id}', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('PUT','/api/v1/self-host/applications/{id}', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications/{id}");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("PUT");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("PUT", "/api/v1/self-host/applications/{id}", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`PUT /api/v1/self-host/applications/{id}`

> Body parameter

```json
{
  "application_name": "string",
  "build_pack": "string",
  "build_variables": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string",
  "updated_at": "2019-08-24T14:15:22Z"
}
```

<h3 id="update_application_details-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[UpdateApplicationForDatabase](#schemaupdateapplicationfordatabase)|true|none|

<h3 id="update_application_details-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application details updated successfully|None|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to update application|None|

<aside class="success">
This operation does not require authentication
</aside>

## delete_application_by_id

<a id="opIddelete_application_by_id"></a>

> Code samples

```shell
# You can also use wget
curl -X DELETE /api/v1/self-host/applications/{id} \
  -H 'Content-Type: application/json'

```

```http
DELETE /api/v1/self-host/applications/{id} HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = '{
  "environment": "Dev",
  "id": "string"
}';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/self-host/applications/{id}',
{
  method: 'DELETE',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.delete '/api/v1/self-host/applications/{id}',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.delete('/api/v1/self-host/applications/{id}', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('DELETE','/api/v1/self-host/applications/{id}', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications/{id}");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("DELETE");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("DELETE", "/api/v1/self-host/applications/{id}", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`DELETE /api/v1/self-host/applications/{id}`

> Body parameter

```json
{
  "environment": "Dev",
  "id": "string"
}
```

<h3 id="delete_application_by_id-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[StopApplicationRequest](#schemastopapplicationrequest)|true|none|

<h3 id="delete_application_by_id-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application deleted successfully|None|

<aside class="success">
This operation does not require authentication
</aside>

## get_logs_of_self_host_application

<a id="opIdget_logs_of_self_host_application"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/self-host/applications/{id}/logs?container_id=string \
  -H 'Accept: application/json'

```

```http
GET /api/v1/self-host/applications/{id}/logs?container_id=string HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/self-host/applications/{id}/logs?container_id=string',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/self-host/applications/{id}/logs',
  params: {
  'container_id' => 'string'
}, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/self-host/applications/{id}/logs', params={
  'container_id': 'string'
}, headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/self-host/applications/{id}/logs', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/applications/{id}/logs?container_id=string");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/self-host/applications/{id}/logs", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/self-host/applications/{id}/logs`

<h3 id="get_logs_of_self_host_application-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|id|path|string|true|Application ID|
|container_id|query|string|true|Container ID|

> Example responses

> 200 Response

```json
[
  null
]
```

<h3 id="get_logs_of_self_host_application-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Application logs|Inline|

<h3 id="get_logs_of_self_host_application-responseschema">Response Schema</h3>

<aside class="success">
This operation does not require authentication
</aside>

## get_image_build_history

<a id="opIdget_image_build_history"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/self-host/build-history \
  -H 'Accept: application/json'

```

```http
GET /api/v1/self-host/build-history HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/self-host/build-history',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/self-host/build-history',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/self-host/build-history', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/self-host/build-history', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/build-history");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/self-host/build-history", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/self-host/build-history`

<h3 id="get_image_build_history-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|image_name|path|string|true|none|
|environment|path|[Environment](#schemaenvironment)|true|none|
|id|path|string|true|none|

#### Enumerated Values

|Parameter|Value|
|---|---|
|environment|Dev|
|environment|Staging|
|environment|Prod|

> Example responses

> 200 Response

```json
[
  null
]
```

<h3 id="get_image_build_history-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Image build history|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to get image build history|None|

<h3 id="get_image_build_history-responseschema">Response Schema</h3>

<aside class="success">
This operation does not require authentication
</aside>

## list_containers_by_image

<a id="opIdlist_containers_by_image"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/self-host/containers

```

```http
GET /api/v1/self-host/containers HTTP/1.1

```

```javascript

fetch('/api/v1/self-host/containers',
{
  method: 'GET'

})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

result = RestClient.get '/api/v1/self-host/containers',
  params: {
  }

p JSON.parse(result)

```

```python
import requests

r = requests.get('/api/v1/self-host/containers')

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/self-host/containers', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/containers");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/self-host/containers", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/self-host/containers`

<h3 id="list_containers_by_image-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|image_name|path|string|true|none|
|environment|path|[Environment](#schemaenvironment)|true|none|
|id|path|string|true|none|

#### Enumerated Values

|Parameter|Value|
|---|---|
|environment|Dev|
|environment|Staging|
|environment|Prod|

<h3 id="list_containers_by_image-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|List of containers|None|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to list containers|None|

<aside class="success">
This operation does not require authentication
</aside>

## get_user_repositories

<a id="opIdget_user_repositories"></a>

> Code samples

```shell
# You can also use wget
curl -X GET /api/v1/self-host/repositories?installation_id=0 \
  -H 'Accept: application/json'

```

```http
GET /api/v1/self-host/repositories?installation_id=0 HTTP/1.1

Accept: application/json

```

```javascript

const headers = {
  'Accept':'application/json'
};

fetch('/api/v1/self-host/repositories?installation_id=0',
{
  method: 'GET',

  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/api/v1/self-host/repositories',
  params: {
  'installation_id' => 'integer(int64)'
}, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Accept': 'application/json'
}

r = requests.get('/api/v1/self-host/repositories', params={
  'installation_id': '0'
}, headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Accept' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('GET','/api/v1/self-host/repositories', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/repositories?installation_id=0");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("GET", "/api/v1/self-host/repositories", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`GET /api/v1/self-host/repositories`

<h3 id="get_user_repositories-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|installation_id|query|integer(int64)|true|GitHub App installation ID|

> Example responses

> 200 Response

```json
[
  null
]
```

<h3 id="get_user_repositories-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|List of user repositories|Inline|

<h3 id="get_user_repositories-responseschema">Response Schema</h3>

<aside class="success">
This operation does not require authentication
</aside>

## github_webhook

<a id="opIdgithub_webhook"></a>

> Code samples

```shell
# You can also use wget
curl -X POST /api/v1/self-host/webhook \
  -H 'Content-Type: application/json'

```

```http
POST /api/v1/self-host/webhook HTTP/1.1

Content-Type: application/json

```

```javascript
const inputBody = 'string';
const headers = {
  'Content-Type':'application/json'
};

fetch('/api/v1/self-host/webhook',
{
  method: 'POST',
  body: inputBody,
  headers: headers
})
.then(function(res) {
    return res.json();
}).then(function(body) {
    console.log(body);
});

```

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json'
}

result = RestClient.post '/api/v1/self-host/webhook',
  params: {
  }, headers: headers

p JSON.parse(result)

```

```python
import requests
headers = {
  'Content-Type': 'application/json'
}

r = requests.post('/api/v1/self-host/webhook', headers = headers)

print(r.json())

```

```php
<?php

require 'vendor/autoload.php';

$headers = array(
    'Content-Type' => 'application/json',
);

$client = new \GuzzleHttp\Client();

// Define array of request body.
$request_body = array();

try {
    $response = $client->request('POST','/api/v1/self-host/webhook', array(
        'headers' => $headers,
        'json' => $request_body,
       )
    );
    print_r($response->getBody()->getContents());
 }
 catch (\GuzzleHttp\Exception\BadResponseException $e) {
    // handle exception or api errors.
    print_r($e->getMessage());
 }

 // ...

```

```java
URL obj = new URL("/api/v1/self-host/webhook");
HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");
int responseCode = con.getResponseCode();
BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream()));
String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();
System.out.println(response.toString());

```

```go
package main

import (
       "bytes"
       "net/http"
)

func main() {

    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
    }

    data := bytes.NewBuffer([]byte{jsonReq})
    req, err := http.NewRequest("POST", "/api/v1/self-host/webhook", data)
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}

```

`POST /api/v1/self-host/webhook`

> Body parameter

```json
"string"
```

<h3 id="github_webhook-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|string|true|GitHub webhook payload|

<h3 id="github_webhook-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Webhook received successfully|None|

<aside class="success">
This operation does not require authentication
</aside>

# Schemas

<h2 id="tocS_Application">Application</h2>
<!-- backwards compatibility -->
<a id="schemaapplication"></a>
<a id="schema_Application"></a>
<a id="tocSapplication"></a>
<a id="tocsapplication"></a>

```json
{
  "alternatives": "string",
  "app_type": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "description": "string",
  "icon": "string",
  "icon_type": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "license": "string",
  "name": "string",
  "reference": "string",
  "repository": "string",
  "repository_link": "string",
  "source": "string",
  "stars": 0,
  "stars_display": "string",
  "tags": [
    "string"
  ],
  "website": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|alternatives|string,null|false|none|none|
|app_type|string|true|none|none|
|created_at|string,null(date-time)|false|none|none|
|description|string,null|false|none|none|
|icon|string,null|false|none|none|
|icon_type|string,null|false|none|none|
|id|string(uuid)|true|none|none|
|license|string,null|false|none|none|
|name|string|true|none|none|
|reference|string|true|none|none|
|repository|string,null|false|none|none|
|repository_link|string,null|false|none|none|
|source|string,null|false|none|none|
|stars|integer(int32)|true|none|none|
|stars_display|string|true|none|none|
|tags|[string]|true|none|none|
|website|string,null|false|none|none|

<h2 id="tocS_ApplicationLogsRequest">ApplicationLogsRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapplicationlogsrequest"></a>
<a id="schema_ApplicationLogsRequest"></a>
<a id="tocSapplicationlogsrequest"></a>
<a id="tocsapplicationlogsrequest"></a>

```json
{
  "container_id": "string",
  "id": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|container_id|string|true|none|none|
|id|string|true|none|none|

<h2 id="tocS_BuildPack">BuildPack</h2>
<!-- backwards compatibility -->
<a id="schemabuildpack"></a>
<a id="schema_BuildPack"></a>
<a id="tocSbuildpack"></a>
<a id="tocsbuildpack"></a>

```json
"Dockerfile"

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|string|false|none|none|

#### Enumerated Values

|Property|Value|
|---|---|
|*anonymous*|Dockerfile|
|*anonymous*|DockerCompose|
|*anonymous*|Static|

<h2 id="tocS_CreateDirectoryRequest">CreateDirectoryRequest</h2>
<!-- backwards compatibility -->
<a id="schemacreatedirectoryrequest"></a>
<a id="schema_CreateDirectoryRequest"></a>
<a id="tocScreatedirectoryrequest"></a>
<a id="tocscreatedirectoryrequest"></a>

```json
{
  "name": "string",
  "path": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|name|string|true|none|none|
|path|string|true|none|none|

<h2 id="tocS_CreateDirectoryResponse">CreateDirectoryResponse</h2>
<!-- backwards compatibility -->
<a id="schemacreatedirectoryresponse"></a>
<a id="schema_CreateDirectoryResponse"></a>
<a id="tocScreatedirectoryresponse"></a>
<a id="tocscreatedirectoryresponse"></a>

```json
{
  "message": "string",
  "success": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|message|string|true|none|none|
|success|boolean|true|none|none|

<h2 id="tocS_CronJob">CronJob</h2>
<!-- backwards compatibility -->
<a id="schemacronjob"></a>
<a id="schema_CronJob"></a>
<a id="tocScronjob"></a>
<a id="tocscronjob"></a>

```json
{
  "bash_script": "string",
  "command": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "description": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "is_active": true,
  "last_run_at": "2019-08-24T14:15:22Z",
  "name": "string",
  "resource_limits": null,
  "schedule": "string",
  "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
  "updated_at": "2019-08-24T14:15:22Z",
  "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|bash_script|string,null|false|none|none|
|command|string|true|none|none|
|created_at|string,null(date-time)|false|none|none|
|description|string,null|false|none|none|
|id|string(uuid)|true|none|none|
|is_active|boolean|true|none|none|
|last_run_at|string,null(date-time)|false|none|none|
|name|string|true|none|none|
|resource_limits|any|false|none|none|
|schedule|string|true|none|none|
|tenant_id|string(uuid)|true|none|none|
|updated_at|string,null(date-time)|false|none|none|
|user_id|string(uuid)|true|none|none|

<h2 id="tocS_DeleteFileRequest">DeleteFileRequest</h2>
<!-- backwards compatibility -->
<a id="schemadeletefilerequest"></a>
<a id="schema_DeleteFileRequest"></a>
<a id="tocSdeletefilerequest"></a>
<a id="tocsdeletefilerequest"></a>

```json
{
  "path": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|path|string|true|none|none|

<h2 id="tocS_DiskUsageData">DiskUsageData</h2>
<!-- backwards compatibility -->
<a id="schemadiskusagedata"></a>
<a id="schema_DiskUsageData"></a>
<a id="tocSdiskusagedata"></a>
<a id="tocsdiskusagedata"></a>

```json
{
  "available": "string",
  "capacity": "string",
  "filesystem": "string",
  "ifree": "string",
  "iused": "string",
  "iused_percentage": "string",
  "mounted_on": "string",
  "size": "string",
  "used": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|available|string|true|none|none|
|capacity|string|true|none|none|
|filesystem|string|true|none|none|
|ifree|string|true|none|none|
|iused|string|true|none|none|
|iused_percentage|string|true|none|none|
|mounted_on|string|true|none|none|
|size|string|true|none|none|
|used|string|true|none|none|

<h2 id="tocS_DiskUsageResponse">DiskUsageResponse</h2>
<!-- backwards compatibility -->
<a id="schemadiskusageresponse"></a>
<a id="schema_DiskUsageResponse"></a>
<a id="tocSdiskusageresponse"></a>
<a id="tocsdiskusageresponse"></a>

```json
{
  "data": {},
  "message": "string",
  "success": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|data|any|false|none|none|

oneOf

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» *anonymous*|null|false|none|none|

xor

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» *anonymous*|[DiskUsageData](#schemadiskusagedata)|false|none|none|

continued

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|message|string|true|none|none|
|success|boolean|true|none|none|

<h2 id="tocS_Environment">Environment</h2>
<!-- backwards compatibility -->
<a id="schemaenvironment"></a>
<a id="schema_Environment"></a>
<a id="tocSenvironment"></a>
<a id="tocsenvironment"></a>

```json
"Dev"

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|string|false|none|none|

#### Enumerated Values

|Property|Value|
|---|---|
|*anonymous*|Dev|
|*anonymous*|Staging|
|*anonymous*|Prod|

<h2 id="tocS_File">File</h2>
<!-- backwards compatibility -->
<a id="schemafile"></a>
<a id="schema_File"></a>
<a id="tocSfile"></a>
<a id="tocsfile"></a>

```json
{
  "created_at": "2019-08-24T14:15:22Z",
  "extension": "string",
  "file_type": "File",
  "group_id": 0,
  "is_hidden": true,
  "name": "string",
  "owner_id": 0,
  "path": "string",
  "permissions": 0,
  "size": 0,
  "updated_at": "2019-08-24T14:15:22Z"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|created_at|string,null(date-time)|false|none|none|
|extension|string,null|false|none|none|
|file_type|[FileType](#schemafiletype)|true|none|none|
|group_id|integer(int32)|true|none|none|
|is_hidden|boolean|true|none|none|
|name|string|true|none|none|
|owner_id|integer(int32)|true|none|none|
|path|string|true|none|none|
|permissions|integer(int32)|true|none|none|
|size|integer(int64)|true|none|none|
|updated_at|string,null(date-time)|false|none|none|

<h2 id="tocS_FileListResponse">FileListResponse</h2>
<!-- backwards compatibility -->
<a id="schemafilelistresponse"></a>
<a id="schema_FileListResponse"></a>
<a id="tocSfilelistresponse"></a>
<a id="tocsfilelistresponse"></a>

```json
{
  "data": [
    {
      "created_at": "2019-08-24T14:15:22Z",
      "extension": "string",
      "file_type": "File",
      "group_id": 0,
      "is_hidden": true,
      "name": "string",
      "owner_id": 0,
      "path": "string",
      "permissions": 0,
      "size": 0,
      "updated_at": "2019-08-24T14:15:22Z"
    }
  ],
  "message": "string",
  "success": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|data|array,null|false|none|none|
|message|string|true|none|none|
|success|boolean|true|none|none|

<h2 id="tocS_FileSizeResponse">FileSizeResponse</h2>
<!-- backwards compatibility -->
<a id="schemafilesizeresponse"></a>
<a id="schema_FileSizeResponse"></a>
<a id="tocSfilesizeresponse"></a>
<a id="tocsfilesizeresponse"></a>

```json
{
  "data": 0,
  "message": "string",
  "success": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|data|integer,null(int64)|false|none|none|
|message|string|true|none|none|
|success|boolean|true|none|none|

<h2 id="tocS_FileType">FileType</h2>
<!-- backwards compatibility -->
<a id="schemafiletype"></a>
<a id="schema_FileType"></a>
<a id="tocSfiletype"></a>
<a id="tocsfiletype"></a>

```json
"File"

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|string|false|none|none|

#### Enumerated Values

|Property|Value|
|---|---|
|*anonymous*|File|
|*anonymous*|Directory|
|*anonymous*|Symlink|
|*anonymous*|Other|

<h2 id="tocS_GetApplicationById">GetApplicationById</h2>
<!-- backwards compatibility -->
<a id="schemagetapplicationbyid"></a>
<a id="schema_GetApplicationById"></a>
<a id="tocSgetapplicationbyid"></a>
<a id="tocsgetapplicationbyid"></a>

```json
{
  "id": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|id|string|true|none|none|

<h2 id="tocS_GetApplicationByName">GetApplicationByName</h2>
<!-- backwards compatibility -->
<a id="schemagetapplicationbyname"></a>
<a id="schema_GetApplicationByName"></a>
<a id="tocSgetapplicationbyname"></a>
<a id="tocsgetapplicationbyname"></a>

```json
{
  "app_id": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|app_id|string|true|none|none|

<h2 id="tocS_InstallApplication">InstallApplication</h2>
<!-- backwards compatibility -->
<a id="schemainstallapplication"></a>
<a id="schema_InstallApplication"></a>
<a id="tocSinstallapplication"></a>
<a id="tocsinstallapplication"></a>

```json
{
  "app_id": "string",
  "app_name": "string",
  "installation_id": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|app_id|string|true|none|none|
|app_name|string|true|none|none|
|installation_id|string|true|none|none|

<h2 id="tocS_ListContainersRequest">ListContainersRequest</h2>
<!-- backwards compatibility -->
<a id="schemalistcontainersrequest"></a>
<a id="schema_ListContainersRequest"></a>
<a id="tocSlistcontainersrequest"></a>
<a id="tocslistcontainersrequest"></a>

```json
{
  "environment": "Dev",
  "id": "string",
  "image_name": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|environment|[Environment](#schemaenvironment)|true|none|none|
|id|string|true|none|none|
|image_name|string|true|none|none|

<h2 id="tocS_ListRepositoriesRequest">ListRepositoriesRequest</h2>
<!-- backwards compatibility -->
<a id="schemalistrepositoriesrequest"></a>
<a id="schema_ListRepositoriesRequest"></a>
<a id="tocSlistrepositoriesrequest"></a>
<a id="tocslistrepositoriesrequest"></a>

```json
{
  "installation_id": 0
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|installation_id|integer(int64)|true|none|none|

<h2 id="tocS_Log">Log</h2>
<!-- backwards compatibility -->
<a id="schemalog"></a>
<a id="schema_Log"></a>
<a id="tocSlog"></a>
<a id="tocslog"></a>

```json
{
  "created_at": "2019-08-24T14:15:22Z",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "level": "string",
  "message": "string",
  "topic": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|created_at|string,null(date-time)|false|none|none|
|id|string(uuid)|true|none|none|
|level|string|true|none|none|
|message|string|true|none|none|
|topic|string,null|false|none|none|

<h2 id="tocS_MemoryUsageData">MemoryUsageData</h2>
<!-- backwards compatibility -->
<a id="schemamemoryusagedata"></a>
<a id="schema_MemoryUsageData"></a>
<a id="tocSmemoryusagedata"></a>
<a id="tocsmemoryusagedata"></a>

```json
{
  "free": 0,
  "total": 0,
  "used": 0
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|free|integer(int64)|true|none|none|
|total|integer(int64)|true|none|none|
|used|integer(int64)|true|none|none|

<h2 id="tocS_MemoryUsageResponse">MemoryUsageResponse</h2>
<!-- backwards compatibility -->
<a id="schemamemoryusageresponse"></a>
<a id="schema_MemoryUsageResponse"></a>
<a id="tocSmemoryusageresponse"></a>
<a id="tocsmemoryusageresponse"></a>

```json
{
  "data": {},
  "message": "string",
  "success": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|data|any|false|none|none|

oneOf

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» *anonymous*|null|false|none|none|

xor

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» *anonymous*|[MemoryUsageData](#schemamemoryusagedata)|false|none|none|

continued

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|message|string|true|none|none|
|success|boolean|true|none|none|

<h2 id="tocS_MoveFileRequest">MoveFileRequest</h2>
<!-- backwards compatibility -->
<a id="schemamovefilerequest"></a>
<a id="schema_MoveFileRequest"></a>
<a id="tocSmovefilerequest"></a>
<a id="tocsmovefilerequest"></a>

```json
{
  "from_path": "string",
  "to_path": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|from_path|string|true|none|none|
|to_path|string|true|none|none|

<h2 id="tocS_NewApplicationRequest">NewApplicationRequest</h2>
<!-- backwards compatibility -->
<a id="schemanewapplicationrequest"></a>
<a id="schema_NewApplicationRequest"></a>
<a id="tocSnewapplicationrequest"></a>
<a id="tocsnewapplicationrequest"></a>

```json
{
  "application_name": "string",
  "build_pack": "Dockerfile",
  "build_variables": "string",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "Dev",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|application_name|string|true|none|none|
|build_pack|[BuildPack](#schemabuildpack)|true|none|none|
|build_variables|string|true|none|none|
|custom_domain|string,null|false|none|none|
|description|string,null|false|none|none|
|docker_image|string|true|none|none|
|docker_ports|string|true|none|none|
|env_variables|string|true|none|none|
|environment|[Environment](#schemaenvironment)|true|none|none|
|installation_id|integer(int64)|true|none|none|
|post_run_commands|string|true|none|none|
|pre_run_commands|string|true|none|none|
|repository|string|true|none|none|
|repository_branch|string|true|none|none|
|repository_owner|string|true|none|none|

<h2 id="tocS_NewCronJob">NewCronJob</h2>
<!-- backwards compatibility -->
<a id="schemanewcronjob"></a>
<a id="schema_NewCronJob"></a>
<a id="tocSnewcronjob"></a>
<a id="tocsnewcronjob"></a>

```json
{
  "bash_script": "string",
  "command": "string",
  "description": "string",
  "is_active": true,
  "name": "string",
  "resource_limits": null,
  "schedule": "string",
  "tenant_id": "34f5c98e-f430-457b-a812-92637d0c6fd0",
  "user_id": "a169451c-8525-4352-b8ca-070dd449a1a5"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|bash_script|string,null|false|none|none|
|command|string|true|none|none|
|description|string,null|false|none|none|
|is_active|boolean|true|none|none|
|name|string|true|none|none|
|resource_limits|any|false|none|none|
|schedule|string|true|none|none|
|tenant_id|string(uuid)|true|none|none|
|user_id|string(uuid)|true|none|none|

<h2 id="tocS_StopApplicationRequest">StopApplicationRequest</h2>
<!-- backwards compatibility -->
<a id="schemastopapplicationrequest"></a>
<a id="schema_StopApplicationRequest"></a>
<a id="tocSstopapplicationrequest"></a>
<a id="tocsstopapplicationrequest"></a>

```json
{
  "environment": "Dev",
  "id": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|environment|[Environment](#schemaenvironment)|true|none|none|
|id|string|true|none|none|

<h2 id="tocS_UpdateApplicationForDatabase">UpdateApplicationForDatabase</h2>
<!-- backwards compatibility -->
<a id="schemaupdateapplicationfordatabase"></a>
<a id="schema_UpdateApplicationForDatabase"></a>
<a id="tocSupdateapplicationfordatabase"></a>
<a id="tocsupdateapplicationfordatabase"></a>

```json
{
  "application_name": "string",
  "build_pack": "string",
  "build_variables": "string",
  "created_at": "2019-08-24T14:15:22Z",
  "custom_domain": "string",
  "description": "string",
  "docker_image": "string",
  "docker_ports": "string",
  "env_variables": "string",
  "environment": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "installation_id": 0,
  "post_run_commands": "string",
  "pre_run_commands": "string",
  "repository": "string",
  "repository_branch": "string",
  "repository_owner": "string",
  "updated_at": "2019-08-24T14:15:22Z"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|application_name|string,null|false|none|none|
|build_pack|string,null|false|none|none|
|build_variables|string,null|false|none|none|
|created_at|string,null(date-time)|false|none|none|
|custom_domain|string,null|false|none|none|
|description|string,null|false|none|none|
|docker_image|string,null|false|none|none|
|docker_ports|string,null|false|none|none|
|env_variables|string,null|false|none|none|
|environment|string,null|false|none|none|
|id|string(uuid)|true|none|none|
|installation_id|integer,null(int64)|false|none|none|
|post_run_commands|string,null|false|none|none|
|pre_run_commands|string,null|false|none|none|
|repository|string,null|false|none|none|
|repository_branch|string,null|false|none|none|
|repository_owner|string,null|false|none|none|
|updated_at|string,null(date-time)|false|none|none|

<h2 id="tocS_UpdateCronJob">UpdateCronJob</h2>
<!-- backwards compatibility -->
<a id="schemaupdatecronjob"></a>
<a id="schema_UpdateCronJob"></a>
<a id="tocSupdatecronjob"></a>
<a id="tocsupdatecronjob"></a>

```json
{
  "bash_script": "string",
  "command": "string",
  "description": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "is_active": true,
  "name": "string",
  "resource_limits": null,
  "schedule": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|bash_script|string,null|false|none|none|
|command|string,null|false|none|none|
|description|string,null|false|none|none|
|id|string(uuid)|true|none|none|
|is_active|boolean,null|false|none|none|
|name|string,null|false|none|none|
|resource_limits|any|false|none|none|
|schedule|string,null|false|none|none|

