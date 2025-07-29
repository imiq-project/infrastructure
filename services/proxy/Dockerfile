FROM kong:3.9-ubuntu
USER root
RUN apt-get update && apt-get install -y gettext-base && apt-get clean
USER kong
COPY kong.yml /etc/kong/kong.yml
ENV KONG_DATABASE=off
ENV KONG_DECLARATIVE_CONFIG=/etc/kong/kong-subst.yml
ENV KONG_PROXY_ACCESS_LOG=/dev/stdout
ENV KONG_ADMIN_ACCESS_LOG=/dev/stdout
ENV KONG_PROXY_ERROR_LOG=/dev/stderr
ENV KONG_ADMIN_ERROR_LOG=/dev/stderr
CMD ["/bin/sh", "-c", "envsubst < /etc/kong/kong.yml > /etc/kong/kong-subst.yml && kong start"]
