ARG BASE_IMAGE

FROM composer as composer

FROM $BASE_IMAGE

RUN apk add npm git openjdk8-jre-base bash

COPY --from=composer /usr/bin/composer /usr/bin/composer

RUN apk add --virtual .build-deps --no-cache --update autoconf file g++ gcc libc-dev make pkgconf re2c zlib-dev && \
    pecl install xdebug && \
    docker-php-ext-enable xdebug && \
    apk del .build-deps