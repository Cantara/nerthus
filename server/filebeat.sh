sudo rpm --import https://packages.elastic.co/GPG-KEY-elasticsearch
sudo wget -O /etc/yum.repos.d/filebeat.repo https://gist.githubusercontent.com/kiversen/b5bd483fa1939190a84d02e9b4c9f4cb/raw/9b5f81aec6f89d5b8e782a5bdfdccfb54a8ba22e/filebeat.repo
sudo yum install -y filebeat
sudo chkconfig --add filebeat
sudo service filebeat restart
sudo su
cat <<'EOF' > /etc/filebeat/filebeat.yml
# ============================== Filebeat inputs ===============================

filebeat.inputs:
  enabled: true
  path: inputs.d/*.yml
# ================================== Outputs ===================================

# Configure what output to use when sending the data collected by the beat.

# ---------------------------- Elasticsearch Output ----------------------------
output.elasticsearch:
  enabled: true
  setup.template.enabled: false
  hosts: ["<filebeat_url>"]

  # Set gzip compression level.
  #compression_level: 0

  # Configure escaping HTML symbols in strings.
  #escape_html: false

  # Protocol - either `http` (default) or `https`.
  #protocol: "https"
  username: "anything"
  password: "<filebeat_password>"

  # Number of workers per Elasticsearch host.
  worker: 1

processors:
  - add_host_metadata:

# ================================== DEBUG ===================================

logging:
  level: info
  to_files: true
  to_syslog: false
  files:
    path: /var/log/filebeat
    name: filebeat.log
    keepfiles: 3
EOF
exit
sudo service filebeat restart
if ! sudo ls /etc/filebeat/inputs.d &> /dev/null ; then
  sudo mkdir /etc/filebeat/inputs.d
fi

