FROM golang:latest
MAINTAINER Kenjiro Nakayama <nakayamakenjiro@gmail.com>

# Set up workdir
WORKDIR /go/src/github.com/alternative-storage/torus

# Add and install torus
ADD . .
RUN make vendor
RUN go install -v github.com/alternative-storage/torus/cmd/torusd
RUN go install -v github.com/alternative-storage/torus/cmd/torusctl
RUN go install -v github.com/alternative-storage/torus/cmd/torusblk

# Expose the port and volume for configuration and data persistence.
VOLUME ["/data", "/plugin"]
EXPOSE 40000 4321

CMD ["./entrypoint.sh"]
