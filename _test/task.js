/*
 * task("echo", function() {
	cd("/tmp/", function() {
		run("touch test");
		run("echo across the internet!");
		sudo(function(){
			run("whoami");
		})
		echo("locally");
	})
})
*/
task("echo", function(){
	run("echo echo");
})

// cloudctl -H host1,host2 echo
