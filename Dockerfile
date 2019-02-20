FROM centos:7

ADD main /opt/app

RUN yum install -y ffmpeg && \
    chmod +x /opt/app

EXPOSE 8080

CMD ./opt/app