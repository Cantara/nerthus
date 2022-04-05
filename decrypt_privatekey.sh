url="localhost:3030/nerthus" #$1
server="demo-entraos-infra-1" #$2
user="sindre:pass" #$3
key="" #$4

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
chmod 0400 $pem_name
echo "ssh ec2-user@$dns -i $pem_name" > $sh_name
chmod +x $sh_name
