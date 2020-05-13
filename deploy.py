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
#nohup python deploy.py 0.0.0.0 xgdedkillaccnqweoiurpelksfcvnbsdw slacktoken serverdev & 
app = web.application(urls, globals())
token = sys.argv[2]
slacktoken = sys.argv[3]
slackchannel = sys.argv[4]
homedir=os.getcwd()
hostname = socket.gethostname()
sc = SlackClient(slacktoken)
sc.api_call("chat.postMessage",channel="#"+slackchannel,text="-- started server with token: "+token+" --")
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
                        sc.api_call("chat.postMessage",channel="#"+slackchannel,text=slackmsg)
                        slackmsg = "==> Package " + packagename + " Deployed <=="
                        slackmsg += "\n- Pull package " + randomnumber + "..."+"\n"
                        try:
                                subprocess.check_output("wget " + packageserver + "packages/" + packagename + "/" + randomnumber + ".pkg -P deploypackages",shell=True,stderr=subprocess.STDOUT)
                        except subprocess.CalledProcessError as e:
                                slackmsg+="command '{}' return with error (code {}): {}".format(e.cmd, e.returncode, e.output)
                        
                        try:
                                slackmsg += subprocess.check_output("mkdir " + packagedir,shell=True,stderr=subprocess.STDOUT)
                        except subprocess.CalledProcessError as e:
                                slackmsg+="command '{}' return with error (code {}): {}".format(e.cmd, e.returncode, e.output)
                        
                        slackmsg += "\n- kill " + runningname+"\n"
                        try:
                                slackmsg += subprocess.check_output("pkill -f " + runningname,shell=True,stderr=subprocess.STDOUT)
                        except subprocess.CalledProcessError as e:
                                slackmsg+="command '{}' return with error (code {}): {}".format(e.cmd, e.returncode, e.output)
                        
                        slackmsg += "\n- Extract " + packagename +"\n"
                        try:
                                slackmsg += subprocess.check_output("tar -xzvf deploypackages/" + randomnumber + ".pkg -C " + packagedir,shell=True,stderr=subprocess.STDOUT)
                        except subprocess.CalledProcessError as e:
                                slackmsg+="command '{}' return with error (code {}): {}".format(e.cmd, e.returncode, e.output)
                        

                        slackmsg += "\n- Run ... "+"\n"
                        sc.api_call("chat.postMessage",channel="#"+slackchannel,text=slackmsg)
                        os.chdir(packagedir)
                        cmdstr="./"+runningname
                        if argstr!="":
                                cmdstr+=" "+argstr
                        cmdstr+=" &"
                        sc.api_call("chat.postMessage",channel="#"+slackchannel,text=cmdstr)
                        try:
                                os.system(cmdstr)                               
                                sc.api_call("chat.postMessage",channel="#"+slackchannel,text="Deploy SUCCESS")
                        except Exception as detail:
                                sc.api_call("chat.postMessage",channel="#"+slackchannel,text=string(detail))
                                sc.api_call("chat.postMessage",channel="#"+slackchannel,text="QUIT")
                        os.chdir(maindir)
                        slackmsg += "\n- remove deploypackages: "
                        try:
                                slackmsg += subprocess.check_output("rm -f -R deploypackages",shell=True,stderr=subprocess.STDOUT)
                        except subprocess.CalledProcessError as e:
                                slackmsg+="command '{}' return with error (code {}): {}".format(e.cmd, e.returncode, e.output)
                        
                return 'Hello, ' + name + '!'
if __name__ == "__main__":
        app.run() 