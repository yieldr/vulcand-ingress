FROM debian
ADD bin/vulcand-ingress /usr/local/bin
CMD ["vulcand-ingress"]
