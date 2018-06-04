# This Dockerfile is used to build an image which is used in circle CI
# to run the the following check or validation:
# markdownlint
# misspell
# ShellCheck 
FROM ubuntu:16.04

RUN apt-get update \
    && apt-get install -y rubygems git curl \
    && gem install rake \
    && gem install bundler \
    && apt-get clean   

# install markdownlint
RUN git clone https://github.com/markdownlint/markdownlint.git \
    && cd markdownlint && git checkout v0.4.0 && rake install

# install misspell
RUN git clone https://github.com/client9/misspell.git \
    && cd misspell && curl -L -o ./install-misspell.sh https://git.io/misspell && sh ./install-misspell.sh

RUN mv misspell/bin/misspell /usr/local/bin/

# install shellcheck
RUN apt-get install shellcheck
