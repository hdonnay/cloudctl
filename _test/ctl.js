console.log("starting ctl.js");
task("echo", function(){
	run("echo echo");
})
console.log(JSON.stringify(cloud));
console.log("ending ctl.js");
