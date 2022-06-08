# Nerthrus

[![Build Status](https://jenkins.entraos.io/buildStatus/icon?job=Cantara-Nerthus%2Fmain)](https://jenkins.entraos.io/job/Cantara-Nerthus/job/main/)

Nerthrus is a service to orchistate new services. It is designed around running services on AWS in EC2 instances. Every action is going to be idempotent to the best of its abilities. And Nerthrus will handle sub deployments aswell as fresh 'servers' / instances. Excamples of subdeployments will follow in a later section.

## TL;DR

Nerthrus orchistrates AWS EC2 instances and the surounding technology needed to run services of that nature.

## Inspiration

## Contents

- [Nerthrus](#nerthrus)
  - [TL;DR](#tl;dr)
  - [Inspiration](#inspiration)
  - [Contents](#contents)
  - [The env file](#the-env-file)
    - [Slack](#slack)
      - [Creating a Slack app](#creating-a-slack-app)
  - [How to deploy](#how-to-deploy)
  - [How to use](#how-to-use)

## The env file

There are some important things to know about the env file. The most important parts are the ami image, aeskey, username and password, port, region, slack_channel_secret, slack_channel_status, slack_token.

### Get ami

ami is a refferance to an aws image, this is unique per aws region and we sudgest you go into the aws dashboard, ec2 and create instace. There we sudgest that you use the newes Amazone linux. You will se ami be listed on the page where you select linux image. Copy that and use that one. It should be updated from time to time.

### Generate AES key

Nerthus provides a way to generate aes keys that are base64 encoded. This is needed to encrypt secrets sent from nerthus. You can either use the go runnable in nerthus/crypto/cmd or Nerthus itself with `nerhus -genAES`
Please do not loose or change this key unless you know what you are doing.

### Username and passwor

The username and password can be changed at any time. This is used to get access to all nerthus apis. The encrypted keys are "worthless" without the username and password, but not locked to any specific username or passord.

### Region

Region is the aws region that Nerthus should opperate in. Nerthus is confined to one region to limit its scope. ex `eu-north-1`

### Slack

Slack will need one api token to send messages. This should be unique to Nerthus. The reason for this is, if it gets leaked, someone could read out all the messages that Nerthus has sent. This includes the encypted keys. It's not the end of the world, but definetly not good. If you want one per env that is also okay.

#### Creating a Slack app

  1. Go to [Slack API](https://api.slack.com/)
  2. Click create an app
  3. Use from scratch
  4. Give it a name and select where to deploy it, what workspace.
  5. Go to Add features and functionality -> Permissions. Aka, OAuth & Permissions
  6. Scroll down to Scpoes and add an OAuth Scope and select `chat:write`
  7. Scroll up and click Install to Workspace and allow
  8. You will now find a token under OAuth & Permissions -> OAuth Tokens for Your Workspace
  Copy this and use it as slack_token

#### Creating channels

You will need two channels one for status messages and one for secrets. The status messages does not contain strictly sensitive information, so it could be adviced to give all developers and atleast all Nerthus users access to this one. The secrets channel should deffinetly be private.
The Nerthus slack app / api have to be added to the channels manually. We don't want nerthus anywhere it shouldn't be.
When you have created your channels and added your App to them you can find their channel ids and add them to their respective env vars, slack_channel_status slack_channel_secret.

## How to deploy

The sudgested way to deploy Nerthus is with Buri auto updater and our pre compiled linux binaries. The following also includes a crontab that handles updating and restarts on craches.
First we download buri and create a link so it can run normally from first run. Then we add the crontab that does the rest.
You will also have to add a .env file with all nerthus configs. Use the template one as a base. tmp.env

```shell
curl --fail --show-error --silent -o "buri-v0.3.5" "https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.3.5/buri-v0.3.5"
ln -s "buri-v0.3.5" "buri"
chmod +x "buri"
cat <<'EOF' > ~/CRON
MAILTO=""
*/6 * * * * ./buri -a buri -g no/cantara/gotools > /dev/null
*/6 * * * * ./buri -a nerthus -g no/cantara/gotools -r > /dev/null
EOF

crontab ~/CRON
```

## How to use

### Rest API

Nerthrus is rest based, but will also work with HTML form and XML data as additions to the rest standard JSON. However Nerthrus will only return JSON.
All enpoints within this API is secured

#### Endpoints

For examples of these endoint look at the .sh files

##### PUT /nerthus/server/:application/*server

This is the main endpoint of this service. It needs a application name but can also take a server name to override exactly what server you want to interact with. In the case where you provide a servername the server has to allready excist and it has to be exact.

This service will not create new Loadbalancers or listeners as setting up the https certificates should be up to your discretion.

The body of this request has to be a object that has different requrements for what you want to do. The following examples includes only requred data.

* For a full clean standalone instance

  1. NO servername in uri
  2. Int port the service is going to use
  3. Uri path, for nerthrus that would be `"path":"nerthus"`
  4. Loadbalancer listener ARN, this is the unique identifier of the listener you want to serve the service
  5. Loadbalancer securitygroup id, this is the id of the security group the loadbalancer uses to access the service

  ```json
  {
    "port": 18080,
    "path": "nerthus",
    "elb_listener_arn": "arn:aws:elasticloadbalancing:us-west-2:493376950721:listener/app/devtest-events2-lb/a3807cba101b280b/90abaa841820e9b2",
    "elb_securitygroup_id": "sg-1325864d"
  }
  ```

If you have enabled slack this endpoint will log every action done to bouth the logout and the slack channal that is specified. And at the end of the request, in addition to returning the key it will send the key in slack.

If there at any point is a error during the request the server will automatically clean up all the changes that it has done.

##### POST /nerthus/key

This endpoint takes a body with a key in it and returns the decrypted key so you can manually log on to the server.
