FROM golang:1.17-alpine

WORKDIR /app

COPY utils ./utils/

COPY core/go.mod ./core/
# COPY core/go.sum ./core/
RUN cd core && go mod download

COPY core/src/*.go ./core/

RUN cd core && go build -o /core .

EXPOSE 9000

ENV PORT=9000

CMD [ "/core" ]