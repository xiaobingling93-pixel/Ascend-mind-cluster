FROM ubuntu:22.04 AS build

RUN useradd -d /home/hwMindX -u 9000 -m -s  /usr/sbin/nologin hwMindX && \
    usermod root -s /usr/sbin/nologin

COPY ./infer-operator  /usr/local/bin/

RUN chown -R hwMindX:hwMindX /usr/local/bin/infer-operator  &&\
    chmod 500 /usr/local/bin/infer-operator &&\
    chmod 750 /home/hwMindX &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

USER 9000