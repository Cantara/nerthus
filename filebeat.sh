 sudo rpm --import https://packages.elastic.co/GPG-KEY-elasticsearch
 sudo wget -O /etc/yum.repos.d/filebeat.repo https://gist.githubusercontent.com/kiversen/b5bd483fa1939190a84d02e9b4c9f4cb/raw/9b5f81aec6f89d5b8e782a5bdfdccfb54a8ba22e/filebeat.repo
 sudo yum install -y filebeat
 sudo chkconfig --add filebeat
 sudo service filebeat restart
 sudo wget -O /etc/filebeat/filebeat.yml https://raw.githubusercontent.com/entraeiendom/entraos-config/main/log/humio/filebeat/devtest/filebeat.entraos.information.yml?token=AA44R65GKYLJ2IBY25IVOE3BV6A5Y
 sudo service filebeat restart
