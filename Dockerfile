FROM golang:1.16.2

RUN go get -u github.com/kisielk/errcheck
RUN go get -u golang.org/x/lint/golint
RUN go get -u honnef.co/go/tools/cmd/staticcheck
RUN go get -u github.com/axw/gocov/gocov
RUN go get -u github.com/securego/gosec/cmd/gosec
RUN go get -u github.com/ory/go-acc
RUN go get -u github.com/client9/misspell/cmd/misspell
RUN go get -u github.com/gordonklaus/ineffassign
RUN go get github.com/fzipp/gocyclo
