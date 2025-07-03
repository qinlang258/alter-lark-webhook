FROM  registry.cn-zhangjiakou.aliyuncs.com/jcrose-devops/ci-golang:1.23

WORKDIR /app/
ADD main /app/main
COPY manifest/config/config.yaml /app/manifest/config/config.yaml

ENTRYPOINT /app/main --gf.gcfg.file=/app/manifest/config/config.yaml
