#!/bin/sh

# backend software source
pushgatewaySource="https://github.com/prometheus/pushgateway/releases/download/v1.2.0/pushgateway-1.2.0.linux-amd64.tar.gz"
prometheusSource="https://github.com/prometheus/prometheus/releases/download/v2.21.0-rc.1/prometheus-2.21.0-rc.1.linux-amd64.tar.gz"
alertmanagerSource="https://github.com/prometheus/alertmanager/releases/download/v0.21.0/alertmanager-0.21.0.linux-amd64.tar.gz"
node_exporter="https://github.com/prometheus/node_exporter/releases/download/v1.0.1/node_exporter-1.0.1.linux-amd64.tar.gz"

# directory
current=$(cd $(dirname $0); pwd)
workspace="$current/workspace"
promspace="$workspace/source"
MonitorDir="$workspace/monitor"
BackendDir="$workspace/backend"
ConfigDir="$workspace/config"

mkdirFunc(){
    mkdir $workspace
    mkdir $promspace
    mkdir $BackendDir
    mkdir $MonitorDir
    mkdir $ConfigDir
    mkdir "$ConfigDir/rules"
}

downloadFunc(){
    sudo wget -O pushgatewayPack.tar.gz $pushgatewaySource
    sudo wget -O prometheusPack.tar.gz $prometheusSource
    sudo wget -O alertmanagerPack.tar.gz $alertmanagerSource
}

tarcopyFunc(){
    sudo tar -zxvf pushgatewayPack.tar.gz -C $promspace
    sudo tar -zxvf prometheusPack.tar.gz -C $promspace
    sudo tar -zxvf alertmanagerPack.tar.gz -C $promspace

    sudo cp $promspace/pushgateway*/pushgateway $BackendDir
    sudo cp $promspace/prometheus*/prometheus $BackendDir/
    sudo cp $promspace/alertmanager*/alertmanager $BackendDir/

    sudo cp $promspace/prometheus*/prometheus.yml $ConfigDir/
    sudo cp $promspace/alertmanager*/alertmanager.yml $ConfigDir/
}

makescriptFunc(){
cat <<EOF | sudo tee "$workspace/start.sh"
#!/bin/bash
rm -rf log
mkdir log
source /etc/profile

#echo "ready to start mockMetrics"
#nohup ./mockdata/mockMetrics >> ./log/mock.log 2>&1 &

echo "ready to start pushgateway"
nohup ./backend/pushgateway >> ./log/pushgateway_error.log 2>&1 &

echo "ready to start prometheus"
nohup ./backend/prometheus --config.file ./config/prometheus.yml --web.enable-lifecycle >> ./log/prometheus_error.log 2>&1 &

echo "ready to start altermanager"
nohup ./backend/alertmanager --config.file ./config/alertmanager.yml >> ./log/alertmanager_error.log 2>&1 &

#echo "ready to start alteragent"
#nohup ./monitor/alertagent -config.file ./config/alertagent.yml >> ./log/alertagent_error.log 2>&1 &

#echo "ready to start exporter"
#nohup ./monitor/exporter -configfile ./config/exporter.toml >> ./log/exporter_error.log 2>&1 &
EOF

cat <<EOF | sudo tee "$workspace/stop.sh"
#!/bin/bash
rm -rf data

pid_prome=$(ps -ef | grep prometheus | awk '{print $2}')
for i in $pid_prome; do
	echo "kill prometheus $pid_prome"
	kill -9 $i
done
pid_push=$(ps -ef | grep pushgateway | awk '{print $2}')
for i in $pid_push; do
        echo "kill pushgateway $i"
        kill -9 $i
done
pid_alertm=$(ps -ef | grep alertmanager | awk '{print $2}')
for i in $pid_alertm; do
        echo "kill alertmanager $i"
        kill -9 $i
done
pid_exporter=$(ps -ef | grep exporter | awk '{print $2}')
for i in $pid_exporter; do
        echo "kill exporter $i"
        kill -9 $i
done
pid_alertagent=$(ps -ef | grep alertagent | awk '{print $2}')
for i in $pid_alertagent; do
        echo "kill alertagent $i"
        kill -9 $i
done
pid_mock=$(ps -ef | grep mockMetrics | awk '{print $2}')
for i in $pid_mock; do
        echo "kill mock $i"
        kill -9 $i
done
EOF

cat <<EOF | sudo tee "$workspace/clear.sh"
#!/bin/bash
rm -rf ./log ./data ./backend/data
EOF

cat <<EOF | sudo tee "$workspace/reload.sh"
#!/bin/bash
echo "ready to reload prometheus"
curl -X POST http://127.0.0.1:9090/-/reload
echo ""

echo "ready to reload alteragent"
curl -X POST http://127.0.0.1:8011/manager/reload
echo ""

echo "ready to reload exporter"
curl -X POST http://127.0.0.1:8012/manager/reload
echo ""
EOF
}

clearFunc(){
    sudo rm -rf $workspace
}

usage(){
    echo ""
    echo "$0 mkdir"
    echo "$0 download"
    echo "$0 tar"
    echo "$0 scripts"
    echo "$0 clear"
    echo ""
}

case $1 in
    mkdir) mkdirFunc;;
    download) downloadFunc;;
    tar) tarcopyFunc;;
    scripts) makescriptFunc;;
    clear) clearFunc;;
    usage) usage;;
    *)
    mkdirFunc
    downloadFunc
    tarcopyFunc
    makescriptFunc
    ;;
esac

