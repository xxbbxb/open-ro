FROM ubuntu:20.04
COPY ./ /rathena/
WORKDIR /rathena/
ENV DEBIAN_FRONTEND="noninteractive" \ 
  LANG="en_US.UTF-8" \
  LANGUAGE="en_US:en" \
  LC_ALL="en_US.UTF-8" \
  apt_get_install="apt-get install -y --no-install-recommends"
RUN apt-get -qy update && \
  $apt_get_install apt-transport-https ca-certificates && \
  $apt_get_install git gettext make libmysqlclient-dev zlib1g-dev libpcre3-dev gcc g++ && \
  apt-get clean && \
  rm -rf /tmp/* /var/tmp/* &&   rm -rf /var/lib/apt/lists/*
RUN ./configure --enable-prere=yes --enable-64bit=yes --enable-epoll --enable-packetver=20190605 && make server

