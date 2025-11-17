#!/bin/bash

replive_pid=$(pgrep -f "replive")

bootstrap_rep_pid=$(pgrep -f "bootstrap_rep")

if [ -n "$replive_pid" ]; then
    echo "找到进程 replive，PID 为 $replive_pid"
    kill -9 "$replive_pid"
    echo "进程 replive 已终止"
else
    echo "未找到进程 replive"
fi

if [ -n "$bootstrap_rep_pid" ]; then
    echo "找到进程 bootstrap_rep，PID 为 $bootstrap_rep_pid"
    kill -9 "$bootstrap_rep_pid"
    echo "进程 bootstrap_rep 已终止"
else
    echo "未找到进程 bootstrap_rep"
fi
