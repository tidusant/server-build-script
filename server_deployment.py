import web
from urlparse import parse_qs, urlparse
import urllib
import json
import subprocess
import os
import requests
import sys
import socket
from slackclient import SlackClient
urls = (
    '/deploy/(.*)', 'hello'
)
app = web.application(urls, globals())
token = sys.argv[2]
slacktoken = sys.argv[3]
slackchannel = sys.argv[4]
homedir=os.getcwd()
hostname = socket.gethostname()
sc = SlackClient(slacktoken)
def slack_message(message):
        sc.api_call("chat.postMessage",channel="#"+slackchannel,text=message)
def output_command(command):
        try:
                slack_message(command)                                
                subprocess.check_output(command,shell=True,stderr=subprocess.STDOUT)                
                return True
        except subprocess.CalledProcessError as e:
                slackmsg="ERROR: {}".format(e.output)
                slackmsg += "\nExit!"
                slack_message(slackmsg)
                return False


slack_message("-- started server with token: "+token+" --")
class hello:
        def POST(self, name):                
                output=''
                if name==token:
                        #runname=name.replace("/","_")
                        runname=name
                        
                        
                        slackmsg="==> Start Deploying at "+ hostname+" "
                        querystr=parse_qs(urllib.unquote(web.data()))
                        maindir=os.getcwd()
                        reponame=querystr['reponame'][0]
                        
                        argstr=querystr['ag'][0]
                        repousername=querystr['rpun'][0]
                        randomnumber=querystr['rn'][0]
                        packageserver=querystr['sv'][0]
                        app_prefix=querystr['app_prefix'][0]
                        runningname=reponame.replace('-','')+app_prefix
                        packagedir=maindir+"/"+repousername+"_"+reponame
                        slackmsg+=reponame+" <=="
                        slackmsg += reponame + " from " + packageserver + " <=="
                        slackmsg += "\n- packagedir: " + packagedir
                        slackmsg += "\n- rannum: " + randomnumber
                        slackmsg += "\n- argstr: " + argstr
                        slackmsg += "\n- packageserver: " + packageserver
                        slackmsg += "\n- token: " + name
                        slackmsg += "\n- app_prefix: " + app_prefix
                        slack_message(slackmsg)
                      
                        if(output_command("wget " + packageserver+app_prefix + "/" + reponame + "/" + randomnumber + ".pkg -P deploypackages")!=True):
                                return None
                       
                        output_command("mkdir " + packagedir+app_prefix)                       
                        output_command("pkill -f " + runningname)
                         
                        if(output_command("tar -xzvf deploypackages/" +randomnumber + ".pkg -C " + packagedir+app_prefix)!=True):
                                return None

                        
                        slack_message("change dir: "+packagedir+app_prefix)
                        os.chdir(packagedir+app_prefix)
                        cmdstr="nohup ./"+runningname
                        if argstr!="":
                                cmdstr+=" "+argstr
                        cmdstr+=" &"
                        slack_message(cmdstr)
                        try:
                                os.system(cmdstr)

                                slack_message("Deploy SUCCESS")
                        except Exception as detail:
                                slack_message("ERROR:" +str(detail))
                        os.chdir(maindir)                        
                        output_command("rm -f -R deploypackages")
                       
                return 'Hello, ' + name + '!'
if __name__ == "__main__":
        app.run() 