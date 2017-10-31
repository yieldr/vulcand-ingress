FROM debian
COPY bin/vulcand-ingress-linux-amd64 /usr/local/bin/vulcand-ingress
CMD ["vulcand-ingress"]
