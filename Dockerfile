FROM registry.cn-hangzhou.aliyuncs.com/spacexnice/netdia:latest


COPY hub-console /app/
COPY js /app/js
COPY pages /app/pages
COPY fonts /app/fonts
COPY css /app/css
COPY entrypoint.sh /app/

ENTRYPOINT ["/app/hub-console"]