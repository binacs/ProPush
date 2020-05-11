#!/bin/bash

sudo cp ./bin/propushd /usr/sbin/propush

cat <<EOF | sudo tee /etc/systemd/system/propush.service
[Unit]
Description=ProPush
Documentation=https://github.com/BinacsLee/ProPush

[Service]
ExecStart=/usr/sbin/propush \\
  -job=binacsProPush \\
  -instance=${hostname} \\
  -endpoint=${endpoint}
Restart=on-failure
RestartSec=5
[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable propush
sudo systemctl start propush