FROM tutum/dnsutils

# Install any other tools desired with apt (tutum image is based on ubuntu)
RUN apt-get update \
    && apt-get install -yq curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists

ADD start.sh /
RUN chmod +x /start.sh
CMD ["/start.sh"]