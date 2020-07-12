package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"

	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nlopes/slack"

	//"io"

	"net/http"
	"net/url"

	//	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

var slackchannel string
var slacktoken string
var slackapi *slack.Client

func main() {
	var port int
	var debug bool

	var mytoken string
	var gittoken string

	var rootpath string
	var gopath string
	var servername string
	//fmt.Println(mycrypto.Encode("abc,efc", 5))

	flag.IntVar(&port, "port", 8081, "help message for flagname")
	flag.StringVar(&slacktoken, "slacktoken", "xoxb-298302086051-SAQWpyog0n576OajH5JScPBz", "slacktoken")
	flag.StringVar(&slackchannel, "slackchannel", "buildserver", "slackchannel")
	flag.StringVar(&mytoken, "mytoken", "abc111", "mytoken")
	flag.StringVar(&gittoken, "gittoken", "abc111", "gittoken")

	flag.StringVar(&servername, "servername", "phubuildserver", "server name")
	flag.BoolVar(&debug, "debug", false, "Indicates if debug messages should be printed in log files")
	flag.Parse()

	if !debug {
		//logLevel = log.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	slackapi = slack.New(slacktoken)
	slackmsg("Server build started at " + servername + "!\n-Token: " + mytoken + "\n-Port: " + strconv.Itoa(port))

	// err = exec.Command("dir").Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//fmt.Printf("The date is %s\n", out)

	// cmd := exec.Run() .Command("go", "get", "github.com/tidusant/c3m-common/...")
	// cmd = exec.Command("go", "get", "github.com/tidusant/chadmin-repo/...")

	//init config

	router := gin.Default()
	rootpath = "/root"
	gopath = "/root/go/src/"
	router.POST("/:action/:name", func(c *gin.Context) {
		strrt := ""
		c.Header("Access-Control-Allow-Origin", "*")
		name := c.Param("name")
		//name = name[1:] //remove slash
		action := c.Param("action")

		slackmsg("\n\n==> Start building on " + servername + "<==")

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

		slackmsg("==> Repo: " + fullname + ", branch: " + branch + " <==")

		if fullname == "" {
			slackmsg("Empty repo")
			return
		} else if mytoken == name {
			if action == "serverbuild" {
				os.Chdir(rootpath)
				repodir := rootpath + "/repo/" + fullname
				buildscriptdir := rootpath + "/repo/" + username + "/buildscript/" + reponame + "/" + branch

				if _, err := os.Stat(repodir); os.IsNotExist(err) {
					outputCmd("mkdir -p " + repodir)
					if outputCmd("git clone -b "+branch+" https://"+gittoken+"@github.com/"+fullname+".git "+repodir) != true {
						return
					}
				}
				slackmsg("change dir: " + repodir )
				os.Chdir(repodir)

				//reset hard, change branch then get latest version
				if outputCmd("git reset --hard") != true || outputCmd("git checkout "+branch) != true || outputCmd("git pull origin") != true {
					return
				}

				//go import === no need now
				slackmsg("readfile: " + repodir + "/import.txt")
				file, err := os.Open(repodir + "/import.txt")
				if err != nil {
					slackmsg("ERROR: " + err.Error())
				} else {
					instruction := ""
					scanner := bufio.NewScanner(file)
					for scanner.Scan() {
						line := strings.Trim(scanner.Text()," ")
						if strings.Trim(line, " ") == "" {
							continue
						}
						if instruction != line && line[:1] == "#" {
							instruction = line
							continue
						}
						if instruction == "#import" {
							if outputCmd("go get "+line) != true {
								return
							}
						}
						if instruction == "#checkout" {
							slackmsg("\nLOAD library: " + line)
							golibpath := gopath + line
							//if not get the lib yet, then go get,
							//then change branch,
							//then pull latest version
							if _, err := os.Stat(golibpath); os.IsNotExist(err) {
								if outputCmd("git clone -b "+branch+" https://"+gittoken+"@github.com/"+strings.Replace(line, "github.com/", "", -1) +".git "+golibpath) != true {
									return
								}
								//read file import.txt to get other library need from this library
								slackmsg("readfile: " + golibpath + "/import.txt")
								file2, err2 := os.Open(golibpath + "/import.txt")		
								if err2 != nil {
									slackmsg("ERROR: " + err2.Error())
								} else {
									
									scanner2 := bufio.NewScanner(file2)
									for scanner2.Scan() {
										line2 := scanner2.Text()
										if outputCmd("go get "+line2) != true {
											return
										}
									}
								}						
							}
							slackmsg("change dir: " + golibpath )
							os.Chdir(golibpath)
							if outputCmd("git checkout "+branch) != true || outputCmd("git pull origin") != true {
								return
							}

						}

					}
				}
				defer file.Close()

				//=========== deploy config
				var serverdeploys []string
				var app_prefix string
				var argstr string
				var packageserver string
				slackmsg("readfile: " + buildscriptdir + "/deploy.txt")
				file, err = os.Open(buildscriptdir + "/deploy.txt")
				if err != nil {
					slackmsg("ERROR: " + err.Error())
					return
				} else {
					scanner := bufio.NewScanner(file)
					//read build script
					instruction := ""

					for scanner.Scan() {
						line := strings.Trim(scanner.Text()," ")
						if line == "" {
							continue
						}
						if instruction != line && line[:1] == "#" {
							instruction = line
							continue
						}
						if instruction == "#deploys_server" {
							slackmsg("- deploys_server: " + line)
							serverdeploys = append(serverdeploys, line)
						} else if instruction == "#argstr" {
							argstr += " " + line
							slackmsg("- Run with args: " + argstr)
						} else if instruction == "#package_server" {
							slackmsg("- package_server: " + line)
							packageserver = line
						} else if instruction == "#app_prefix" {
							slackmsg("- app_prefix: " + line)
							app_prefix = line
						}

					}

				}
				defer file.Close()

				//copy config file
				if outputCmd("cp "+buildscriptdir+"/config.toml "+repodir+"/config.toml") != true {
					return
				}

				//build go
				slackmsg("change dir: " + repodir )
				os.Chdir(repodir)
				
				runningname := strings.Replace(reponame, "-", "", -1) + app_prefix
				if outputCmd("env CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o "+runningname+" .") != true {
					return
				}

				//random
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				randomnumber := strconv.Itoa(r.Intn(1000000000))

				//remove old package
				packagepath := "/var/www/repo_publish/" + branch + "/"+app_prefix+ "/"
				outputCmd("rm -f -R " + packagepath + reponame)
				outputCmd("mkdir -p " + packagepath + reponame)

				//check folder data & html exit
				var htmlFolder = ""
				var htmlData = ""
				if _, err := os.Stat("html"); err == nil {
					htmlFolder = "html"
				}
				if _, err := os.Stat("data"); err == nil {
					htmlData = "data"
				}

				outputCmd("tar -czf "+packagepath+reponame+"/"+randomnumber+".pkg "+htmlFolder+" "+htmlData+" "+runningname+" config.toml")
				
				

				os.Chdir(rootpath)
				//deploy server:
				for _, server := range serverdeploys {

					go func(server, mytoken, reponame, app_prefix, randomnumber, argstr, packageserver, username string) {
						slackmsg2 := "\n- Trigger server:" + server + "deploy/" + mytoken
						form := url.Values{}
						form.Add("reponame", reponame)
						form.Add("rn", randomnumber)
						form.Add("ag", argstr)
						form.Add("sv", packageserver)
						form.Add("rpun", username)
						form.Add("app_prefix", app_prefix)
						slackapi.PostMessage(slackchannel, slack.MsgOptionText(slackmsg2, false))

						req, _ := http.NewRequest("POST", server+"deploy/"+mytoken, strings.NewReader(form.Encode()))
						req.PostForm = form
						req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
						hc := http.Client{}
						hc.Do(req)
					}(server, mytoken, reponame, app_prefix, randomnumber, argstr, packageserver, username)

				}
			} else {
				slackmsg("Invalid action: " + action)
			}

		} else {
			slackmsg("Invalid token")

		}

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
func slackmsg(message string) {
	slackapi.PostMessage(slackchannel, slack.MsgOptionText(message, false))
}
func outputCmd(cmdstr string) bool {
	rt := true
	slackmsg(cmdstr)
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
		rt = false
		slackmsg(" ERROR: " + stderr.String())

	}

	return rt
}
