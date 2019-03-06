#!/bin/bash
ip=`ifconfig eth0|grep "inet addr"|awk '{print $2}'|awk -F: '{print $2}'`
path=""
token=""

clean_log(){
    for port in `ls /u01/mysql|grep 33`; do
        slow_path="/u01/mysql/$port/log/slow"
        slow_log="$slow_path/mysql-slow.log"
        bak_log="$slow_path/bak.log"
        cat ${slow_log} >> ${bak_log}
        echo "" > ${slow_log}
        if [[ ${path} == "" ]];then
            path="{\"addr\":\"$ip:$port\",\"log_file\":\"$slow_log\"}"
        else
    	path=${path},"{\"addr\":\"$ip:$port\",\"log_file\":\"$slow_log\"}"
        fi
    done
    check_conf ${path}
}

check_conf(){
    paths=$1
    if [[ ! -f conf.ini ]];then
	cat >> conf.ini <<EOF
[base]
query_time=3
[dingTalk]
token=${token}
[logs]
path=[${paths}]
EOF
    fi
    if [[ `grep PATH conf.ini` != "" ]];then
        sed -i "s#PATH#[$paths]#g" conf.ini
    fi
}

restart_prog(){
    clean_log
    ps -ef|grep slowLogParser|grep -v grep|awk '{print $2}'|xargs kill -9
    nohup ./slowLogParser >> all.log &
}

restart_prog