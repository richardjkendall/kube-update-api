FROM alpine

# add user
ARG USER=default
RUN adduser -D $USER

# add app
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
ADD cmd/kube-api/kube-api .

# install libc6 compatability library
RUN apk add --no-cache libc6-compat

# switch user
USER $USER

CMD [ "/usr/src/app/kube-api" ]