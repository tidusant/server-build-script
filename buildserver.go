package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/nlopes/slack"

	//"io"

	"net/http"
	"net/url"

	//	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	var port int
	var debug bool
	var slacktoken string
	var mytoken string
	var gittoken string
	var slackchannel string
	var rootpath string
	var servername string
	//fmt.Println(mycrypto.Encode("abc,efc", 5))
	// nohup ./buildserver --servername="aro4viet" -port=8081 -mytoken=xgdedkillaccnqweoiurpelksfcvnbsdw --gittoken=e4955a08780681068807698fae5ce99997cb430a & 
	flag.IntVar(&port, "port", 8081, "help message for flagname")
	flag.StringVar(&slacktoken, "slacktoken", "xoxb-298302086051-Q5ZYSQxIndUCo05vD6QfAyQi", "slacktoken")
	flag.StringVar(&slackchannel, "slackchannel", "buildserver", "slackchannel")
	flag.StringVar(&mytoken, "mytoken", "abc111", "mytoken")
	flag.StringVar(&gittoken, "gittoken", "abc111", "gittoken")
	flag.StringVar(&rootpath, "rootpath", "/home/ec2-user", "root path")
	flag.StringVar(&servername, "servername", "phubuildserver", "server name")
	flag.BoolVar(&debug, "debug", false, "Indicates if debug messages should be printed in log files")
	flag.Parse()

	if !debug {
		//logLevel = log.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	slackapi := slack.New(slacktoken)
	_, _, err := slackapi.PostMessage(slackchannel, slack.MsgOptionText("Server build started at "+servername+"!\n-Token: "+mytoken+"\n-Port: "+strconv.Itoa(port), false))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	// err = exec.Command("dir").Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//fmt.Printf("The date is %s\n", out)

	// cmd := exec.Run() .Command("go", "get", "github.com/tidusant/c3m-common/...")
	// cmd = exec.Command("go", "get", "github.com/tidusant/chadmin-repo/...")

	//init config

	router := gin.Default()

	router.POST("/:action/:name", func(c *gin.Context) {
		packageserver := "http://" + GetOutboundIP() + "/"
		slackmsg := "==> Start building from "
		strrt := ""
		c.Header("Access-Control-Allow-Origin", "*")
		name := c.Param("name")
		//name = name[1:] //remove slash
		action := c.Param("action")
		userIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
		slackmsg += userIP + " on " + packageserver + "<==\n - Token: " + name
		slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, true))

		x, _ := ioutil.ReadAll(c.Request.Body)
		datastr := strings.Replace(string(x), "payload=", "", 1)
		datastr, _ = url.QueryUnescape(datastr)
		//fmt.Printf("datastr: %s", datastr)
		data := make(map[string]json.RawMessage)
		json.Unmarshal([]byte(datastr), &data)
		//fmt.Printf("data: %v", string(data["ref"]))
		branch := ""
		json.Unmarshal(data["ref"], &branch)
		branch = strings.Replace(branch, "refs/heads/", "", 1)

		repository := make(map[string]json.RawMessage)
		json.Unmarshal(data["repository"], &repository)
		reponame := ""
		json.Unmarshal(repository["name"], &reponame)

		fullname := ""
		json.Unmarshal(repository["full_name"], &fullname)
		username := strings.Replace(fullname, "/"+reponame, "", 1)

		if branch == "" {
			json.Unmarshal(repository["default_branch"], &branch)
		}

		slackmsg = "==> Repo " + fullname + " Built <=="

		slackmsg += "\n reponame: " + fullname + " - branch: " + branch
		slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
		slackapi.PostMessage(slackchannel, slack.MsgOptionText("name:"+name+" mytoken:"+mytoken, false))

		if fullname == "" {
			slackmsg += "\nEmpty repo"
		} else if mytoken == name {
			if action == "serverbuild" {

				repodir := rootpath + "/repo/" + fullname
				buildscriptdir := rootpath + "/repo/" + username + "/buildscript/" + reponame + "/" + branch
				
				// if _, err := os.Stat(repodir); os.IsNotExist(err) {
				// 	slackmsg = "\n- MKDIR " + repodir
				// 	slackmsg += outputCmd("mkdir " + repodir)
				// 	slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				// }
				slackmsg = "\n- clear all file for checkout "
				
				slackmsg += outputCmd("rm -Rf " + repodir+"/")
				slackmsg += outputCmd("mkdir " + repodir)
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))

				
				slackmsg = "\n- GIT CLONE"
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
			
				output := outputCmd("git clone -b "+branch+" https://" + gittoken + "@github.com/" + fullname + ".git " + repodir)
				slackmsg = string(output)
				
				if slackmsg != "" {
					slackmsg += "\n" + slackmsg+"\n QUIT"
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
					return
				}
				os.Chdir(repodir)
					

				/*
				} else {

					os.Chdir(repodir)

					
					// output, err := exec.Command("yes|", "rm","-R","./*").Output()
					// slackmsg = "\n" + string(output)
					// if err != nil {
					// 	slackmsg += "\n" + fmt.Sprintf("error:%v", err) + "\n QUIT"
					// 	slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
					// 	return
					// }
					// slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))

					

					slackmsg = "\n- GIT CHECKOUT " + branch
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
					output, err := exec.Command("git", "checkout", branch).Output()
					slackmsg = "\n" + string(output)
					if err != nil {
						slackmsg += "\n" + err.Error() + "\n QUIT"
						slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
						return
					}
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))

					slackmsg = "\n- GIT RESET "
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
					output, err = exec.Command("git", "reset", "--hard").Output()
					slackmsg = "\n" + string(output)
					if err != nil {
						slackmsg += "\n" + fmt.Sprintf("error:%v", err) + "\n QUIT"
						slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
						return
					}
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				}


				//go import === no need now
				/*
				file, err := os.Open(repodir + "/imports.txt")
				if err != nil {
					slackmsg = "\n" + err.Error()
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				}
				defer file.Close()
				scanner := bufio.NewScanner(file)
				slackmsg = ""
				for scanner.Scan() {
					line := scanner.Text()

					if strings.Trim(line, " ") == "" || line[:1] == "#" {
						continue
					}
					slackmsg += "\n- go get " + line

					output := outputCmd("go get " + line)
					if output != "" {
						//exit when error
						slackmsg += output + "\nQUIT"
						slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
						return
					}

				}
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				*/




				//=========== deploy config
				file, err := os.Open(buildscriptdir + "/deploy.txt")
				if err != nil {
					slackmsg = "\n" + err.Error()
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
					return
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				//read build script
				instruction := ""
				argstr := ""
				slackmsg = ""
				var serverdeploys []string
				for scanner.Scan() {
					line := scanner.Text()
					if strings.Trim(line, " ") == "" {
						continue
					}
					if instruction != line && line[:1] == "#" {
						instruction = line
						continue
					}
					if instruction == "#deploys_server" {
						slackmsg += "\n- deploys_server: " + line
						serverdeploys = append(serverdeploys, line)
					} else if instruction == "#argstr" {

						argstr += " " + line
						slackmsg += "\n- Run with args: " + argstr
					} else if instruction == "#package_server" {
						slackmsg += "\n- package_server: " + line
						packageserver = line
					}

				}
				if err := scanner.Err(); err != nil {
					slackmsg += "\n" + err.Error()
				}
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))

				//copy config file
				slackmsg = "\n- Copy config file"
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				slackmsg = outputCmd("cp " + buildscriptdir + "/config.toml " + repodir + "/config.toml")

				//slackmsg += outputCmd("cp " + buildscriptdir + "/Dockerfile " + repodir + "/Dockerfile")
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, true))
				//build go
				packagename := reponame
				runningname := strings.Replace(packagename, "-", "", -1)
				slackmsg = "\n- build go " + runningname
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				outputcompile := outputCmd("env CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o " + runningname + " .")
				if strings.TrimSpace(outputcompile) != "" {
					slackmsg = "\nERROR:" + outputcompile + "\nQUIT"
					slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
					return
				}
				slackmsg = " DONE"
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))

				//random
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				randomnumber := strconv.Itoa(r.Intn(1000000000))
				slackmsg = "\n- Package " + packagename

				//remove old package
				packagepath := "/var/www/repo_publish/" + branch + "/"
				slackmsg += outputCmd("rm -f -R " + packagepath + packagename)
				slackmsg += outputCmd("mkdir " + packagepath + packagename)

				//check folder data & html exit
				var htmlFolder = ""
				var htmlData = ""
				if _, err := os.Stat("html"); err == nil {
					htmlFolder = "html"
				}
				if _, err := os.Stat("data"); err == nil {
					htmlData = "data"
				}

				slackapi.PostMessage(slackchannel, slack.MsgOptionText("Create package "+randomnumber, false))
				
				// output2, err := exec.Command("tar","-czf",packagepath + packagename + "/" + randomnumber + ".pkg",htmlFolder,htmlData,runningname,"config.toml").Output()
				// slackmsg = "\n" + string(output2)
				// if err != nil {
				// 	slackmsg += "\n" + fmt.Sprintf("error:%v", err) + "\n QUIT"
				// 	slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				// 	return
				// }
				// slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))

				slackmsg += outputCmd("tar -czf " + packagepath + packagename + "/" + randomnumber + ".pkg " + htmlFolder + " " + htmlData + " " + runningname + " config.toml")
				slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
				

				slackmsg = ""
				os.Chdir(rootpath)
				//deploy server:
				for _, server := range serverdeploys {

					go func(server, mytoken, packagename, randomnumber, argstr, packageserver, username string) {
						slackmsg2 := "\n- Trigger server:" + server + "deploy/" + mytoken
						form := url.Values{}
						form.Add("pn", packagename)
						form.Add("rn", randomnumber)
						form.Add("ag", argstr)
						form.Add("sv", packageserver)
						form.Add("rpun", username)
						slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg2, false))

						req, _ := http.NewRequest("POST", server+"deploy/"+mytoken, strings.NewReader(form.Encode()))
						req.PostForm = form
						req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
						hc := http.Client{}
						hc.Do(req)
					}(server, mytoken, packagename, randomnumber, argstr, packageserver, username)

				}
			} else if action == "updatelib" {
				cmdstr := "go get -u github.com/" + fullname + "/..."
				slackmsg += "\n -" + cmdstr
				slackmsg += outputCmd(cmdstr)
			} else {
				slackmsg += "\nInvalid action: " + action
			}

		} else {
			slackmsg += "\nInvalid token"

		}
		slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg, false))
		c.String(http.StatusOK, strrt)

	})
	router.Run(":" + strconv.Itoa(port))

}

// Get preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
func outputCmd(cmdstr string) string {
	rt := ""
	args := strings.Split(cmdstr, " ")

	// if len(args) > 1 {
	// 	[]string{"what", "ever", "you", "like"}
	// }

	cmd := exec.Command(args[0])
	// cmd := exec.Command("/bin/sh")
	cmd.Args = args

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	//fmt.Printf("args: %v", cmd.Args)
	cmd.Env = os.Environ()
	//output, err := cmd.Output()
	err := cmd.Run()

	if err != nil {
		rt += "\nrunning " + cmdstr
		rt += " ERROR: " + stderr.String()
	}
	rt += "\n" + out.String()
	return rt
}
