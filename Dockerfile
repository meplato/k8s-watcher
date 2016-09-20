FROM alpine:3.4
MAINTAINER Oliver Eilhard <oliver.eilhard@meplato.com>
ADD k8s-watcher /k8s-watcher
ENTRYPOINT ["/k8s-watcher"]
