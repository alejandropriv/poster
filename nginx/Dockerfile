FROM nginx:alpine

ENV POSTER_API_URL=http://poster-api:8080 \
    NGINX_ENVSUBST_FILTER=POSTER_API_URL

RUN rm /etc/nginx/conf.d/default.conf

COPY web/index.html /usr/share/nginx/html/index.html
COPY nginx/conf/default.conf.template /etc/nginx/templates/default.conf.template

EXPOSE 80