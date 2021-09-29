FROM golang:1.17-alpine

WORKDIR /app

COPY utils ./utils/

COPY authorization/go.mod ./auth/
COPY authorization/go.sum ./auth/
RUN cd auth && go mod download

COPY authorization/src/*.go ./auth/

RUN cd auth && go build -o /auth .

EXPOSE 7000

ENV PORT=7000

CMD [ "/auth" ]