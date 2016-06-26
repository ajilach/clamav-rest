FROM scratch 

ADD target/clamrest / 

CMD ["/clamrest"]

EXPOSE 8080
