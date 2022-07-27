FROM ubuntu:latest
WORKDIR /app
COPY ./weather-reporter /app/weather-reporter
COPY ./docker-run.sh /app/docker-run.sh
RUN apt-get update -qq  && apt-get install --no-install-recommends -y ca-certificates && rm -rf /var/lib/apt/lists/*
RUN ["chmod", "+x", "/app/docker-run.sh"]
CMD ["/app/docker-run.sh"]

### -------------------------------------------------------------
### for building a smaller image, but fails to run at the momemnt
### -------------------------------------------------------------
# FROM alpine:latest
# WORKDIR /app
# COPY ./weather-reporter /app/weather-reporter
# COPY ./docker-run.sh /app/docker-run.sh
# RUN apk update && apk add ca-certificates && rm -rf /var/lib/apt/lists/*
# RUN ["chmod", "+x", "/app/docker-run.sh"]
# CMD ["/app/docker-run.sh"]