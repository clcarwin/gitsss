package main

import (
	"fmt"
	"net/http"
	"strings"
	"os/exec"
	"io"
)

// http://xxx/username/test.git
// http://xxx/username/test.git/HEAD
// http://xxx/username/test.git/info/refs
// http://xxx/username/test.git/git-<command>
// http://xxx/username/test.git/src/master/xx.md


func requesthandler(w http.ResponseWriter, r *http.Request) {
	plist := strings.Split(r.URL.Path,"/")
	if plist[0]=="" { plist = plist[1:] }	// delete first empty element

	if len(plist)>=4 && plist[2] == "info" && plist[3] == "refs" {
		// reponame := plist[1]
		repopath := "./" + plist[0] + "/" + plist[1]

		service := r.FormValue("service")
		contenttype := fmt.Sprintf("application/x-%s-advertisement",service)
		w.Header().Add("Content-type",contenttype)
		gitcmd := exec.Command(
			"git",
			string(service[4:]),
			"--stateless-rpc",
			"--advertise-refs",
			repopath)
		out, err := gitcmd.CombinedOutput()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, "Internal Server Error")
			w.Write(out)
		} else {
			serveradvert := fmt.Sprintf("# service=%s",service)
			length := len(serveradvert) + 4
			fmt.Fprintf(w, "%04x%s0000", length, serveradvert)
			w.Write(out)
		}
	} else if len(plist)>=3 && plist[2][:4]=="git-" {
		// reponame := plist[1]
		repopath := "./" + plist[0] + "/" + plist[1]

		service := plist[2]
		contenttype := fmt.Sprintf("application/x-%s-result", service)
		w.Header().Add("Content-type",contenttype)
		w.WriteHeader(200)
		gitcmd := exec.Command(
			"git",
			string(service[4:]),
			"--stateless-rpc",
			repopath)

		cmdin,_ := gitcmd.StdinPipe()
		cmdout,_ := gitcmd.StdoutPipe()
		gitcmd.Start()
		io.Copy(cmdin,r.Body)
		io.Copy(w, cmdout)

		if service == "git-receive-pack" {
			updatecmd := exec.Command(
				"git",
				"--git-dir",
				repopath,
				"update-server-info")
			updatecmd.Start()
		}

	} else if len(plist)>=4 && plist[2]=="src" {
		// branchname = plist[3]

	} else {
		http.ServeFile(w,r,"."+r.URL.Path)
	}
}


func main() {
	http.HandleFunc("/",requesthandler)

	http.ListenAndServe(":2000",nil)
}




