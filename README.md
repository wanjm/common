# 说明
1. 本项目是go功能的基础支持项目，设计go 服务端项目的基础细化结构，放在这里	
2. 为了让其他工程能拉到本包，需要在本项目的git->Settings->Deploy keys -> privately accessible deploy keys 中的pm2deploy 点击enable；
3. 然后宿主项目的Dockerfile文件中 用"FROM docker.plaso.cn/golang:1.22.6-alpine-buildv1 as builder"
