#!/bin/bash
for port in 7000 7001 7002 7003 7004 7005
do
	redis-server "./redis-cluster/$port/redis.conf"
done
