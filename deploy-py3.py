import web
from urllib.parse import urljoin
import urllib
import json
import subprocess
import os
import requests
import sys
import slack
urls = (
    '/deploy/(.*)', 'hello'
)
#python2.7 deploy.py 0.0.0.0 xxx xxx &
app = web.application(urls, globals())
token = sys.argv[2]
slacktoken = sys.argv[3]
homedir=os.getcwd()
sc = slack.WebClient(token=os.environ[slacktoken])
#sc.api_call("chat.postMessage",channel="#serverdeploy",text="-- started deploy server with token: "+token+" --")
sc.chat_postMessage(channel="#serverdeploy",text="-- started deploy server with token: "+token+" --")
class hello:
        def POST(self, name):                
                output=''
                if name==token:
                        #runname=name.replace("/","_")
                        runname=name
                        
                        slackmsg="==> Start Deploying "
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
                        sc.api_call("chat.postMessage",channel="#serverdeploy",text=slackmsg)
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
                        os.chdir(packagedir)
                        cmdstr="nohup ./"+runningname
                        if argstr!="":
                                cmdstr+=" "+argstr
                        cmdstr+=" &"
                        sc.api_call("chat.postMessage",channel="#serverdeploy",text=cmdstr)
                        try:
                                os.system(cmdstr)
                                slackmsg += "\n- Deploy SUCCESS "
                                sc.api_call("chat.postMessage",channel="#serverdeploy",text=slackmsg)
                        except Exception as detail:
                                slackmsg+=str(detail)
                        os.chdir(maindir)
                        slackmsg += "\n- remove deploypackages: "
                        try:
                                slackmsg += subprocess.check_output("rm -f -R deploypackages",shell=True,stderr=subprocess.STDOUT)
                        except subprocess.CalledProcessError as e:
                                slackmsg+="command '{}' return with error (code {}): {}".format(e.cmd, e.returncode, e.output)
                        
                return 'Hello, ' + name + '!'
if __name__ == "__main__":
        app.run() 