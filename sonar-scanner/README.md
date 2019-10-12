# Sonarqube
[оф. сайт](https://www.sonarqube.org/)

## Локальный запуск
### 1. Запустим сервер sonarqube
[hub.docker](https://hub.docker.com/_/sonarqube)
```
docker run --rm --name sonarqube -p 9000:9000 sonarqube
```

### 2. Запустим сканер
[Описание](https://docs.sonarqube.org/display/PLUG/SonarGo)
```
docker build -t sonar-scanner .
docker run --rm -it --network="host" -v ~/go/src/gitlab.ozon.ru/travel/customer-api/:/app sonar-scanner sonar-scanner -Dsonar.projectKey=travel:customer-api -Dsonar.projectName="customer-api" -Dsonar.sources=. -Dsonar.exclusions=**/*_test.go,**/vendor/**,**/*.pb.go,**/*.pb.goclay.go
```

Такое выполнение запустит сканер с дефолтной конфигурацией (<install_directory>/conf/sonar-scanner.properties):
```
sonar.host.url=http://localhost:9000
```

[Подробнее](https://docs.sonarqube.org/display/SCAN/Analyzing+with+SonarQube+Scanner)