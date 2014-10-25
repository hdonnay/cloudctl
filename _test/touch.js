cloud.ShellTimeout = 0;
ret = shell("touch", "test");
if (ret.exit_code != 0) {
	console.log("touch.js touched a thing");
} else {
	console.log("touch.js didn't touched a thing");
}

function touch(f) {
	if (shell("touch", f).exit_code != 0) {
		return false;
	}
	return true;
}
