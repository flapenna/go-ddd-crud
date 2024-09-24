FROM mongo:7.0
RUN openssl rand -base64 756 > /etc/mongo-keyfile
RUN chmod 400 /etc/mongo-keyfile
RUN chown mongodb:mongodb /etc/mongo-keyfile