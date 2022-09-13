FROM alpine

# add user
ARG USER=default
RUN adduser -D $USER

# add app
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
ADD cmd/kube-api/kube-api .

# switch user
USER $USER

CMD ["./kube-api"]