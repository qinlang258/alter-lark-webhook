FROM  registry.cn-zhangjiakou.aliyuncs.com/jcrose-devops/ci-golang:1.23

WORKDIR /app/
ADD main /app/main

ENTRYPOINT /app/main
