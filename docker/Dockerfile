FROM alpine:latest
LABEL Osiloke Emoekpere <me@osiloke.com>
RUN apk update && apk add --no-cache bash ca-certificates && rm -rf /var/cache/apk/*
ADD mailspace-bot /app/mailspace-bot
RUN chmod +x /app/mailspace-bot 
WORKDIR /app
VOLUME /app/session
COPY ./actions.json /app/
WORKDIR /app
CMD ["/app/mailspace-bot"]