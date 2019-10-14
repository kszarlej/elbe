FROM nginx:latest

ADD conf/* /etc/nginx/

EXPOSE 8082
