FROM nginx:stable-alpine

RUN rm -f /etc/nginx/conf.d/default.conf \
    && mkdir -p /var/www/sites/main

COPY main_site.conf /etc/nginx/conf.d/main_site.conf

VOLUME ["/var/www/sites/main"]

EXPOSE 80

RUN nginx -t