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
  - [How to deploy](#how-to-deploy)
  - [How to use](#how-to-use)

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
