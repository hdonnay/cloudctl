Run commands on many machines at once!

Define tasks in javascript and have them execute!

Currently, a few commands are passed in:

`shell(cmd)`: run a command in a local shell and capture the output.

`include(file)`: eval a local javascript file.

`Task(name, func)`: define a function.

The `func` argument should be a function that takes another function as an
argument, usually called `run`. `run` is used to execute commands on the remote
host.

TODO:

* learn javascript
* "context functions" (`cd("to/path", function(){...});`)
