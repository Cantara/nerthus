# Switch to user
./su_to_<username>.sh


# Install semantic-versioning and visuale
curl -s "https://raw.githubusercontent.com/Cantara/visuale/master/agent/scripts/download_and_setup_visuale_reporting.sh" | bash -s

mv ~/scripts/kill-service.sh_template ~/scripts/kill-service.sh
chmod +x ~/scripts/kill-service.sh
rm ~/scripts/*_template

cat <<'EOF' > ~/scripts/semantic_update_service.properties
<semantic_update_service_properties>
EOF

cat <<'EOF' > ~/scripts/start-service.properties
JVM_ARGS=""
EOF

cat <<'EOF' > ~/local_override.properties
<local_override_properties>
EOF

cat <<'EOF' > ~/scripts/reportServiceHealthToVisuale.properties
healthUrl=http://localhost:<port>/<path>/health
reportToUrl1='<health_report_enpoint>'
reportToUrl2='<health_report_enpoint>'
EOF

cat <<'EOF' > ~/scripts/CRON
MAILTO=""
*/6 * * * * ./scripts/download_and_restart_if_new.sh > /dev/null
#*/6 * * * * ./scripts/semantic_update_service.sh > /dev/null
*/6 * * * * ./buri -a buri -g no/cantara/gotools > /dev/null
#*/6 * * * * ./scripts/start-vili.sh > /dev/null
* * * * * ./scripts/reportServiceHealthToVisuale.sh > /dev/null
EOF

ln -s scripts/CRON CRON

crontab ~/CRON

curl --fail --show-error --silent -o "buri-v0.3.5" "https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/buri/v0.3.5/buri-v0.3.5"
ln -s "buri-v0.3.5" "buri"
chmod +x "buri"

cat <<'EOF' > ~/scripts/start-vili.sh
#!/bin/sh
./buri -a vili -g no/cantara/gotools -r &
EOF
chmod +x ~/scripts/start-vili.sh

cat <<'EOF' > ~/scripts/kill-vili.sh
#!/bin/sh
pkill -9 vili
pkill -9 java
rm -rf nohup.out
EOF
chmod +x ~/scripts/kill-vili.sh


cat <<'EOF' > ~/.env
port="<port>"
scheme="http"
endpoint="localhost"
port_range="<port_from>-<port_to>"
identifier="<application>"
log_dir="logs_vili"
properties_file_name="local_override.properties"
port_identifier="server.port"
manualcontrol="false"

entraos_api_uri="https://api-devtest.entraos.io"
slack_channel="C02T3A66D2N"
app_icon="<app_icon>"
env_icon="<env_icon>"
env="<env>"

whydah_uri="https://entrasso-devtest.entraos.io"
whydah_application_name="EntraOS Vili"
whydah_application_id="<vili_whydah_id>"
whydah_application_secret="<vili_whydah_secret>"
EOF

~/scripts/semantic_update_service.sh
./buri -a buri -g no/cantara/gotools
./buri -a vili -g no/cantara/gotools -r
~/scripts/start-vili.sh

# Clear history which contains passwords and secrets
echo '' > ~/.bash_history
history -c
exit
