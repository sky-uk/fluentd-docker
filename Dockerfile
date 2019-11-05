FROM debian:stretch-slim

ARG DEBIAN_FRONTEND=noninteractive

COPY Gemfile /Gemfile

    # Install & configure dependencies.
    # Install fluentd and plugins via ruby.
    # Remove build dependencies.
    # Cleanup leftover caches & files.
    # Ensure fluent has enough file descriptors
RUN BUILD_DEPS="make gcc g++ libc6-dev ruby-dev libffi-dev build-essential git " \
    && apt-get update \
    && apt-get install -y software-properties-common \
    && apt-get install -y --no-install-recommends $BUILD_DEPS \
                                                  ca-certificates \
                                                  libjemalloc1 \
                                                  ruby \
                                                  libgmp-dev \
    && echo 'gem: --no-document' >> /etc/gemrc \
    && gem install --file Gemfile \
    && git clone https://github.com/rewiko/fluentd.git fluentd  && cd fluentd \
    && git checkout merge-log-size-and-throttling \
    && gem install bundler && bundle && bundle exec rake build:all \
    && gem install --local ./pkg/fluentd-1.8.0.rc2.gem \
    && cd .. && rm -rf fluentd \
    && git clone https://github.com/rewiko/fluent-plugin-prometheus.git fluent-plugin-prometheus \
    && cd fluent-plugin-prometheus && git checkout add-file-size-metric \
    && bundle install && bundle exec rake build && gem install --local ./pkg/fluent-plugin-prometheus-1.6.1.gem \
    && cd .. && rm -rf fluentd-plugin-prometheus \
    && apt-get purge -y --auto-remove \
                     -o APT::AutoRemove::RecommendsImportant=false \
                     $BUILD_DEPS \
    && apt-get clean -y \
    && rm -rf /var/cache/debconf/* \
              /var/lib/apt/lists/* \
              /var/log/* \
              /tmp/* \
              /var/tmp/* \
              /usr/lib/ruby/gems/*/cache/*.gem \
    && ulimit -n 65536


# Copy the Fluentd configuration file for logging Docker container logs.
COPY fluent.conf /etc/fluent/fluent.conf
COPY run.sh /run.sh

# Expose prometheus metrics.
EXPOSE 80

ENV LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libjemalloc.so.1

# Start Fluentd to pick up our config that watches Docker container logs.
CMD ["/run.sh"]
