# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/ajroetker/nochlehead

# Build the dujour command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get gopkg.in/mgo.v2
RUN go install github.com/ajroetker/nochlehead

# Run the dujour command by default when the container starts.
CMD ["/go/bin/nochlehead", "-log=/go/src/github.com/ajroetker/nochlehead/nochlehead.log"]

# Document that the service listens on port 8080.
EXPOSE 8080
