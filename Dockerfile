FROM golang:1.8

COPY . /go/src/github.com/cd1/motofretado-server
RUN go install github.com/cd1/motofretado-server/...

EXPOSE 8080
ENTRYPOINT ["motofretado-server"]
CMD ["--debug", "--port", "8080"]
