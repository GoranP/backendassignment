FROM ubuntu:xenial


RUN mkdir -p /opt/weblayer
ADD weblayer /opt/weblayer/

EXPOSE 9999

ENTRYPOINT [""]
CMD /opt/weblayer/weblayer --redis redis --port 9999


