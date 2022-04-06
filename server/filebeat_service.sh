sudo su
cat <<'EOF' > /etc/filebeat/inputs.d/<filebeat_server_name>.yml
# Service <filebeat_artifact_id> config
- type: log
  enabled: true
  paths:
    - /home/<filebeat_user_name>/logs_<filebeat_artifact_id>-running/json/<filebeat_artifact_id>.log
  encoding: utf-8
  fields:
    name: "<filebeat_server_name>"
    vili: "running"
    tags: ["<filebeat_artifact_id>"]
- type: log
  enabled: true
  paths:
    - /home/<filebeat_user_name>/logs_<filebeat_artifact_id>-test/json/<filebeat_artifact_id>.log
  encoding: utf-8
  fields:
    name: "<filebeat_server_name>"
    vili: "test"
    tags: ["<filebeat_artifact_id>"]
- type: log
  enabled: true
  paths:
    - /home/<filebeat_user_name>/logs_vili/json/vili.log
  encoding: utf-8
  fields:
    name: "<filebeat_server_name>"
    vili: "vili"
    tags: ["<filebeat_artifact_id>-vili"]
EOF
exit
sudo service filebeat restart
