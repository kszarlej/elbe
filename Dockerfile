FROM nginx:latest

ADD conf/* /etc/nginx/
ADD credentials_md5 /etc/nginx/includes/creds.passwd
RUN mkdir -p /var/www \
    && echo "<html><head></head><body><h1>Hello</h1></body></html>" > /var/www/index.html

EXPOSE 8082