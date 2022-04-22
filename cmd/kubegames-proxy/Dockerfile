FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD kubegames-proxy /kubegames-proxy
RUN chmod -R 777 /kubegames-proxy

WORKDIR /

ENTRYPOINT ./kubegames-proxy -p 8080