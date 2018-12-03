FROM openjdk:8-slim as builder

WORKDIR /app

COPY ["build.gradle", "gradlew", "./"]
COPY gradle gradle
RUN ./gradlew downloadRepos

COPY . .
RUN ./gradlew installDist 

FROM openjdk:8-alpine

RUN apk add --no-cache libc6-compat curl

WORKDIR /app
COPY --from=builder /app .

EXPOSE 10101 10102 10103
ENTRYPOINT ["/app/build/install/javaservice/bin/javaservice"]
