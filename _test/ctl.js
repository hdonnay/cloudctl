function lookup(role){
	var re = new RegExp(role, "g");
	var hosts = shell("cat hostlist").stdout.match(re);
	return hosts
}
Task("echo", function(run){
	run(
		"echo echo",
		"touch /tmp/cloudctl"
	);
});
Task("where", function(run){
	run(
		"whoami",
		"hostname"
	);
});
