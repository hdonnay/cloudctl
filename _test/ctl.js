console.log("starting ctl.js");
console.log(JSON.stringify(cloud));
include("echo.js");
include("touch.js");
console.log(JSON.stringify(cloud));
touch("test1");
echo("echo");
echo(shell("ls test*").stdout);
console.log("ending ctl.js");
