package main

var (
	env = []byte(`
var Task = function(name, fn){
	cloud._tasks[name] = fn;
};`)
	mainJS = []byte(`// mainloop
if(typeof lookup !== "undefined"){
	cloud.Hosts = _.filter(
		_.union(_.flatten(_.map(cloud.Roles, lookup)), cloud.Hosts),
		function(s){return s != ""});
}
_do(cloud);`)
)

const (
	restTask = `Task("_rest", function(run){
	run("%s");
});
cloud.Tasks = ["_rest"];
`
)
