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
                slackmsg="command '{}' return with error (code {}): {}".format(e.cmd, e.returncode, e.output)
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
                        packagename=querystr['pn'][0]
                        runningname=packagename.replace('-','')
                        argstr=querystr['ag'][0]
                        repousername=querystr['rpun'][0]
                        randomnumber=querystr['rn'][0]
                        packageserver=querystr['sv'][0]
                        packagedir=maindir+"/"+repousername+"_"+packagename
                        slackmsg+=packagename+" <=="
                        slackmsg += packagename + " from " + packageserver + " <=="
                        slackmsg += "\n- packagedir: " + packagedir
                        slackmsg += "\n- rannum: " + randomnumber
                        slackmsg += "\n- argstr: " + argstr
                        slackmsg += "\n- packageserver: " + packageserver
                        slackmsg += "\n- token: " + name
                        slack_message(slackmsg)
                      
                        if(output_command("wget " + packageserver + "packages/" + packagename + "/" + randomnumber + ".pkg -P deploypackages")!=True):
                                return None
                       
                        output_command("mkdir " + packagedir)                       
                        output_command("pkill -f " + runningname)
                         
                        if(output_command("tar -xzvf deploypackages/" +randomnumber + ".pkg -C " + packagedir)!=True):
                                return None

                        slackmsg += "\n- Run ... "+"\n"
                        os.chdir(packagedir)
                        cmdstr="./"+runningname
                        if argstr!="":
                                cmdstr+=" "+argstr
                        cmdstr+=" &"
                        slack_message(cmdstr)
                        try:
                                os.system(cmdstr)
                                slackmsg += "\n- Deploy SUCCESS "
                                slack_message(slackmsg)
                        except Exception as detail:
                                slackmsg+=str(detail)
                        os.chdir(maindir)
                        slackmsg += "\n- remove deploypackages: "

                        if(output_command("rm -f -R deploypackages")!=True):
                                return None
                       
                return 'Hello, ' + name + '!'
if __name__ == "__main__":
        app.run() 