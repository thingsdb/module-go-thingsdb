# ThingsDB Module (Go)

ThingsDB module written using the [Go language](https://golang.org).


## Installation

Install the module by running the following command in the `@thingsdb` scope:

```javascript
new_module('thingsdb', 'github.com/thingsdb/module-go-thingsdb');
```

Optionally, you can choose a specific version by adding a `@` followed with the release tag. For example: `@v0.1.0`.

## Configuration

The `host` is required and either a `username` & `password` _or_ a valid `token`.

Property    | Type            | Description
----------- | --------------- | -----------
host        | str (required)  | Address of the ThingsDB node.
port        | int (optional)  | Port of the node, defaults to `9200`.
username    | str (optional)  | Database user to authenticate with.
password    | str (optional)  | Password for the database user.
token       | str (optional)  | Database to connect to.
use_ssl     | bool (optional) | Use SSL for the connection, default to `false`.
skip_verify | bool (optional) | Only for SSL, default to `false`.
nodes       | list (optional) | List with additional nodes as `{host: <NODE_ADDRESS>, port: <PORT>}` where `port` is optional.

Example configuration:

```javascript
set_module_conf('thingsdb', {
    host: 'localhost',
    username: 'admin',
    password: 'pass',
});
```

## Exposed functions

Name              | Description
----------------- | -----------
[query](#query)   | Perform a query.
[run](#run)       | Run a procedure.

### query

Syntax: `query(scope, code, vars)`

#### Arguments

- `scope`: The scope to run the query in.
- `code`: The code for the query.
- `vars`: Optional variable for the query.

#### Example:

```javascript
thingsdb.query("//stuff", ".has(key);", {key: "example"}).then(|res| {
    res;  // just return the response.
});
```

### run

Syntax: `query(scope, code, vars)`

#### Arguments

- `scope`: The scope to run the query in.
- `code`: The code for the query.
- `vars`: Optional variable for the query.

#### Example:

```javascript
thingsdb.run("//stuff", "addition", {a: 4, b: 7}).then(|res| {
    res;  // just return the response.
});
```