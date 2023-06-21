# mysql_udf_http_golang
Now works for Mysql 8+ also.

[![MySQL UDF](https://img.shields.io/badge/MySQL-UDF-blue.svg)](https://dev.mysql.com/) [![MariaDB UDF](https://img.shields.io/badge/MariaDB-UDF-blue.svg)](https://mariadb.com/)

[MySQL](https://dev.mysql.com/) or [MariaDB](https://mariadb.com/) UDF(User-Defined Functions) HTTP Client Plugin

Call RESTful API on query.

Setup 
---
- **Clone Source**
```shell
git clone https://github.com/parthasai/mysql_udf_http_golang.git udf
cd udf
```

- **Auto Build**
```shell
bash ./install.sh {username} {password}
```

{username} replace your MySQL or MariaDB Username.  
{password} replace your MySQL or MariaDB Password(Optional).

- **Manual Build**
```shell
bash ./build.sh
```
Build output is `http.so`, move file to `plugin_dir` path.   
if you don't know `plugin_dir` path.  
Command input this on MySQL, MariaDB connection.

```sql
SHOW VARIABLES LIKE 'plugin_dir';
```

**Ex)**
```shell
$ mysql -u root -p
Enter password: 
```
**And**
```sql
MariaDB [(none)]> SHOW VARIABLES LIKE 'plugin_dir';
+---------------+-----------------------------------------------+
| Variable_name | Value                                         |
+---------------+-----------------------------------------------+
| plugin_dir    | /usr/local/Cellar/mariadb/10.3.12/lib/plugin/ |
+---------------+-----------------------------------------------+
1 row in set (0.001 sec)
```

and `http.so` move to `Value` path.
```shell
mv ./http.so /usr/local/Cellar/mariadb/10.3.12/lib/plugin/
```
### Finally, execute query

- **Http Help**
```sql
CREATE FUNCTION http_help RETURNS STRING SONAME 'http.so';
```
- **Http Get Method**
```sql
CREATE FUNCTION http_get RETURNS STRING SONAME 'http.so';
```
- **Http Post Method**
```sql
CREATE FUNCTION http_post RETURNS STRING SONAME 'http.so';
```


Usage
---

### - Help

```sql
SELECT http_help();
```

### - GET Method

- **Prototype**
```sql
SELECT http_get(url, options...);
```   

- **Simple Request**
```sql
SELECT http_get('http://example.com');
```
**Return**
```javascript
{
    "Body" : String(HTML(Default), Base64, Hexdecimal)
}
```

- **Output Option**  

```sql
SELECT http_get('http://example.com', '-O FULL');
```
**Return**
```javascript
{
    "Proto"  : String(Http Version, HTTP/1.0, HTTP/1.1, HTTP/2.0),
    "Status" : String(Status Code, 200 OK, 404 NOT FOUND...),
    "Header" : JSON(`{Key : Array, ...}`),
    "Body"   : String(HTML(Default), Base64, Hexdecimal)
}
```
`-O {outputType}`	Define kind of result.  
`PROTO`, `STATUS` or `STATUS_CODE`, `HEADER`, `BODY`(default), `FULL`    
`-O PROTO|STATUS|HEADER|BODY` same this `-O FULL`.


- **Custom Header**  

```sql
SELECT http_get('http://example.com', '-O FULL', '-H CustomKey:CustomValue', '-H Authorization:Bearer a1b2c3d4-123e-5678-9fgh-ijk098765432')
```
**Like this**
```http
GET / HTTP/1.1
Host: example.com
CustomKey: CustomValue
Authorization: Bearer a1b2c3d4-123e-5678-9fgh-ijk098765432
User-Agent: Go-http-client/1.1
Accept-Encoding: gzip
```

Option param input  `-H {key}:{value}`.  

### - POST Method
- **Prototype**
```sql
SELECT http_post(url, contentType, body, options...)
```
- **Simple Request(No Body)**
```sql
SELECT http_post('http://example.com', '', '');
```
- **Simple Request(Json Body)**
```sql
SELECT http_post('http://example.com', 'application/json', '{"Hello":"World"}');
```
**Like this**
```http
POST / HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 17
User-Agent: Go-http-client/1.1
Accept-Encoding: gzip


{"Hello":"World"}
```
### - Raw Method
- **Prototype**
```sql
SELECT http_raw(method, url, body, options...)
```

- **PUT**
```sql
SELECT http_raw('PUT', url, body, options...)
```
- **PATCH**
```sql
SELECT http_raw('PATCH', url, body, options...)
```
- **DELETE**
```sql
SELECT http_raw('DELETE', url, body, options...)
```

License
---
[`THE BEER-WARE LICENSE (Revision 42)`](http://en.wikipedia.org/wiki/Beerware)
