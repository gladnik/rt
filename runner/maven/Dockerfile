FROM maven:3.3.9-jdk-8-alpine

COPY pom.xml.tmpl /
COPY runner /

ENTRYPOINT ["/runner"]
