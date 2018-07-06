FROM debian:stable

RUN apt-get update && apt-get install -y ca-certificates

COPY main /main

CMD ["/main"]
