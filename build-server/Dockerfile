FROM ubuntu:focal

WORKDIR /app

// Install curl tool
RUN sudo apt install ca-certificates curl gnupg git wget

// Install go and setup path
RUN wget -c https://go.dev/dl/go1.23.1.linux-amd64.tar.gz \
&& sudo tar -xzf go1.23.1.linux-amd64.tar.gz -C /usr/local \
&& echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc \
&& source ~/.bashrc

// Install Node JS and npm
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | sudo bash - \
&& sudo apt-get install -y nodejs

COPY /app


ENTRYPOINT ["entry.sh"]
