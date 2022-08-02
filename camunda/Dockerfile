FROM --platform=linux/arm64 camunda/camunda-bpm-platform:latest

ARG REDIS_HOST
# Session Manager part
RUN wget https://repo.maven.apache.org/maven2/redis/clients/jedis/2.9.0/jedis-2.9.0.jar -P lib/ && \
    wget https://repo1.maven.org/maven2/de/javakaffee/msm/memcached-session-manager/2.3.2/memcached-session-manager-2.3.2.jar -P lib/ && \
    wget https://repo1.maven.org/maven2/de/javakaffee/msm/memcached-session-manager-tc9/2.3.2/memcached-session-manager-tc9-2.3.2.jar -P lib/

## Add Manager to context
RUN sed -i '/^<\/Context>/i \
    <Manager className="de.javakaffee.web.msm.MemcachedBackupSessionManager" \
    memcachedNodes="redis://'${REDIS_HOST}'" \
    sticky="true" \
    sessionBackupAsync="false" \
    storageKeyPrefix="context" \
    lockingMode="all" \
    operationTimeout="3000" \
    requestUriIgnorePattern=".*\.(ico|png|gif|jpg|css|js)$" \
    />' conf/context.xml
