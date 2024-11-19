FROM maven:3.9.9-amazoncorretto-23-alpine AS builder

WORKDIR /application
COPY pom.xml pom.xml
COPY src src

RUN mvn clean package -DskipTests

FROM amazoncorretto:23-alpine
WORKDIR /application
COPY --from=builder /application/target/*.jar app.jar
ENTRYPOINT ["java", "-jar", "app.jar"]

#FROM amazoncorretto:23-alpine
#
#RUN addgroup -S spring && adduser -S spring -G spring
#USER spring:spring
#
#COPY mvnw .
#COPY .mvn .mvn
#COPY pom.xml .
#COPY src src
#
#RUN chmod +x ./mvnw
#RUN ./mvnw dependency:resolve && ./mvnw install
#RUN cp ./target/*.jar app.jar
#
#
#ENTRYPOINT ["java", "-jar", "app.jar"]

#FROM ubuntu:latest
#LABEL authors="task-composer"
#
#ENTRYPOINT ["top", "-b"]