# ThingsDB Module (Go)

ThingsDB module written using the [Go language](https://golang.org).

With this module you gain access other scopes. This works both on the same node as with a connection to another ThingsDB cluster.

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
username    | str (optional)  | Username for authentication.
password    | str (optional)  | Password for the username.
token       | str (optional)  | Token for authentication _(instead of `username` and `password`)_.
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
    res;  // response as "mpdata"
});
```

### run

Syntax: `run(scope, name, args)`

#### Arguments

- `scope`: The scope of the procedure.
- `name`: Name fot the procedure to run.
- `args`: Optional arguments as a _list_ or _thing_ for the procedure.

#### Example:

```javascript
thingsdb.run("//stuff", "addition", {a: 4, b: 7}).then(|res| {
    res;  // response as "mpdata"
});
```