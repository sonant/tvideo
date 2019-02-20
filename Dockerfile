FROM ubuntu

ADD main /opt/app

RUN apt-get update && \
    apt-get install -y ffmpeg && \
    chmod +x /opt/app

EXPOSE 8080

CMD ./opt/app