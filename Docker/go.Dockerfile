FROM golang:1.11-alpine



# # Build
# RUN apk add wget nginx supervisor curl git mercurial \
#     && go get github.com/fatih/color \
#     && go get gopkg.in/mgo.v2 \
#     && go get gopkg.in/mgo.v2/bson \
#     && go get github.com/gorilla/mux \
#     && go get github.com/BurntSushi/toml \
#     && go get github.com/gorilla/sessions \
#     && go get github.com/tomasen/realip \
#     && go get golang.org/x/crypto/bcrypt \
#     && go get github.com/tomasen/realip \
#     && go get github.com/dgrijalva/jwt-go \
#     && go get github.com/rs/xid \
#     && go get github.com/influxdata/influxdb/client/v2

# ENV SRCDIR=/build/railway
# ENV WRKDIR=/opt/railway
# ENV PUBDIR=/var/www/dev/railway

# #Create foldder and copy files
# RUN mkdir -p $SRCDIR \
#     && mkdir -p $WRKDIR \
#     && mkdir -p $PUBDIR
# ADD ./ $SRCDIR

# # Configure NGINX
# ADD conf/nginx.conf /etc/nginx/nginx.conf
# RUN set -x ; \
#   addgroup -g 82 -S www-data ; \
#   adduser -u 82 -D -S -G www-data www-data && exit 0 ; exit 1

# # Configure supervisord
# ADD conf/supervisord.conf /etc/supervisord.conf
# RUN mkdir -p /var/log/supervisor/

# # Build Railway
# RUN go build -o railway /build/railway/src/*.go \
#     && cp railway $WRKDIR

# # Move server
# ADD ./public $WRKDIR/public
# ADD ./public $PUBDIR/public
# ADD ./env/docker.linux.sh $WRKDIR/env.sh
# ADD ./startServerDocker.sh $WRKDIR/startServer.sh
# RUN chmod +x $WRKDIR/startServer.sh
# ADD ./conf/docker.linux.toml $WRKDIR/conf/railway.toml

# EXPOSE 27017
# EXPOSE 3895

# # CMD ["sh","-c","/opt/railway/railway"]
# WORKDIR /
# CMD		["supervisord", "--nodaemon", "--configuration", "/etc/supervisord.conf"]