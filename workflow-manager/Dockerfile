FROM maven:3.9.9-amazoncorretto-23 AS builder

WORKDIR /application

COPY pom.xml .
RUN mvn dependency:go-offline -B

COPY src ./src
RUN mvn clean package -DskipTests

FROM amazoncorretto:23-alpine
WORKDIR /application

COPY --from=builder /application/target/*.jar app.jar

ENTRYPOINT ["java", "-jar", "app.jar"]