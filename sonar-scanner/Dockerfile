FROM debian:latest

RUN apt-get update && apt-get install -y wget unzip \
	&& wget https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-3.3.0.1492-linux.zip -O /tmp/scanner.zip \
	&& unzip /tmp/scanner.zip -d /scanner \
	&& ln -s /scanner/sonar-scanner-3.3.0.1492-linux/bin/sonar-scanner /usr/local/bin/sonar-scanner \
	&& mkdir /app
WORKDIR /app
CMD sonar-scanner 
