package main

var (
	env = []byte(`// some javascript
function Task(fn) {
	this.fn = fn;
	this.run = function(host) {
		console.log("Task.run was called");
		Remote(h).run(this.fn);
	}
}
var _tasks = {};
var _connections = {};
function task(name, fn) {
	_tasks[name] = new Task(fn);
}`)
	mainJS = []byte(`// mainloop
if (typeof lookup !== "undefined") {
	cloud.Hosts = _.intersection(_.flatten(_.map(cloud.Roles, lookup)));
}
for (i = 0; i < cloud.Tasks.length; ++i) {
	_do(_tasks[cloud.Tasks[i]], cloud.Hosts)
}`)
)
