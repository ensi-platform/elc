ARG BASE_IMAGE

FROM composer as composer

FROM $BASE_IMAGE

RUN apk add --virtual .build-deps --no-cache --update autoconf file g++ gcc libc-dev make pkgconf re2c zlib-dev && \
    pecl install xdebug && \
    docker-php-ext-enable xdebug && \
    apk del .build-deps

ARG USER_ID
ARG GROUP_ID

RUN addgroup -g $GROUP_ID developer
RUN adduser -u $USER_ID -S -D -H -G developer developer
RUN addgroup www-data developer
RUN addgroup developer www-data

RUN apk add bash

COPY --from=composer /usr/bin/composer /usr/bin/composer