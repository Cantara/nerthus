#!/bin/bash
read -p 'Username: ' uservar
read -sp 'Password: ' passvar
user=$uservar":"$passvar
url="<url>"
server="<server>"
key="<key>"

key="{\"key\":\"$key\"}"

var=$(curl \
  --header "Content-Type: application/json" \
  --request POST \
  --data-raw $key \
  -u $user \
  ${url}/key)

pem_name=$(echo $var | jq .key.pem_name -r)
scope="$(echo $var | jq .scope -r)"
sh_name="${server}.sh"

var_dns=$(curl \
  --header "Content-Type: application/json" \
  --request GET \
  -u $user \
  ${url}/dns/${scope}/${server})

dns="$(echo $var_dns | jq .public_dns -r)"

echo $var | jq .key.material -r > $pem_name
chmod 0600 $pem_name
ssh ec2-user@$dns -i $pem_name
rm -f $pem_name
