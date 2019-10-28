FROM nginx:latest

ADD conf/* /etc/nginx/
ADD credentials_md5 /etc/nginx/includes/creds.passwd

EXPOSE 8082
