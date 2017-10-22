FROM debian
COPY bin/vulcand-ingress /usr/local/bin
CMD ["vulcand-ingress"]
