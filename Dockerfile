FROM 61.160.36.122:8080/lightvm:latest

# ��װ���̿�ִ���ļ����� myapp.go ���룩
COPY registry-console /app/
COPY js /app/js
COPY pages /app/pages
COPY fonts /app/fonts
COPY css /app/css
COPY entrypoint.sh /app/

# �����Զ��������
RUN mkdir /etc/service/registry-console
COPY entrypoint.sh /etc/service/registry-console/run
RUN chmod +x /etc/service/registry-console/run

# The entrypoint of lightvm will start everything
# under `/etc/service` as daemon