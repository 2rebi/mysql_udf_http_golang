# mysql_udf_http_golang
[MySQL](https://dev.mysql.com/) or [MariaDB](https://mariadb.com/) UDF(User-Defined Functions) HTTP Client Plugin

Call RESTful API on query.

Setup 
---
- **Clone Source**
```shell
git clone https://github.com/RebirthLee/mysql_udf_http_golang.git udf
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

```query
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
SELECT http_get(url, option...);
```   

- **Simple Request**
```sql
SELECT http_get('http://example.com');
```
**Return**
```javascript
{
    "Body": String(HTML(Default), Base64, Hexdecimal)
}
```

- **Output Option**  
`-O outputType`	Define kind of result.

- Kinds  
`PROTO`, `STATUS` or `STATUS_CODE`, `HEADER`, `BODY`(default), `FULL`    
`-O PROTO|STATUS|HEADER|BODY` same this `-O FULL`.

```sql
SELECT http_get('http://example.com', '-O FULL');
```
**Return**
```javascript
{
    "Proto": String(Http Version, HTTP/1.0, HTTP/1.1, HTTP/2.0),
    "Status": String(Status Code, 200 OK, 404 NOT FOUND),
    "Header": JSON(`{Key : Array, ...}`),
    "Body": String(HTML(Default), Base64, Hexdecimal)
}
```
#
#
#
#
### writing...  plz wait...
### Add Soon
### - POST Method
```sql
SELECT http_post();
```
  
License
---
[`THE BEER-WARE LICENSE (Revision 42)`](http://en.wikipedia.org/wiki/Beerware)