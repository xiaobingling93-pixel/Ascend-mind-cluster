FROM ubuntu:18.04

RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&\
    usermod root -s /usr/sbin/nologin

COPY ./clusterd /usr/local/bin
RUN chmod 555 /usr/local/bin/clusterd &&\
    chmod 750 /home/hwMindX &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

# hwMindX is used as the default user of the container
USER hwMindX