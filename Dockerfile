FROM debian:stretch-slim

ARG DEBIAN_FRONTEND=noninteractive

COPY Gemfile /Gemfile

    # Install & configure dependencies.
    # Install fluentd and plugins via ruby.
    # Remove build dependencies.
    # Cleanup leftover caches & files.
    # Ensure fluent has enough file descriptors
RUN BUILD_DEPS="make gcc g++ libc6-dev ruby-dev libffi-dev" \
    && apt-get update \
    && apt-get install -y --no-install-recommends $BUILD_DEPS \
                                                  ca-certificates \
                                                  libjemalloc1 \
                                                  ruby \
    && echo 'gem: --no-document' >> /etc/gemrc \
    && gem install --file Gemfile \
    && apt-get purge -y --auto-remove \
                     -o APT::AutoRemove::RecommendsImportant=false \
                     $BUILD_DEPS \
    && apt-get clean -y \
    && rm -rf /var/cache/debconf/* \
              /var/lib/apt/lists/* \
              /var/log/* \
              /tmp/* \
              /var/tmp/* \
              /usr/lib/ruby/gems/*/cache/*.gem
#    && ulimit -n 65536

# Copy the Fluentd configuration file for logging Docker container logs.
COPY fluent.conf /etc/fluent/fluent.conf
COPY run.sh /run.sh

# Expose prometheus metrics.
EXPOSE 80

ENV LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libjemalloc.so.1

# Start Fluentd to pick up our config that watches Docker container logs.
CMD ["/run.sh"]